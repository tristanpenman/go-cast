package internal

import (
	"encoding/json"
)

const connectNamespace = "urn:x-cast:com.google.cast.tp.connection"

type connectRequest struct {
	ConnType json.Number `json:"connType"`
}

func (clientConnection *ClientConnection) handleConnectMessage(data string) {
	var request connectRequest
	err := json.Unmarshal([]byte(data), &request)
	if err != nil {
		clientConnection.log.Error("failed to connect data", "err", err)
		return
	}

	// TODO: store connection data
}
