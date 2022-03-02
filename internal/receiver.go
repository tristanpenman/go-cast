package internal

import (
	"encoding/json"
)

const receiverNamespace = "urn:x-cast:com.google.cast.receiver"

type receiverMessage struct {
	Type string `json:"type"`
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
}

type getStatusResponse struct {
}

func (clientConnection *ClientConnection) handleGetStatus() {

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
	var parsed receiverMessage
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
		clientConnection.handleGetStatus()
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
