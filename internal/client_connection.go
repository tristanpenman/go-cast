package internal

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"

	"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/proto"

	"github.com/tristanpenman/go-cast/internal/channel"
)

type ClientConnection struct {
	castChannel CastChannel
	conn        net.Conn
	device      *Device
	id          int
	log         hclog.Logger
	sessions    map[string]*Session
}

func (clientConnection *ClientConnection) sendBinary(namespace string, payloadBinary []byte, sourceId string, destinationId string) {
	payloadType := channel.CastMessage_BINARY
	protocolVersion := channel.CastMessage_CASTV2_1_0
	castMessage := channel.CastMessage{
		DestinationId:   &destinationId,
		Namespace:       &namespace,
		PayloadBinary:   payloadBinary,
		PayloadType:     &payloadType,
		ProtocolVersion: &protocolVersion,
		SourceId:        &sourceId,
	}

	if clientConnection.log.IsDebug() {
		clientConnection.log.Debug("sending", "castMessage", castMessage.String())
	} else {
		clientConnection.log.Info("sending",
			"namespace", *castMessage.Namespace,
			"sourceId", *castMessage.SourceId,
			"destinationId", *castMessage.DestinationId,
			"payloadType", "BINARY")
	}

	clientConnection.castChannel.Send(&castMessage)
}

func (clientConnection *ClientConnection) sendUtf8(namespace string, payloadUtf8 *string, sourceId string, destinationId string) {
	payloadType := channel.CastMessage_STRING
	protocolVersion := channel.CastMessage_CASTV2_1_0
	castMessage := channel.CastMessage{
		DestinationId:   &destinationId,
		Namespace:       &namespace,
		PayloadUtf8:     payloadUtf8,
		PayloadType:     &payloadType,
		ProtocolVersion: &protocolVersion,
		SourceId:        &sourceId,
	}

	if clientConnection.log.IsDebug() {
		clientConnection.log.Debug("sending", "castMessage", castMessage.String())
	} else {
		clientConnection.log.Info("sending",
			"namespace", *castMessage.Namespace,
			"sourceId", *castMessage.SourceId,
			"destinationId", *castMessage.DestinationId,
			"payloadType", "STRING",
			"payloadUtf8", *castMessage.PayloadUtf8)
	}

	clientConnection.castChannel.Send(&castMessage)
}

type connectRequest struct {
	ConnType json.Number `json:"connType"`
}

func (clientConnection *ClientConnection) handleConnectMessage(castMessage *channel.CastMessage) {
	var request connectRequest
	err := json.Unmarshal([]byte(*castMessage.PayloadUtf8), &request)
	if err != nil {
		clientConnection.log.Error("failed to connect data", "err", err)
		return
	}

	clientConnection.device.registerSubscription(clientConnection, *castMessage.SourceId, *castMessage.DestinationId)
}

func (clientConnection *ClientConnection) handleCastMessage(castMessage *channel.CastMessage) {
	if *castMessage.Namespace == connectionNamespace {
		// CONNECT messages are special, and are essentially used to
		// subscribe to status updates from a receiver
		clientConnection.handleConnectMessage(castMessage)
	} else {
		// All other messages can be forwarded via the device hub
		clientConnection.device.forwardCastMessage(castMessage)
	}
}

func (clientConnection *ClientConnection) handleDeviceAuthChallenge(message *channel.CastMessage, manifest map[string]string) {
	var deviceAuthMessage channel.DeviceAuthMessage
	err := proto.Unmarshal(message.PayloadBinary, &deviceAuthMessage)
	if err != nil {
		clientConnection.log.Error("failed to parse device auth message", "err", err)
		return
	}

	// certificates are in PEM format
	pu, _ := pem.Decode([]byte(manifest["pu"]))
	ica, _ := pem.Decode([]byte(manifest["ica"]))
	cpu, _ := pem.Decode([]byte(manifest["cpu"]))

	// TODO: check that we don't have any remaining data in `rest`

	// Signature is just base64
	sig, _ := base64.StdEncoding.DecodeString(manifest["sig"])

	hashAlgorithm := DetectAlgorithm(cpu, pu, sig)
	if hashAlgorithm == nil {
		clientConnection.log.Warn("failed to identify hashing algorithm")
	} else if *hashAlgorithm == channel.HashAlgorithm_SHA1 {
		clientConnection.log.Info("detected SHA1 as hashing algorithm")
	} else {
		clientConnection.log.Info("detected SHA256 as hashing algorithm")
	}

	intermediateCertificate := [][]byte{ica.Bytes}

	deviceAuthMessage = channel.DeviceAuthMessage{
		Response: &channel.AuthResponse{
			Signature:               sig,
			ClientAuthCertificate:   cpu.Bytes,
			IntermediateCertificate: intermediateCertificate,
			HashAlgorithm:           hashAlgorithm,
		},
	}

	payloadBinary, err := proto.Marshal(&deviceAuthMessage)
	if err != nil {
		clientConnection.log.Error("failed to encode device auth response", "err", err)
		return
	}

	clientConnection.sendBinary(deviceAuthNamespace, payloadBinary, *message.DestinationId, *message.SourceId)
}

func NewClientConnection(
	device *Device,
	conn net.Conn,
	id int,
	manifest map[string]string,
) *ClientConnection {
	log := NewLogger(fmt.Sprintf("client-connection (%d)", id))

	castChannel := CreateCastChannel(conn, log)

	clientConnection := ClientConnection{
		castChannel: castChannel,
		conn:        conn,
		device:      device,
		id:          id,
		sessions:    make(map[string]*Session),
		log:         log,
	}

	receiver := NewReceiver(device, "receiver-0", clientConnection.id)
	device.registerTransport(receiver)
	device.registerSubscription(&clientConnection, "sender-0", "receiver-0")

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
					if log.IsDebug() {
						log.Debug("received", "message", castMessage.String())
					} else if *castMessage.PayloadType == channel.CastMessage_BINARY {
						log.Info("received",
							"namespace", *castMessage.Namespace,
							"sourceId", *castMessage.SourceId,
							"destinationId", *castMessage.DestinationId,
							"payloadType", "BINARY")
					} else {
						log.Info("received",
							"namespace", *castMessage.Namespace,
							"sourceId", *castMessage.SourceId,
							"destinationId", *castMessage.DestinationId,
							"payloadType", "STRING",
							"payloadUtf8", *castMessage.PayloadUtf8)
					}

					if *castMessage.Namespace == deviceAuthNamespace {
						// device authentication is always handled locally
						clientConnection.handleDeviceAuthChallenge(castMessage, manifest)
					} else {
						clientConnection.handleCastMessage(castMessage)
					}
				}
			}
		}
	}()

	return &clientConnection
}
