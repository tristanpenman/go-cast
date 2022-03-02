package internal

import (
	"encoding/json"
)

const heartbeatNamespace = "urn:x-cast:com.google.cast.tp.heartbeat"

type heartbeatMessage struct {
	Type string `json:"type"`
}

func (clientConnection *ClientConnection) handleHeartbeatMessage(data string) []byte {
	var message heartbeatMessage
	err := json.Unmarshal([]byte(data), &message)
	if err != nil {
		clientConnection.log.Error("failed to parse certificate manifest file", "err", err)
		return nil
	}

	if message.Type != "PING" {
		clientConnection.log.Error("received unexpected heartbeat message type", "type", message.Type)
		return nil
	}

	// turn the message into a pong message
	message.Type = "PONG"
	bytes, err := json.Marshal(message)
	if err != nil {
		clientConnection.log.Error("failed to marshall heartbeat response")
		return nil
	}

	return bytes
}
