package sender

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	// third-party
	"github.com/hashicorp/go-hclog"

	// internal
	"github.com/tristanpenman/go-cast/pkg"
	"github.com/tristanpenman/go-cast/pkg/channel"
)

// Default source and destination IDs used for the receiver control channel.
const (
	DefaultSenderID   = "sender-0"
	DefaultReceiverID = "receiver-0"
)

const (
	receiverNamespace   = "urn:x-cast:com.google.cast.receiver"
	connectionNamespace = "urn:x-cast:com.google.cast.tp.connection"
	heartbeatNamespace  = "urn:x-cast:com.google.cast.tp.heartbeat"
)

// Application describes a running receiver application, as reported by a
// RECEIVER_STATUS message.
type Application struct {
	AppID       string `json:"appId"`
	DisplayName string `json:"displayName"`
	StatusText  string `json:"statusText"`
	TransportID string `json:"transportId"`
}

// ReceiverStatus is the most recent status reported by the receiver.
type ReceiverStatus struct {
	Applications []Application
}

// requestMessage is the common envelope shared by receiver control messages.
type requestMessage struct {
	RequestID int    `json:"requestId"`
	Type      string `json:"type"`
}

type launchRequest struct {
	requestMessage
	AppID string `json:"appId"`
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

// Sender wraps an internal.Client and implements the Cast sender protocol:
// connecting and authenticating to a receiver, sending CONNECT, GET_STATUS,
// LAUNCH and app namespace messages, and tracking the receiver's reported
// status, running sessions, transport IDs and errors.
type Sender struct {
	client *internal.Client
	log    hclog.Logger

	senderID   string
	receiverID string

	mu        sync.Mutex
	cond      *sync.Cond
	requestID int
	status    *ReceiverStatus
	err       error
	closed    bool
}

// New creates a Sender that drives the given client and starts consuming
// incoming messages to track receiver state.
func New(client *internal.Client, log hclog.Logger) *Sender {
	if log == nil {
		log = internal.NewLogger("sender")
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
	s.client.SendMessage(newUTF8CastMessage(connectionNamespace, sourceID, destinationID, payload))
}

// RequestStatus sends a GET_STATUS message to the receiver.
func (s *Sender) RequestStatus() {
	request := requestMessage{RequestID: s.nextRequestID(), Type: "GET_STATUS"}
	payloadBytes, _ := json.Marshal(request)
	s.client.SendMessage(newUTF8CastMessage(receiverNamespace, s.senderID, s.receiverID, string(payloadBytes)))
}

// LaunchApp sends a LAUNCH message asking the receiver to start an app.
func (s *Sender) LaunchApp(appID string) {
	request := launchRequest{
		requestMessage: requestMessage{RequestID: s.nextRequestID(), Type: "LAUNCH"},
		AppID:          appID,
	}
	payloadBytes, _ := json.Marshal(request)
	s.client.SendMessage(newUTF8CastMessage(receiverNamespace, s.senderID, s.receiverID, string(payloadBytes)))
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
		if app.AppID == appID && app.TransportID != "" {
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
	timer := time.AfterFunc(timeout, func() {
		s.mu.Lock()
		s.cond.Broadcast()
		s.mu.Unlock()
	})
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

func (s *Sender) readLoop() {
	for castMessage := range s.client.Incoming {
		if castMessage == nil || castMessage.Namespace == nil {
			continue
		}

		switch *castMessage.Namespace {
		case heartbeatNamespace:
			if castMessage.PayloadUtf8 != nil && *castMessage.PayloadUtf8 == `{"type":"PING"}` {
				s.SendAppMessage(heartbeatNamespace, s.receiverID, `{"type":"PONG"}`)
			}
		case receiverNamespace:
			s.handleReceiverMessage(castMessage)
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

	switch envelope.Type {
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
