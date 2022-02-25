package internal

import "encoding/json"

const heartbeatNamespace = "urn:x-cast:com.google.cast.tp.heartbeat"

func (clientConnection *ClientConnection) handleHeartbeatMessage(data string) {
	var parsed map[string]string
	err := json.Unmarshal([]byte(data), &parsed)
	if err != nil {
		clientConnection.log.Error("failed to parse certificate manifest file", "err", err)
		return
	}

	dType := parsed["type"]
	if dType == "PING" {
		clientConnection.log.Info("PING")
	}
}
