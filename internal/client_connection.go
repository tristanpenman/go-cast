package internal

import (
	"fmt"
	"net"

	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/cast"
)

const debugNamespace = "urn:x-cast:com.google.cast.debug"
const remotingNamespace = "urn:x-cast:com.google.cast.remoting"

type Namespace struct {
	Name string `json:"name"`
}

type Application struct {
	AppId       string      `json:"appId"`
	DisplayName string      `json:"displayName"`
	Namespaces  []Namespace `json:"namespaces"`
	SessionId   string      `json:"sessionId"`
	StatusText  string      `json:"statusText"`
	TransportId string      `json:"transportId"`
}

type ClientConnection struct {
	applications []Application
	castChannel  cast.CastChannel
	conn         net.Conn
	log          hclog.Logger
	receiverId   string
	relayClient  *Client
	senderId     string
}

func defaultApplications() []Application {
	namespaces := make([]Namespace, 3)
	namespaces[0] = Namespace{Name: receiverNamespace}
	namespaces[1] = Namespace{Name: debugNamespace}
	namespaces[2] = Namespace{Name: remotingNamespace}

	applications := make([]Application, 1)
	applications[0] = Application{
		AppId:       "E8C28D3C",
		DisplayName: "Backdrop",
		Namespaces:  namespaces,
		SessionId:   "AD3DFC60-A6AE-4532-87AF-18504DA22607",
		StatusText:  "",
		TransportId: "pid-22607",
	}

	return applications
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

func NewClientConnection(conn net.Conn, id string, manifest map[string]string, relayClient *Client) *ClientConnection {
	var log = NewLogger(fmt.Sprintf("client-connection (%s)", id))

	castChannel := cast.CreateCastChannel(conn, log)

	clientConnection := ClientConnection{
		applications: defaultApplications(),
		castChannel:  castChannel,
		conn:         conn,
		log:          log,
		receiverId:   "0",
		relayClient:  relayClient,
		senderId:     "0",
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
