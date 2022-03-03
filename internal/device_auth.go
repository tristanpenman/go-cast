package internal

import (
	"encoding/base64"
	"encoding/pem"

	"github.com/golang/protobuf/proto"

	"github.com/tristanpenman/go-cast/internal/cast"
)

func (clientConnection *ClientConnection) handleDeviceAuthChallenge(manifest map[string]string) bool {
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

	namespace := deviceAuthNamespace
	payloadType := cast.CastMessage_BINARY
	protocolVersion := cast.CastMessage_CASTV2_1_0
	destinationId := "sender-0"
	sourceId := "receiver-0"
	message := cast.CastMessage{
		DestinationId:   &destinationId,
		Namespace:       &namespace,
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

func (client *Client) sendDeviceAuthChallenge() bool {
	payloadBytes, err := proto.Marshal(&cast.DeviceAuthMessage{
		Challenge: &cast.AuthChallenge{},
	})

	if err != nil {
		client.log.Error("failed to encode device auth challenge", "err", err)
		return false
	}

	namespace := deviceAuthNamespace
	payloadType := cast.CastMessage_BINARY
	protocolVersion := cast.CastMessage_CASTV2_1_0
	sourceId := "sender-0"
	destinationId := "receiver-0"
	message := cast.CastMessage{
		DestinationId:   &destinationId,
		Namespace:       &namespace,
		PayloadBinary:   payloadBytes,
		PayloadType:     &payloadType,
		ProtocolVersion: &protocolVersion,
		SourceId:        &sourceId,
	}

	return client.castChannel.Send(&message)
}
