package internal

import (
	"net"

	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/cast"
)

type ClientConnection struct {
	castChannel cast.CastChannel
	conn        net.Conn
	log         hclog.Logger
	relayClient *Client
}

func (clientConnection *ClientConnection) handleCastMessage(castMessage *cast.CastMessage) {
	clientConnection.log.Info("handle cast message")
	switch *castMessage.Namespace {
	case heartbeatNamespace:
		clientConnection.handleHeartbeatMessage(*castMessage.PayloadUtf8)
		break
	case receiverNamespace:
		clientConnection.handleReceiverMessage(*castMessage.PayloadUtf8)
		break
	default:
		clientConnection.log.Info("unhandled message", "namespace", *castMessage.Namespace)
	}
}

func (clientConnection *ClientConnection) relayCastMessage(castMessage *cast.CastMessage) {
	clientConnection.log.Info("relay cast message")

	clientConnection.relayClient.SendMessage(castMessage)
}

func NewClientConnection(conn net.Conn, manifest map[string]string, relayClient *Client) *ClientConnection {
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
						// device authentication is always handled locally
						clientConnection.handleDeviceAuthChallenge(manifest)
					} else if relayClient != nil {
						// all other messages are relayed when in relay mode
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

	if relayClient != nil {
		go func() {
			select {
			case castMessage := <-relayClient.Incoming:
				clientConnection.castChannel.Send(castMessage)
			}
		}()
	}

	return &clientConnection
}
