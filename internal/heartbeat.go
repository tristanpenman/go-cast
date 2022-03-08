package internal

import (
	"encoding/json"
)

type heartbeatMessage struct {
	Type string `json:"type"`
}

func (clientConnection *ClientConnection) handleHeartbeatMessage(data string) {
	var message heartbeatMessage
	err := json.Unmarshal([]byte(data), &message)
	if err != nil {
		clientConnection.log.Error("failed to parse certificate manifest file", "err", err)
		return
	}

	if message.Type != "PING" {
		clientConnection.log.Error("received unexpected heartbeat message type", "type", message.Type)
		return
	}

	// turn the message into a pong message
	message.Type = "PONG"
	bytes, err := json.Marshal(message)
	if err != nil {
		clientConnection.log.Error("failed to marshall heartbeat response")
		return
	}

	payloadUtf8 := string(bytes)
	clientConnection.sendUtf8(heartbeatNamespace, &payloadUtf8)
}
