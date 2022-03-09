package internal

import (
	"encoding/base64"
	"encoding/pem"

	"github.com/golang/protobuf/proto"

	"github.com/tristanpenman/go-cast/internal/cast"
)

func (clientConnection *ClientConnection) handleDeviceAuthChallenge(message *cast.CastMessage, manifest map[string]string) {
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

	deviceAuthMessage := &cast.DeviceAuthMessage{
		Response: &cast.AuthResponse{
			Signature:               sig,
			ClientAuthCertificate:   platform.Bytes,
			IntermediateCertificate: intermediateCertificate,
			HashAlgorithm:           &hashAlgorithm,
		},
	}

	payloadBinary, err := proto.Marshal(deviceAuthMessage)
	if err != nil {
		clientConnection.log.Error("failed to encode device auth response", "err", err)
		return
	}

	clientConnection.sendBinary(deviceAuthNamespace, payloadBinary, *message.DestinationId, *message.SourceId)
}

func (client *Client) sendDeviceAuthChallenge() bool {
	deviceAuthMessage := &cast.DeviceAuthMessage{
		Challenge: &cast.AuthChallenge{},
	}

	payloadBinary, err := proto.Marshal(deviceAuthMessage)
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
		PayloadBinary:   payloadBinary,
		PayloadType:     &payloadType,
		ProtocolVersion: &protocolVersion,
		SourceId:        &sourceId,
	}

	return client.castChannel.Send(&message)
}
