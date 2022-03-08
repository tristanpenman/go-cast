package internal

import (
	"fmt"
	"net"

	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/cast"
)

type ClientConnection struct {
	castChannel cast.CastChannel
	conn        net.Conn
	device      *Device
	id          int
	log         hclog.Logger
	relayClient *Client
	sessions    map[string]*Session

	// virtual connection state
	connected  bool
	receiverId string
	senderId   string
}

func (clientConnection *ClientConnection) sendBinary(namespace string, payloadBinary []byte) {
	payloadType := cast.CastMessage_BINARY
	protocolVersion := cast.CastMessage_CASTV2_1_0
	castMessage := cast.CastMessage{
		DestinationId:   &clientConnection.senderId,
		Namespace:       &namespace,
		PayloadBinary:   payloadBinary,
		PayloadType:     &payloadType,
		ProtocolVersion: &protocolVersion,
		SourceId:        &clientConnection.receiverId,
	}

	clientConnection.log.Info("sending", "castMessage", castMessage.String())
	clientConnection.castChannel.Send(&castMessage)
}

func (clientConnection *ClientConnection) sendUtf8(namespace string, payloadUtf8 *string) {
	payloadType := cast.CastMessage_STRING
	protocolVersion := cast.CastMessage_CASTV2_1_0
	castMessage := cast.CastMessage{
		DestinationId:   &clientConnection.senderId,
		Namespace:       &namespace,
		PayloadUtf8:     payloadUtf8,
		PayloadType:     &payloadType,
		ProtocolVersion: &protocolVersion,
		SourceId:        &clientConnection.receiverId,
	}

	clientConnection.log.Info("sending", "castMessage", castMessage.String())
	clientConnection.castChannel.Send(&castMessage)
}

func (clientConnection *ClientConnection) handleCastMessage(castMessage *cast.CastMessage) {
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
		clientConnection.handleHeartbeatMessage(*castMessage.PayloadUtf8)
		return
	case receiverNamespace:
		clientConnection.handleReceiverMessage(*castMessage.PayloadUtf8)
		return
	case setupNamespace:
		clientConnection.handleSetupMessage(*castMessage.PayloadUtf8)
		return
	// unsupported namespaces
	case debugNamespace:
	case mediaNamespace:
		clientConnection.log.Warn("received message for known but unsupported namespace", "namespace", *castMessage.Namespace)
		break
	default:
		clientConnection.log.Info("received message for unknown namespace", "namespace", *castMessage.Namespace)
	}
}

func (clientConnection *ClientConnection) relayCastMessage(castMessage *cast.CastMessage) {
	clientConnection.log.Info("relay cast message")

	clientConnection.relayClient.SendMessage(castMessage)
}

func NewClientConnection(device *Device, conn net.Conn, id int, manifest map[string]string, relayClient *Client) *ClientConnection {
	var log = NewLogger(fmt.Sprintf("client-connection (%d)", id))

	castChannel := cast.CreateCastChannel(conn, log)

	clientConnection := ClientConnection{
		castChannel: castChannel,
		conn:        conn,
		device:      device,
		id:          id,
		sessions:    make(map[string]*Session),
		log:         log,
		relayClient: relayClient,

		// virtual connection state
		connected:  false,
		receiverId: "0",
		senderId:   "0",
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
					if *castMessage.Namespace == deviceAuthNamespace {
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
