package internal

import (
	"encoding/json"
)

const receiverNamespace = "urn:x-cast:com.google.cast.receiver"

type ReceiverMessage struct {
	RequestId int    `json:"requestId"`
	Type      string `json:"type"`
}

type getAppAvailabilityRequest struct {
	AppId []string `json:"appId"`
}

type getAppAvailabilityResponse struct {
}

func (clientConnection *ClientConnection) handleGetAppAvailability(data string) {
	var request getAppAvailabilityRequest
	err := json.Unmarshal([]byte(data), &request)
	if err != nil {
		clientConnection.log.Error("failed to connect data", "err", err)
		return
	}
}

type getStatusRequest struct {
	*ReceiverMessage
}

type volume struct {
	Level int  `json:"level"`
	Muted bool `json:"muted"`
}

type status struct {
	Applications  []Application `json:"applications"`
	IsActiveInput bool          `json:"isActiveInput"`
	Volume        volume        `json:"volume"`
}

type getStatusResponse struct {
	*ReceiverMessage

	Status status `json:"status"`
}

func (clientConnection *ClientConnection) handleGetStatus(requestId int) {
	response := getStatusResponse{
		ReceiverMessage: &ReceiverMessage{
			RequestId: requestId,
			Type:      "GET_STATUS",
		},
		Status: status{
			Applications:  clientConnection.applications,
			IsActiveInput: true,
			Volume: volume{
				Level: 1,
				Muted: false,
			},
		},
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		clientConnection.log.Error("failed to marshall response for GET_STATUS message")
		return
	}

	clientConnection.sendUtf8Message(bytes, receiverNamespace)
}

type launchRequest struct {
}

type launchResponse struct {
}

func (clientConnection *ClientConnection) handleLaunch() {

}

type stopRequest struct {
}

type stopResponse struct {
}

func (clientConnection *ClientConnection) handleStop() {

}

func (clientConnection *ClientConnection) handleReceiverMessage(data string) {
	var parsed ReceiverMessage
	err := json.Unmarshal([]byte(data), &parsed)
	if err != nil {
		clientConnection.log.Error("failed to parse receiver message", "err", err)
		return
	}

	switch parsed.Type {
	case "GET_APP_AVAILABILITY":
		clientConnection.handleGetAppAvailability(data)
		break
	case "GET_STATUS":
		clientConnection.handleGetStatus(parsed.RequestId)
		break
	case "LAUNCH":
		clientConnection.handleLaunch()
		break
	case "STOP":
		clientConnection.handleStop()
		break
	default:
		clientConnection.log.Error("unknown receiver message type", "type", parsed.Type)
		break
	}
}
