package internal

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/hashicorp/go-hclog"
	"github.com/tristanpenman/go-cast/internal/cast"
)

type Receiver struct {
	clientId int
	device   *Device
	id       string
	log      hclog.Logger
}

type ReceiverMessage struct {
	RequestId int    `json:"requestId"`
	Type      string `json:"type"`
}

type getAppAvailabilityRequest struct {
	*ReceiverMessage

	AppId []string `json:"appId"`
}

type getAppAvailabilityResponse struct {
	*ReceiverMessage

	Availability map[string]string `json:"availability"`
}

func (receiver *Receiver) handleGetAppAvailability(data string) {
	var request getAppAvailabilityRequest
	err := json.Unmarshal([]byte(data), &request)
	if err != nil {
		receiver.log.Error("failed to connect data", "err", err)
		return
	}

	availableApps := receiver.device.AvailableApps
	availability := make(map[string]string)
	for _, appId := range request.AppId {
		if sort.SearchStrings(availableApps, appId) < len(availableApps) {
			availability[appId] = "APP_AVAILABLE"
		} else {
			availability[appId] = "APP_UNAVAILABLE"
		}
	}

	response := getAppAvailabilityResponse{
		Availability: availability,
		ReceiverMessage: &ReceiverMessage{
			RequestId: request.RequestId,
			Type:      request.Type,
		},
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		receiver.log.Error("failed to marshall response for GET_APP_AVAILABILITY message")
		return
	}

	// broadcast app availability to all subscribers
	payloadUtf8 := string(bytes)
	receiver.device.broadcastUtf8(receiverNamespace, &payloadUtf8, receiver.id)
}

type volume struct {
	Level float32 `json:"level"`
	Muted bool    `json:"muted"`
}

type Namespace struct {
	Name string `json:"name"`
}

type Application struct {
	AppId        string      `json:"appId"`
	DisplayName  string      `json:"displayName"`
	IsIdleScreen bool        `json:"isIdleScreen"`
	Namespaces   []Namespace `json:"namespaces"`
	SessionId    string      `json:"sessionId"`
	StatusText   string      `json:"statusText"`
	TransportId  string      `json:"transportId"`
}

type Status struct {
	Applications  []Application `json:"applications"`
	IsActiveInput bool          `json:"isActiveInput"`
	Volume        volume        `json:"volume"`
}

type getStatusResponse struct {
	*ReceiverMessage

	Status Status `json:"status"`
}

func marshallNamespaces(namespaces []string) []Namespace {
	marshalled := make([]Namespace, len(namespaces))
	for index, namespace := range namespaces {
		marshalled[index] = Namespace{
			Name: namespace,
		}
	}

	return marshalled
}

func marshallApplicationStatuses(sessions map[string]*Session) []Application {
	marshalled := make([]Application, len(sessions))
	var index = 0
	for _, session := range sessions {
		marshalled[index] = Application{
			AppId:        session.AppId,
			DisplayName:  session.DisplayName,
			IsIdleScreen: false,
			Namespaces:   marshallNamespaces(session.Namespaces()),
			SessionId:    session.SessionId,
			StatusText:   session.StatusText,
			TransportId:  session.TransportId(),
		}
		index++
	}

	return marshalled
}

func (receiver *Receiver) handleGetStatus(requestId int) {
	response := getStatusResponse{
		ReceiverMessage: &ReceiverMessage{
			RequestId: requestId,
			Type:      "RECEIVER_STATUS",
		},
		Status: Status{
			Applications:  marshallApplicationStatuses(receiver.device.Sessions),
			IsActiveInput: true,
			Volume: volume{
				Level: 1.0,
				Muted: false,
			},
		},
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		receiver.log.Error("failed to marshall RECEIVER_STATUS message")
		return
	}

	payloadUtf8 := string(bytes)
	receiver.device.broadcastUtf8(receiverNamespace, &payloadUtf8, receiver.id)
}

type launchRequest struct {
	*ReceiverMessage

	AppId string `json:"appId"`
}

func (receiver *Receiver) handleLaunch(data string) {
	var request launchRequest
	var err = json.Unmarshal([]byte(data), &request)
	if err != nil {
		receiver.log.Error("failed to unmarshall launch request", "err", err)
		return
	}

	err = receiver.device.startApplication(request.AppId, receiver.clientId)
	if err != nil {
		receiver.log.Error("failed to start application", "err", err)
	}

	receiver.handleGetStatus(request.RequestId)
}

type stopRequest struct {
	*ReceiverMessage

	SessionId string `json:"sessionId"`
}

func (receiver *Receiver) handleStop(data string) {
	var request stopRequest
	err := json.Unmarshal([]byte(data), &request)
	if err != nil {
		receiver.log.Error("failed to unmarshall stop request", "err", err)
		return
	}

	receiver.device.stopApplication(request.SessionId)

	receiver.handleGetStatus(request.RequestId)
}

func (receiver *Receiver) handleReceiverMessage(castMessage *cast.CastMessage) {
	var parsed ReceiverMessage
	err := json.Unmarshal([]byte(*castMessage.PayloadUtf8), &parsed)
	if err != nil {
		receiver.log.Error("failed to parse receiver message", "err", err)
		return
	}

	switch parsed.Type {
	case "GET_APP_AVAILABILITY":
		receiver.handleGetAppAvailability(*castMessage.PayloadUtf8)
		break
	case "GET_STATUS":
		receiver.handleGetStatus(parsed.RequestId)
		break
	case "LAUNCH":
		receiver.handleLaunch(*castMessage.PayloadUtf8)
		break
	case "STOP":
		receiver.handleStop(*castMessage.PayloadUtf8)
		break
	default:
		receiver.log.Error("unknown receiver message type", "type", parsed.Type)
		break
	}
}

type DeviceInfoResponse struct {
	*ReceiverMessage

	ControlNotifications int    `json:"controlNotifications"`
	DeviceCapabilities   int    `json:"deviceCapabilities"`
	DeviceIconUrl        string `json:"deviceIconUrl"`
	DeviceId             string `json:"deviceId"`
	DeviceModel          string `json:"deviceModel"`
	FriendlyName         string `json:"friendlyName"`
	ReceiverMetricsId    string `json:"receiverMetricsId"`
	WifiProximityId      string `json:"wifiProximityId"`
}

func (receiver *Receiver) handleDiscoveryMessage(castMessage *cast.CastMessage) {
	var message ReceiverMessage
	err := json.Unmarshal([]byte(*castMessage.PayloadUtf8), &message)
	if err != nil {
		receiver.log.Error("failed to unmarshall discovery request")
		return
	}

	if message.Type != "GET_DEVICE_INFO" {
		receiver.log.Error("received unexpected discovery message type", "type", message.Type)
		return
	}

	response := DeviceInfoResponse{
		ReceiverMessage: &ReceiverMessage{
			RequestId: message.RequestId,
			Type:      "DEVICE_INFO",
		},
		ControlNotifications: 1,
		DeviceCapabilities:   4101,
		DeviceIconUrl:        "/setup/icon.png",
		DeviceId:             receiver.device.Id,
		DeviceModel:          receiver.device.DeviceModel,
		FriendlyName:         receiver.device.FriendlyName,
		ReceiverMetricsId:    "",
		WifiProximityId:      "",
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		receiver.log.Error("failed to marshall discovery response")
		return
	}

	payloadUtf8 := string(bytes)
	receiver.device.sendUtf8(discoveryNamespace, &payloadUtf8, *castMessage.DestinationId, *castMessage.SourceId)
}

type heartbeatMessage struct {
	Type string `json:"type"`
}

func (receiver *Receiver) handleHeartbeatMessage(castMessage *cast.CastMessage) {
	var message heartbeatMessage
	err := json.Unmarshal([]byte(*castMessage.PayloadUtf8), &message)
	if err != nil {
		receiver.log.Error("failed to unmarshall heartbeat request", "err", err)
		return
	}

	if message.Type != "PING" {
		receiver.log.Error("received unexpected heartbeat message type", "type", message.Type)
		return
	}

	// turn the message into a pong message
	message.Type = "PONG"
	bytes, err := json.Marshal(message)
	if err != nil {
		receiver.log.Error("failed to marshall heartbeat response")
		return
	}

	payloadUtf8 := string(bytes)
	receiver.device.sendUtf8(heartbeatNamespace, &payloadUtf8, *castMessage.DestinationId, *castMessage.SourceId)
}

type setupMessage struct {
	RequestId int    `json:"request_id"`
	Type      string `json:"type"`
}

type setupResponseDeviceInfo struct {
	SsdpUdn string `json:"ssdp_udn"`
}

type setupResponseData struct {
	DeviceInfo setupResponseDeviceInfo `json:"device_info"`
	Name       string                  `json:"name"`
	Version    int                     `json:"version"`
}

type setupResponse struct {
	*setupMessage

	Data setupResponseData `json:"data"`

	ResponseCode   int    `json:"response_code"`
	ResponseString string `json:"response_string"`
}

func (receiver *Receiver) handleSetupMessage(castMessage *cast.CastMessage) {
	var message setupMessage
	err := json.Unmarshal([]byte(*castMessage.PayloadUtf8), &message)
	if err != nil {
		receiver.log.Error("failed to parse setup message", "err", err)
		return
	}

	if message.Type != "eureka_info" {
		receiver.log.Error("received unexpected setup message type", "type", message.Type)
		return
	}

	response := setupResponse{
		setupMessage: &setupMessage{
			RequestId: message.RequestId,
			Type:      "eureka_info",
		},
		Data: setupResponseData{
			DeviceInfo: setupResponseDeviceInfo{
				SsdpUdn: receiver.device.Udn,
			},
			Name:    receiver.device.FriendlyName,
			Version: 8,
		},
		ResponseCode:   200,
		ResponseString: "OK",
	}

	// turn the message into a pong message
	bytes, err := json.Marshal(response)
	if err != nil {
		receiver.log.Error("failed to marshall setup response")
		return
	}

	payloadUtf8 := string(bytes)
	receiver.device.sendUtf8(setupNamespace, &payloadUtf8, *castMessage.DestinationId, *castMessage.SourceId)
}

func (receiver *Receiver) HandleCastMessage(castMessage *cast.CastMessage) {
	switch *castMessage.Namespace {
	case heartbeatNamespace:
		receiver.handleHeartbeatMessage(castMessage)
		return
	case discoveryNamespace:
		receiver.handleDiscoveryMessage(castMessage)
		return
	case receiverNamespace:
		receiver.handleReceiverMessage(castMessage)
		return
	case setupNamespace:
		receiver.handleSetupMessage(castMessage)
		return
	default:
		receiver.log.Info("received message for unknown namespace", "namespace", *castMessage.Namespace)
	}
}

func (receiver *Receiver) TransportId() string {
	return receiver.id
}

func NewReceiver(device *Device, id string, clientId int) *Receiver {
	log := NewLogger(fmt.Sprintf("receiver (%d) [%s]", clientId, id))

	return &Receiver{
		clientId: clientId,
		device:   device,
		id:       id,
		log:      log,
	}
}
