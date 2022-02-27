package internal

import (
	"net"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/cast"
)

type ClientConnection struct {
	castChannel cast.CastChannel
	conn        net.Conn
	log         hclog.Logger
	relayClient *Client
}

func (clientConnection *ClientConnection) sendDeviceAuthResponse() bool {
	payloadBytes, err := proto.Marshal(&cast.DeviceAuthMessage{
		Response: &cast.AuthResponse{},
	})

	if err != nil {
		clientConnection.log.Error("failed to encode device auth response", "err", err)
		return false
	}

	payloadType := cast.CastMessage_BINARY
	message := cast.CastMessage{
		Namespace:     &cast.DeviceAuthNamespace,
		PayloadBinary: payloadBytes,
		PayloadType:   &payloadType,
	}

	return clientConnection.castChannel.Send(&message)
}

func (clientConnection *ClientConnection) relayCastMessage(castMessage *cast.CastMessage) {
	clientConnection.log.Info("relay cast message")
}

func (clientConnection *ClientConnection) handleCastMessage(castMessage *cast.CastMessage) {
	clientConnection.log.Info("handle cast message")
	switch *castMessage.Namespace {
	case heartbeatNamespace:
		clientConnection.handleHeartbeatMessage(*castMessage.PayloadUtf8)
		break
	default:
		clientConnection.log.Info("unhandled message", "namespace", *castMessage.Namespace)
	}
}

func NewClientConnection(conn net.Conn, relayClient *Client) *ClientConnection {
	var log = NewLogger("client-connection")

	castChannel := cast.CreateCastChannel(conn, log)

	clientConnection := ClientConnection{
		castChannel: castChannel,
		conn:        conn,
		log:         log,
		relayClient: relayClient,
	}

	go func() {
		defer func() {
			_ = conn.Close()
			log.Info("connection closed")
		}()

		for {
			select {
			case castMessage, ok := <-castChannel.Messages:
				if castMessage != nil {
					log.Info("received", "message", castMessage)
					if *castMessage.Namespace == cast.DeviceAuthNamespace {
						clientConnection.sendDeviceAuthResponse()
					} else if relayClient != nil {
						clientConnection.relayCastMessage(castMessage)
					} else {
						clientConnection.handleCastMessage(castMessage)
					}
				}
				if !ok {
					log.Info("channel closed")
					return
				}
			}
		}
	}()

	return &clientConnection
}
