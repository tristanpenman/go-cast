package main

import (
	"crypto/tls"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/tristanpenman/go-cast/internal/cast"
)

func sendDeviceAuthChallenge(castChannel *cast.CastChannel) bool {
	payloadBytes, err := proto.Marshal(&cast.DeviceAuthMessage{
		Challenge: &cast.AuthChallenge{},
	})

	if err != nil {
		logger.Error("failed to encode device auth challenge", "err", err)
		return false
	}

	payloadType := cast.CastMessage_BINARY
	return castChannel.Send(&cast.CastMessage{
		Namespace: &cast.DeviceAuthNamespace,
		PayloadBinary: payloadBytes,
		PayloadType: &payloadType,
	})
}

func startClient(hostname *string, port *uint) {
	addr := fmt.Sprintf("%s:%d", *hostname, *port)
	logger.Info(fmt.Sprintf("addr: %s", addr))

	config := tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		logger.Error("client: dial error", "err", err)
		return
	}

	defer func() {
		_ = conn.Close()
		logger.Info("connection closed")
	}()

	castChannel := cast.CreateCastChannel(conn, logger)

	sendDeviceAuthChallenge(&castChannel)

	for {
		select {
		case castMessage, ok := <-castChannel.Messages:
			if castMessage != nil {
				logger.Info("received", "message", castMessage)
			}
			if !ok {
				logger.Info("channel closed")
				return
			}
		}
	}
}
