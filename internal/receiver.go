package internal

import (
	"encoding/json"
	"sort"
)

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

func (clientConnection *ClientConnection) handleGetAppAvailability(data string) {
	var request getAppAvailabilityRequest
	err := json.Unmarshal([]byte(data), &request)
	if err != nil {
		clientConnection.log.Error("failed to connect data", "err", err)
		return
	}

	availability := make(map[string]string)
	for _, appId := range request.AppId {
		if sort.SearchStrings(clientConnection.availableApps, appId) < len(clientConnection.availableApps) {
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
		clientConnection.log.Error("failed to marshall response for GET_APP_AVAILABILITY message")
		return
	}

	clientConnection.sendUtf8Message(bytes, receiverNamespace)
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

func flattenApplications(applications map[string]*Application) []Application {
	flattened := make([]Application, len(applications))
	var index = 0
	for _, application := range applications {
		// TODO: convert from a more natural internal state to the Application interface
		flattened[index] = *application
		index++
	}

	return flattened
}

func (clientConnection *ClientConnection) handleGetStatus(requestId int) {
	response := getStatusResponse{
		ReceiverMessage: &ReceiverMessage{
			RequestId: requestId,
			Type:      "GET_STATUS",
		},
		Status: status{
			Applications:  flattenApplications(clientConnection.applications),
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
	*ReceiverMessage

	AppId string `json:"appId"`
}

func (clientConnection *ClientConnection) handleLaunch(data string) {
	var request launchRequest
	err := json.Unmarshal([]byte(data), &request)
	if err != nil {
		clientConnection.log.Error("failed to unmarshall launch request", "err", err)
		return
	}

	if clientConnection.startApplication(request.AppId) {
		clientConnection.handleGetStatus(request.RequestId)
	} else {
		// TODO: How to handle application not being started?
	}
}

type stopRequest struct {
	*ReceiverMessage

	SessionId string `json:"sessionId"`
}

func (clientConnection *ClientConnection) handleStop(data string) {
	var request stopRequest
	err := json.Unmarshal([]byte(data), &request)
	if err != nil {
		clientConnection.log.Error("failed to unmarshall stop request", "err", err)
		return
	}

	clientConnection.handleGetStatus(request.RequestId)
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
		clientConnection.handleLaunch(data)
		break
	case "STOP":
		clientConnection.handleStop(data)
		break
	default:
		clientConnection.log.Error("unknown receiver message type", "type", parsed.Type)
		break
	}
}
