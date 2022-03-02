package internal

import (
	"encoding/json"
)

const connectNamespace = "urn:x-cast:com.google.cast.tp.connection"

type connectMessage struct {
	ConnType json.Number `json:"connType"`
}

func (clientConnection *ClientConnection) handleConnectMessage(data string) {
	var parsed connectMessage
	err := json.Unmarshal([]byte(data), &parsed)
	if err != nil {
		clientConnection.log.Error("failed to connect data", "err", err)
		return
	}

	clientConnection.log.Info("connect", "data", data)
}
