package internal

import (
	"errors"
	"fmt"
	"image"

	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/channel"
)

const androidMirroringAppId = "674A0243"
const chromeMirroringAppId = "0F5096E8"

type Subscription struct {
	clientConnection *ClientConnection
	remoteId         string
}

type Transport struct {
	castTransport CastTransport
	subscriptions []Subscription
}

type Device struct {
	AvailableApps []string
	DeviceModel   string
	FriendlyName  string
	Id            string
	Sessions      map[string]*Session
	Udn           string

	// implementation
	images     chan *image.RGBA
	jpegOutput bool
	log        hclog.Logger
	nextPid    int
	transports map[string]*Transport
}

func (device *Device) forwardCastMessage(castMessage *channel.CastMessage) {
	transport := device.transports[*castMessage.DestinationId]
	if transport == nil {
		device.log.Error("message destination does not exist", "destinationId", *castMessage.DestinationId)
		return
	}

	transport.castTransport.HandleCastMessage(castMessage)
}

//
// Functions to register transports and subscribe to their broadcasts
//

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
	device.transports[castTransport.TransportId()] = &Transport{
		castTransport: castTransport,
		subscriptions: make([]Subscription, 0),
	}
}

//
// Functions to send messages
//

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

func (device *Device) startAndroidMirroringSession(clientId int) {
	transportId := fmt.Sprintf("pid-%d", device.nextPid)
	device.nextPid++

	session := NewSession(androidMirroringAppId, clientId, device, "Android Mirroring", device.jpegOutput, uuid.New().String(), transportId)
	session.Start()

	device.Sessions[session.SessionId] = session
	device.registerTransport(session)
}

func (device *Device) startChromeMirroringSession(clientId int) {
	transportId := fmt.Sprintf("pid-%d", device.nextPid)
	device.nextPid++

	session := NewSession(chromeMirroringAppId, clientId, device, "Chrome Mirroring", device.jpegOutput, uuid.New().String(), transportId)
	session.Start()

	device.Sessions[session.SessionId] = session
	device.registerTransport(session)
}

func (device *Device) startApplication(appId string, clientId int) error {
	for _, session := range device.Sessions {
		if session.AppId == appId {
			return errors.New("application already started")
		}
	}

	switch appId {
	case androidMirroringAppId:
		device.startAndroidMirroringSession(clientId)
		break
	case chromeMirroringAppId:
		device.startChromeMirroringSession(clientId)
		break
	default:
		return errors.New("unsupported app")
	}

	return nil
}

func (device *Device) stopApplication(sessionId string) error {
	session := device.Sessions[sessionId]
	if session == nil {
		return errors.New("session does not exist")
	}

	delete(device.Sessions, sessionId)
	session.Stop()
	return nil
}

func (device *Device) DisplayImage(image *image.RGBA) {
	device.images <- image
}

func NewDevice(images chan *image.RGBA, deviceModel string, friendlyName string, id string, jpegOutput bool, udn string) *Device {
	log := NewLogger(fmt.Sprintf("device (%s)", id))

	// Allow clients to start Android or Chrome mirroring apps
	availableApps := make([]string, 2)
	availableApps[0] = androidMirroringAppId
	availableApps[1] = chromeMirroringAppId

	device := Device{
		AvailableApps: availableApps,
		DeviceModel:   deviceModel,
		FriendlyName:  friendlyName,
		Id:            id,
		Sessions:      map[string]*Session{},
		Udn:           udn,

		// implementation
		images:     images,
		jpegOutput: jpegOutput,
		log:        log,
		nextPid:    1,
		transports: make(map[string]*Transport),
	}

	return &device
}
