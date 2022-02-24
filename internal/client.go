package internal

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
		Logger.Error("failed to encode device auth challenge", "err", err)
		return false
	}

	payloadType := cast.CastMessage_BINARY
	return castChannel.Send(&cast.CastMessage{
		Namespace:     &cast.DeviceAuthNamespace,
		PayloadBinary: payloadBytes,
		PayloadType:   &payloadType,
	})
}

func StartClient(hostname *string, port *uint, authChallenge bool) {
	addr := fmt.Sprintf("%s:%d", *hostname, *port)
	Logger.Info(fmt.Sprintf("addr: %s", addr))

	config := tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		Logger.Error("client: dial error", "err", err)
		return
	}

	defer func() {
		_ = conn.Close()
		Logger.Info("connection closed")
	}()

	castChannel := cast.CreateCastChannel(conn, Logger)

	if authChallenge {
		sendDeviceAuthChallenge(&castChannel)
	}

	for {
		select {
		case castMessage, ok := <-castChannel.Messages:
			if castMessage != nil {
				Logger.Info("received", "message", castMessage)
			}
			if !ok {
				Logger.Info("channel closed")
				return
			}
		}
	}
}
