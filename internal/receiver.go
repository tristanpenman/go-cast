package internal

import "encoding/json"

const receiverNamespace = "urn:x-cast:com.google.cast.receiver"

func (clientConnection *ClientConnection) handleGetAppAvailability() {

}

func (clientConnection *ClientConnection) handleGetStatus() {

}

func (clientConnection *ClientConnection) handleLaunch() {

}

func (clientConnection *ClientConnection) handleStop() {

}

func (clientConnection *ClientConnection) handleReceiverMessage(data string) {
	var parsed map[string]*string
	err := json.Unmarshal([]byte(data), &parsed)
	if err != nil {
		clientConnection.log.Error("failed to parse receiver message", "err", err)
		return
	}

	dType := parsed["type"]
	if dType == nil {
		clientConnection.log.Error("receiver message type missing")
	}

	switch *dType {
	case "GET_APP_AVAILABILITY":
		clientConnection.handleGetAppAvailability()
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
		clientConnection.log.Error("unknown receiver message type", "type", dType)
		break
	}
}
