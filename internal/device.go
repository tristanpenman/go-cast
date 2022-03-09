package internal

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/cast"
)

const androidMirroringAppId = "674A0243"
const backdropAppId = "E8C28D3C"
const chromeMirroringAppId = "0F5096E8"

type Application struct {
	AppId       string
	DisplayName string
	Namespaces  []string
	SessionId   string
	StatusText  string
	TransportId string
}

type Subscription struct {
	clientConnection *ClientConnection
	remoteId         string
}

type Transport struct {
	castTransport CastTransport
	subscriptions []Subscription
}

type Device struct {
	Applications  map[string]*Application
	AvailableApps []string
	DeviceModel   string
	FriendlyName  string
	Id            string

	// implementation
	log        hclog.Logger
	transports map[string]*Transport
}

func (device *Device) forwardCastMessage(castMessage *cast.CastMessage) {
	transport := device.transports[*castMessage.DestinationId]
	if transport == nil {
		device.log.Error("message destination does not exist", "destinationId", *castMessage.DestinationId)
		return
	}

	transport.castTransport.HandleCastMessage(castMessage)
}

func (device *Device) broadcastBinary(namespace string, payloadBinary []byte, sourceId string) {
	transport := device.transports[sourceId]
	if transport == nil {
		device.log.Error("source transport is not registered", "sourceId", sourceId)
		return
	}

	var clientConnections = map[*ClientConnection]bool{}
	for _, subscription := range transport.subscriptions {
		clientConnections[subscription.clientConnection] = true
	}

	for clientConnection, _ := range clientConnections {
		clientConnection.sendBinary(namespace, payloadBinary, sourceId, "*")
	}
}

func (device *Device) broadcastUtf8(namespace string, payloadUtf8 *string, sourceId string) {
	transport := device.transports[sourceId]
	if transport == nil {
		device.log.Error("source transport is not registered", "sourceId", sourceId)
		return
	}

	var clientConnections = map[*ClientConnection]bool{}
	for _, subscription := range transport.subscriptions {
		clientConnections[subscription.clientConnection] = true
	}

	for clientConnection, _ := range clientConnections {
		clientConnection.sendUtf8(namespace, payloadUtf8, sourceId, "*")
	}
}

func (device *Device) registerSubscription(clientConnection *ClientConnection, remoteId string, localId string) {
	// localId should be a valid transport
	// remoteId can be anything really
	// clientConnection is just how we get to the remote
	// from this, we can construct a Peer

	transport := device.transports[localId]
	if transport == nil {
		device.log.Error("attempt to register subscription for non-existent local transport ID", "localId", localId)
		return
	}

	subscription := Subscription{
		clientConnection: clientConnection,
		remoteId:         remoteId,
	}

	transport.subscriptions = append(transport.subscriptions, subscription)
}

func (device *Device) registerTransport(castTransport CastTransport) {
	device.transports[castTransport.Id()] = &Transport{
		castTransport: castTransport,
		subscriptions: make([]Subscription, 0),
	}
}

func (device *Device) sendBinary(namespace string, payloadBinary []byte, sourceId string, destinationId string) {
	transport := device.transports[sourceId]
	if transport == nil {
		device.log.Error("attempt to send from unregistered transport", "sourceId", sourceId)
		return
	}

	var clientConnections = map[*ClientConnection]bool{}
	for _, subscription := range transport.subscriptions {
		if subscription.remoteId == destinationId {
			clientConnections[subscription.clientConnection] = true
		}
	}

	for clientConnection, _ := range clientConnections {
		clientConnection.sendBinary(namespace, payloadBinary, sourceId, destinationId)
	}
}

func (device *Device) sendUtf8(namespace string, payloadUtf8 *string, sourceId string, destinationId string) {
	transport := device.transports[sourceId]
	if transport == nil {
		device.log.Error("attempt to send from unregistered transport", "sourceId", sourceId)
		return
	}

	var clientConnections = map[*ClientConnection]bool{}
	for _, subscription := range transport.subscriptions {
		if subscription.remoteId == destinationId {
			clientConnections[subscription.clientConnection] = true
		}
	}

	for clientConnection, _ := range clientConnections {
		clientConnection.sendUtf8(namespace, payloadUtf8, sourceId, destinationId)
	}
}

func (device *Device) startApplication(appId string) error {
	for _, application := range device.Applications {
		if application.AppId == appId {
			return errors.New("application already started")
		}
	}

	namespaces := make([]string, 2)
	namespaces[0] = receiverNamespace
	namespaces[1] = debugNamespace

	var application Application
	switch appId {
	case androidMirroringAppId:
		application = Application{
			AppId:       androidMirroringAppId,
			DisplayName: "Android Mirroring",
			Namespaces:  namespaces,
			SessionId:   "835ff891-f76f-4a04-8618-a5dc95477075",
			StatusText:  "",
			TransportId: "web-5",
		}
		break
	case chromeMirroringAppId:
		application = Application{
			AppId:       chromeMirroringAppId,
			DisplayName: "Chrome Mirroring",
			Namespaces:  namespaces,
			SessionId:   "7E2FF513-CDF6-9A91-2B28-3E3DE7BAC174",
			StatusText:  "",
			TransportId: "web-5",
		}
		break
	default:
		return errors.New("unsupported app")
	}

	device.Applications[application.SessionId] = &application

	return nil
}

func NewDevice(deviceModel string, friendlyName string, id string) *Device {
	log := NewLogger(fmt.Sprintf("device (%s)", id))

	// Namespaces covered by the Backdrop appllication
	namespaces := make([]string, 3)
	namespaces[0] = receiverNamespace
	namespaces[1] = debugNamespace
	namespaces[2] = remotingNamespace

	// Create session for Backdrop application
	sessionId := "AD3DFC60-A6AE-4532-87AF-18504DA22607"
	applications := make(map[string]*Application, 1)
	applications[sessionId] = &Application{
		AppId:       backdropAppId,
		DisplayName: "Backdrop",
		Namespaces:  namespaces,
		SessionId:   sessionId,
		StatusText:  "",
		TransportId: "pid-22607",
	}

	// Allow clients to start Android or Chrome mirroring apps
	availableApps := make([]string, 2)
	availableApps[0] = androidMirroringAppId
	availableApps[1] = chromeMirroringAppId

	return &Device{
		Applications:  applications,
		AvailableApps: availableApps,
		DeviceModel:   deviceModel,
		FriendlyName:  friendlyName,
		Id:            id,

		// implementation
		log:        log,
		transports: make(map[string]*Transport),
	}
}
