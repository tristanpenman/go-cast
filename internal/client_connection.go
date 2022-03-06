package internal

import (
	"fmt"
	"net"

	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/cast"
)

const debugNamespace = "urn:x-cast:com.google.cast.debug"
const deviceAuthNamespace = "urn:x-cast:com.google.cast.tp.deviceauth"
const heartbeatNamespace = "urn:x-cast:com.google.cast.tp.heartbeat"
const mediaNamespace = "urn:x-cast:com.google.cast.media"
const receiverNamespace = "urn:x-cast:com.google.cast.receiver"
const remotingNamespace = "urn:x-cast:com.google.cast.remoting"
const setupNamespace = "urn:x-cast:com.google.cast.setup"

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
	device       Device
	log          hclog.Logger
	mirrors      map[string]*Mirror
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

func (clientConnection *ClientConnection) startApplication(appId string) bool {
	for _, application := range clientConnection.applications {
		if application.AppId == appId {
			clientConnection.log.Warn("application already started", "appId", appId)
			return true
		}
	}

	namespaces := make([]Namespace, 2)
	namespaces[0] = Namespace{Name: receiverNamespace}
	namespaces[1] = Namespace{Name: debugNamespace}

	var application Application
	switch appId {
	case androidMirroringAppId:
		application = Application{
			AppId:       appId,
			DisplayName: "Android Mirroring",
			Namespaces:  namespaces,
			SessionId:   "835ff891-f76f-4a04-8618-a5dc95477075",
			StatusText:  "",
			TransportId: "web-5",
		}
		break
	case chromeMirroringAppId:
		application = Application{
			AppId:       appId,
			DisplayName: "Chrome Mirroring",
			Namespaces:  namespaces,
			SessionId:   "7E2FF513-CDF6-9A91-2B28-3E3DE7BAC174",
			StatusText:  "",
			TransportId: "web-5",
		}
		break
	default:
		clientConnection.log.Error("Unsupported app", "appId", appId)
		return false
	}

	// create new mirroring session
	mirror := NewMirror()
	clientConnection.mirrors[application.SessionId] = mirror
	clientConnection.applications = append(clientConnection.applications, application)

	return true
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
	case setupNamespace:
		responsePayload := clientConnection.handleSetupMessage(*castMessage.PayloadUtf8)
		if responsePayload != nil {
			clientConnection.sendUtf8Message(responsePayload, setupNamespace)
		}
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

func NewClientConnection(device Device, conn net.Conn, id int, manifest map[string]string, relayClient *Client) *ClientConnection {
	var log = NewLogger(fmt.Sprintf("client-connection (%d)", id))

	castChannel := cast.CreateCastChannel(conn, log)

	clientConnection := ClientConnection{
		applications: defaultApplications(),
		castChannel:  castChannel,
		conn:         conn,
		device:       device,
		mirrors:      make(map[string]*Mirror),
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
