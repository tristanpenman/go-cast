package internal

import (
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-hclog"
	"github.com/tristanpenman/go-cast/internal/cast"
	"net"
)

type ClientConnection struct {
	conn        net.Conn
	castChannel cast.CastChannel
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
}

func NewClientConnection(conn net.Conn, relayClient *Client) *ClientConnection {
	var log = NewLogger("client-connection")

	defer func() {
		_ = conn.Close()
		log.Info("connection closed")
	}()

	castChannel := cast.CreateCastChannel(conn, log)

	clientConnection := ClientConnection{
		conn:        conn,
		castChannel: castChannel,
		log:         log,
		relayClient: relayClient,
	}

	go func() {
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
