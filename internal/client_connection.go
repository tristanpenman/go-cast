package internal

import (
	"encoding/base64"
	"encoding/pem"
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

func (clientConnection *ClientConnection) sendDeviceAuthResponse(manifest map[string]string) bool {
	// intermediate and platform certs are in PEM format
	// TODO: check that we don't have any remaining data in `rest`
	ica, _ := pem.Decode([]byte(manifest["ica"]))
	platform, _ := pem.Decode([]byte(manifest["cpu"]))

	// Signature is just base64
	sig, _ := base64.StdEncoding.DecodeString(manifest["sig"])

	// TODO: is there a tidier way to do this?
	intermediateCertificate := make([][]byte, 1)
	intermediateCertificate[0] = ica.Bytes

	hashAlgorithm := cast.HashAlgorithm_SHA256

	payloadBytes, err := proto.Marshal(&cast.DeviceAuthMessage{
		Response: &cast.AuthResponse{
			Signature:               sig,
			ClientAuthCertificate:   platform.Bytes,
			IntermediateCertificate: intermediateCertificate,
			HashAlgorithm:           &hashAlgorithm,
		},
	})

	if err != nil {
		clientConnection.log.Error("failed to encode device auth response", "err", err)
		return false
	}

	payloadType := cast.CastMessage_BINARY
	protocolVersion := cast.CastMessage_CASTV2_1_0
	destinationId := "sender-0"
	sourceId := "receiver-0"
	message := cast.CastMessage{
		DestinationId:   &destinationId,
		Namespace:       &cast.DeviceAuthNamespace,
		PayloadBinary:   payloadBytes,
		PayloadType:     &payloadType,
		ProtocolVersion: &protocolVersion,
		SourceId:        &sourceId,
	}

	if clientConnection.log.IsDebug() {
		clientConnection.log.Debug("sending device auth response", "message", message.String())
	} else {
		clientConnection.log.Info("sending device auth response")
	}

	return clientConnection.castChannel.Send(&message)
}

func (clientConnection *ClientConnection) relayCastMessage(castMessage *cast.CastMessage) {
	clientConnection.log.Info("relay cast message")

	clientConnection.relayClient.SendMessage(castMessage)
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
						clientConnection.sendDeviceAuthResponse(manifest)
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
