package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	// third-party
	"github.com/hashicorp/go-hclog"

	// internal
	"github.com/tristanpenman/go-cast/internal/channel"
	"github.com/tristanpenman/go-cast/internal/common"
)

// Default source and destination IDs used for the receiver control channel.
const (
	DefaultSenderID   = "sender-0"
	DefaultReceiverID = "receiver-0"
)

// Application describes a running receiver application, as reported by a
// RECEIVER_STATUS message.
type Application struct {
	AppID          string `json:"appId"`
	AppType        string `json:"appType"`
	DisplayName    string `json:"displayName"`
	IsIdleScreen   bool   `json:"isIdleScreen"`
	SessionID      string `json:"sessionId"`
	StatusText     string `json:"statusText"`
	TransportID    string `json:"transportId"`
	UniversalAppID string `json:"universalAppId"`
}

func (a Application) MatchesAppID(appID string) bool {
	return a.AppID == appID || a.UniversalAppID == appID
}

// ReceiverStatus is the most recent status reported by the receiver.
type ReceiverStatus struct {
	Applications []Application
}

// requestMessage is the common envelope shared by receiver control messages.
type requestMessage struct {
	RequestID    int    `json:"requestId"`
	Type         string `json:"type,omitempty"`
	ResponseType string `json:"responseType,omitempty"`
}

func (m requestMessage) messageType() string {
	if m.Type != "" {
		return m.Type
	}
	return m.ResponseType
}

type launchRequest struct {
	requestMessage
	AppID string `json:"appId"`
}

type stopRequest struct {
	requestMessage
	SessionID string `json:"sessionId"`
}

type appAvailabilityRequest struct {
	requestMessage
	AppIDs []string `json:"appId"`
}

type appAvailabilityMessage struct {
	requestMessage
	Availability map[string]string `json:"availability"`
}

type statusPayload struct {
	Applications []Application `json:"applications"`
}

type receiverStatusMessage struct {
	requestMessage
	Status statusPayload `json:"status"`
}

type errorMessage struct {
	requestMessage
	Reason string `json:"reason"`
}

// Sender wraps a Client and implements the Cast sender protocol:
// connecting and authenticating to a receiver, sending CONNECT, GET_STATUS,
// LAUNCH and app namespace messages, and tracking the receiver's reported
// status, running sessions, transport IDs and errors.
type Sender struct {
	client *Client
	log    hclog.Logger

	senderID   string
	receiverID string

	mu              sync.Mutex
	cond            *sync.Cond
	requestID       int
	status          *ReceiverStatus
	availability    map[string]string
	youtubeScreenID string
	err             error
	closed          bool
}

// NewSender creates a Sender that drives the given client and starts consuming
// incoming messages to track receiver state.
func NewSender(client *Client, log hclog.Logger) *Sender {
	if log == nil {
		log = common.NewLogger("sender")
	}

	s := &Sender{
		client:     client,
		log:        log,
		senderID:   DefaultSenderID,
		receiverID: DefaultReceiverID,
	}
	s.cond = sync.NewCond(&s.mu)

	go s.readLoop()

	return s
}

func newUTF8CastMessage(namespace, sourceID, destinationID, payload string) *channel.CastMessage {
	payloadType := channel.CastMessage_STRING
	protocolVersion := channel.CastMessage_CASTV2_1_0

	return &channel.CastMessage{
		DestinationId:   &destinationID,
		Namespace:       &namespace,
		PayloadType:     &payloadType,
		PayloadUtf8:     &payload,
		ProtocolVersion: &protocolVersion,
		SourceId:        &sourceID,
	}
}

func (s *Sender) nextRequestID() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requestID++
	return s.requestID
}

// Connect sends a CONNECT message to the receiver's control channel.
func (s *Sender) Connect() {
	s.sendConnection(s.senderID, s.receiverID)
}

// ConnectTransport sends a CONNECT message to a specific transport (an app
// session), which must be done before exchanging app namespace messages.
func (s *Sender) ConnectTransport(transportID string) {
	s.sendConnection(s.senderID, transportID)
}

func (s *Sender) sendConnection(sourceID, destinationID string) {
	payload := `{"type":"CONNECT"}`
	s.client.SendMessage(newUTF8CastMessage(common.ConnectionNamespace, sourceID, destinationID, payload))
}

// RequestStatus sends a GET_STATUS message to the receiver.
func (s *Sender) RequestStatus() {
	request := requestMessage{RequestID: s.nextRequestID(), Type: "GET_STATUS"}
	payloadBytes, _ := json.Marshal(request)
	s.client.SendMessage(newUTF8CastMessage(common.ReceiverNamespace, s.senderID, s.receiverID, string(payloadBytes)))
}

// RequestAppAvailability asks whether the receiver can launch the supplied app IDs.
func (s *Sender) RequestAppAvailability(appIDs []string) {
	request := appAvailabilityRequest{
		requestMessage: requestMessage{RequestID: s.nextRequestID(), Type: "GET_APP_AVAILABILITY"},
		AppIDs:         appIDs,
	}
	payloadBytes, _ := json.Marshal(request)
	s.client.SendMessage(newUTF8CastMessage(common.ReceiverNamespace, s.senderID, s.receiverID, string(payloadBytes)))
}

// LaunchApp sends a LAUNCH message asking the receiver to start an app.
func (s *Sender) LaunchApp(appID string) {
	s.clearError()
	request := launchRequest{
		requestMessage: requestMessage{RequestID: s.nextRequestID(), Type: "LAUNCH"},
		AppID:          appID,
	}
	payloadBytes, _ := json.Marshal(request)
	s.client.SendMessage(newUTF8CastMessage(common.ReceiverNamespace, s.senderID, s.receiverID, string(payloadBytes)))
}

// StopApp asks the receiver to terminate an application session.
func (s *Sender) StopApp(sessionID string) {
	s.clearError()
	request := stopRequest{
		requestMessage: requestMessage{RequestID: s.nextRequestID(), Type: "STOP"},
		SessionID:      sessionID,
	}
	payloadBytes, _ := json.Marshal(request)
	s.client.SendMessage(newUTF8CastMessage(common.ReceiverNamespace, s.senderID, s.receiverID, string(payloadBytes)))
}

// SendAppMessage sends a UTF-8 payload on an app-specific namespace to a
// transport (session) destination.
func (s *Sender) SendAppMessage(namespace, transportID, payload string) {
	s.client.SendMessage(newUTF8CastMessage(namespace, s.senderID, transportID, payload))
}

// Status returns a copy of the most recently reported receiver status, or nil
// if no status has been received yet.
func (s *Sender) Status() *ReceiverStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statusLocked()
}

// Availability returns the most recently reported app availability map.
func (s *Sender) Availability() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.availability == nil {
		return nil
	}
	result := make(map[string]string, len(s.availability))
	for appID, value := range s.availability {
		result[appID] = value
	}
	return result
}

// WaitForStatus blocks until the receiver has reported its status.
func (s *Sender) WaitForStatus(timeout time.Duration) (*ReceiverStatus, error) {
	deadline := time.Now().Add(timeout)
	timer := s.broadcastAtTimeout(timeout)
	defer timer.Stop()

	s.mu.Lock()
	defer s.mu.Unlock()
	for s.status == nil {
		if err := s.waitErrorLocked(deadline, "receiver status"); err != nil {
			return nil, err
		}
		s.cond.Wait()
	}
	return s.statusLocked(), nil
}

// WaitForAvailability blocks until the receiver answers an app availability query.
func (s *Sender) WaitForAvailability(timeout time.Duration) (map[string]string, error) {
	deadline := time.Now().Add(timeout)
	timer := s.broadcastAtTimeout(timeout)
	defer timer.Stop()

	s.mu.Lock()
	defer s.mu.Unlock()
	for s.availability == nil {
		if err := s.waitErrorLocked(deadline, "app availability"); err != nil {
			return nil, err
		}
		s.cond.Wait()
	}
	result := make(map[string]string, len(s.availability))
	for appID, value := range s.availability {
		result[appID] = value
	}
	return result, nil
}

func (s *Sender) statusLocked() *ReceiverStatus {
	if s.status == nil {
		return nil
	}
	apps := make([]Application, len(s.status.Applications))
	copy(apps, s.status.Applications)
	return &ReceiverStatus{Applications: apps}
}

// TransportID returns the transport ID for a running app, or an empty string if
// the app is not currently reported as running.
func (s *Sender) TransportID(appID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.transportIDLocked(appID)
}

func (s *Sender) transportIDLocked(appID string) string {
	if s.status == nil {
		return ""
	}
	for _, app := range s.status.Applications {
		if app.MatchesAppID(appID) && !app.IsIdleScreen && app.TransportID != "" {
			return app.TransportID
		}
	}
	return ""
}

// SessionTransportID returns the transport ID for an app session even when the
// receiver marks that session as an idle screen. This is useful for apps such
// as YouTube, whose ready-to-cast session is still a valid message transport.
func (s *Sender) SessionTransportID(appID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessionTransportIDLocked(appID)
}

func (s *Sender) sessionTransportIDLocked(appID string) string {
	if s.status == nil {
		return ""
	}
	for _, app := range s.status.Applications {
		if app.MatchesAppID(appID) && app.TransportID != "" {
			return app.TransportID
		}
	}
	return ""
}

// Err returns the most recent error reported by the receiver, if any.
func (s *Sender) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// WaitForApp blocks until the receiver reports the given app as running and
// returns its transport ID, or fails on timeout, receiver error, or a closed
// connection.
func (s *Sender) WaitForApp(appID string, timeout time.Duration) (string, error) {
	timer := s.broadcastAtTimeout(timeout)
	defer timer.Stop()

	deadline := time.Now().Add(timeout)

	s.mu.Lock()
	defer s.mu.Unlock()
	for {
		if transportID := s.transportIDLocked(appID); transportID != "" {
			return transportID, nil
		}
		if s.err != nil {
			return "", s.err
		}
		if s.closed {
			return "", errors.New("connection closed")
		}
		if !time.Now().Before(deadline) {
			return "", fmt.Errorf("timed out waiting for %s session", appID)
		}
		s.cond.Wait()
	}
}

// WaitForAppTransport blocks until the receiver reports a message transport
// for the app. Unlike WaitForApp, it accepts an idle-marked app session.
func (s *Sender) WaitForAppTransport(appID string, timeout time.Duration) (string, error) {
	timer := s.broadcastAtTimeout(timeout)
	defer timer.Stop()

	deadline := time.Now().Add(timeout)

	s.mu.Lock()
	defer s.mu.Unlock()
	for {
		if transportID := s.sessionTransportIDLocked(appID); transportID != "" {
			return transportID, nil
		}
		if s.err != nil {
			return "", s.err
		}
		if s.closed {
			return "", errors.New("connection closed")
		}
		if !time.Now().Before(deadline) {
			return "", fmt.Errorf("timed out waiting for %s message transport (%s)", appID, s.appSessionStateLocked(appID))
		}
		s.cond.Wait()
	}
}

func (s *Sender) appSessionStateLocked(appID string) string {
	if s.status == nil {
		return "no receiver status received"
	}
	for _, app := range s.status.Applications {
		if app.MatchesAppID(appID) {
			return fmt.Sprintf("app reported with idle=%t, sessionId=%q, transportId=%q", app.IsIdleScreen, app.SessionID, app.TransportID)
		}
	}
	return "app not present in receiver status"
}

// WaitForAppStopped blocks until an app is no longer present in receiver status.
func (s *Sender) WaitForAppStopped(appID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	timer := s.broadcastAtTimeout(timeout)
	defer timer.Stop()

	s.mu.Lock()
	defer s.mu.Unlock()
	for s.appRunningLocked(appID) {
		if err := s.waitErrorLocked(deadline, appID+" to stop"); err != nil {
			return err
		}
		s.cond.Wait()
	}
	return nil
}

func (s *Sender) appRunningLocked(appID string) bool {
	if s.status == nil {
		return false
	}
	for _, app := range s.status.Applications {
		if app.MatchesAppID(appID) && !app.IsIdleScreen {
			return true
		}
	}
	return false
}

func (s *Sender) broadcastAtTimeout(timeout time.Duration) *time.Timer {
	return time.AfterFunc(timeout, func() {
		s.mu.Lock()
		s.cond.Broadcast()
		s.mu.Unlock()
	})
}

func (s *Sender) waitErrorLocked(deadline time.Time, waitingFor string) error {
	if s.err != nil {
		return s.err
	}
	if s.closed {
		return errors.New("connection closed")
	}
	if !time.Now().Before(deadline) {
		return fmt.Errorf("timed out waiting for %s", waitingFor)
	}
	return nil
}

func (s *Sender) readLoop() {
	for castMessage := range s.client.Incoming {
		if castMessage == nil || castMessage.Namespace == nil {
			continue
		}

		switch *castMessage.Namespace {
		case common.HeartbeatNamespace:
			if castMessage.PayloadUtf8 != nil && *castMessage.PayloadUtf8 == `{"type":"PING"}` {
				s.SendAppMessage(common.HeartbeatNamespace, s.receiverID, `{"type":"PONG"}`)
			}
		case common.ReceiverNamespace:
			s.handleReceiverMessage(castMessage)
		case youtubeNamespace:
			s.handleYouTubeMessage(castMessage)
		}
	}

	s.mu.Lock()
	s.closed = true
	s.cond.Broadcast()
	s.mu.Unlock()
}

func (s *Sender) handleReceiverMessage(castMessage *channel.CastMessage) {
	if castMessage.PayloadUtf8 == nil {
		return
	}

	payload := []byte(*castMessage.PayloadUtf8)

	var envelope requestMessage
	if err := json.Unmarshal(payload, &envelope); err != nil {
		s.log.Warn("failed to parse receiver payload", "err", err)
		return
	}

	switch envelope.messageType() {
	case "GET_APP_AVAILABILITY":
		var msg appAvailabilityMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			s.log.Warn("failed to parse app availability", "err", err)
			return
		}
		s.updateAvailability(msg.Availability)
	case "RECEIVER_STATUS":
		var msg receiverStatusMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			s.log.Warn("failed to parse receiver status", "err", err)
			return
		}
		s.updateStatus(msg.Status)
	case "LAUNCH_ERROR", "INVALID_REQUEST", "LOAD_FAILED":
		var msg errorMessage
		_ = json.Unmarshal(payload, &msg)
		s.setError(fmt.Errorf("receiver reported %s: %s", envelope.Type, msg.Reason))
	}
}

func (s *Sender) updateAvailability(availability map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.availability = availability
	s.cond.Broadcast()
}

func (s *Sender) updateStatus(payload statusPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = &ReceiverStatus{Applications: payload.Applications}
	s.cond.Broadcast()
}

func (s *Sender) setError(err error) {
	s.log.Warn("receiver error", "err", err)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
	s.cond.Broadcast()
}

func (s *Sender) clearError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = nil
}
