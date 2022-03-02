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
	receiverId  string
	relayClient *Client
	senderId    string
}

func (clientConnection *ClientConnection) sendUtf8Message(payload []byte, namespace string) {
	payloadType := cast.CastMessage_STRING
	payloadUtf8 := string(payload)
	protocolVersion := cast.CastMessage_CASTV2_1_0

	castMessage := cast.CastMessage{
		DestinationId:   &clientConnection.senderId,
		Namespace:       &namespace,
		PayloadUtf8:     &payloadUtf8,
		PayloadType:     &payloadType,
		ProtocolVersion: &protocolVersion,
		SourceId:        &clientConnection.receiverId,
	}

	clientConnection.log.Info("sending", "castMessage", castMessage.String())

	clientConnection.castChannel.Send(&castMessage)
}

func (clientConnection *ClientConnection) handleCastMessage(castMessage *cast.CastMessage) {
	clientConnection.log.Info("handle cast message")

	// only handle connect messages if not already connected
	if clientConnection.receiverId == "0" && clientConnection.senderId == "0" {
		if *castMessage.Namespace == connectNamespace {
			clientConnection.handleConnectMessage(*castMessage.PayloadUtf8)
			clientConnection.senderId = *castMessage.SourceId
			clientConnection.receiverId = *castMessage.DestinationId
		} else {
			clientConnection.log.Error("not connected; ignoring message", "namespace", *castMessage.Namespace)
		}
		return
	}

	switch *castMessage.Namespace {
	case connectNamespace:
		clientConnection.log.Error("already connected; ignoring connection message")
		return
	case heartbeatNamespace:
		responsePayload := clientConnection.handleHeartbeatMessage(*castMessage.PayloadUtf8)
		if responsePayload != nil {
			clientConnection.sendUtf8Message(responsePayload, heartbeatNamespace)
		}
		return
	case receiverNamespace:
		clientConnection.handleReceiverMessage(*castMessage.PayloadUtf8)
		return
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
		receiverId:  "0",
		relayClient: relayClient,
		senderId:    "0",
	}

	go func() {
		defer func() {
			_ = conn.Close()
			log.Info("connection closed")
		}()

		for {
			select {
			case castMessage, ok := <-castChannel.Messages:
				if !ok {
					log.Info("channel closed")
					return
				}

				if castMessage != nil {
					log.Info("received", "message", castMessage)
					if *castMessage.Namespace == cast.DeviceAuthNamespace {
						// device authentication is always handled locally
						clientConnection.handleDeviceAuthChallenge(manifest)
					} else if relayClient == nil {
						clientConnection.handleCastMessage(castMessage)
					} else {
						// all other messages are relayed when in relay mode
						clientConnection.relayCastMessage(castMessage)
					}
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
