package internal

import (
	"encoding/json"
)

type SetupMessage struct {
	Type string `json:"type"`
}

type setupResponseDeviceInfo struct {
	SsdpUdn string `json:"ssdp_udn"`
}

type setupResponseData struct {
	DeviceInfo setupResponseDeviceInfo `json:"deviceInfo"`
	Name       string                  `json:"Name"`
	Version    int                     `json:"version"`
}

type setupResponse struct {
	*SetupMessage

	Data setupResponseData
}

func (clientConnection *ClientConnection) handleSetupMessage(data string) []byte {
	clientConnection.log.Info("setup", "data", data)

	var message SetupMessage
	err := json.Unmarshal([]byte(data), &message)
	if err != nil {
		clientConnection.log.Error("failed to parse setup message", "err", err)
		return nil
	}

	response := setupResponse{
		SetupMessage: &SetupMessage{
			Type: "eureka_info",
		},
		Data: setupResponseData{
			DeviceInfo: setupResponseDeviceInfo{
				SsdpUdn: "ce391871-f16d-4b9c-8bab-05e856297f0a",
			},
			Name:    clientConnection.device.FriendlyName,
			Version: 8,
		},
	}

	bytes, _ := json.Marshal(response)

	return bytes
}