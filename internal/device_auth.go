package internal

import (
	"encoding/base64"
	"encoding/pem"

	"github.com/tristanpenman/go-cast/internal/channel"
	"google.golang.org/protobuf/proto"
)

func (clientConnection *ClientConnection) handleDeviceAuthChallenge(message *channel.CastMessage, manifest map[string]string) {
	var deviceAuthMessage channel.DeviceAuthMessage
	err := proto.Unmarshal(message.PayloadBinary, &deviceAuthMessage)
	if err != nil {
		clientConnection.log.Error("failed to parse device auth message", "err", err)
		return
	}

	// intermediate and platform certs are in PEM format
	// TODO: check that we don't have any remaining data in `rest`
	ica, _ := pem.Decode([]byte(manifest["ica"]))
	platform, _ := pem.Decode([]byte(manifest["cpu"]))

	// Signature is just base64
	sig, _ := base64.StdEncoding.DecodeString(manifest["sig"])

	// TODO: is there a tidier way to do this?
	intermediateCertificate := make([][]byte, 1)
	intermediateCertificate[0] = ica.Bytes

	hashAlgorithm := channel.HashAlgorithm_SHA256

	crl := make([]byte, 0)
	deviceAuthMessage = channel.DeviceAuthMessage{
		Response: &channel.AuthResponse{
			Crl:                     crl,
			Signature:               sig,
			ClientAuthCertificate:   platform.Bytes,
			IntermediateCertificate: intermediateCertificate,
			HashAlgorithm:           &hashAlgorithm,
		},
	}

	payloadBinary, err := proto.Marshal(&deviceAuthMessage)
	if err != nil {
		clientConnection.log.Error("failed to encode device auth response", "err", err)
		return
	}

	clientConnection.sendBinary(deviceAuthNamespace, payloadBinary, *message.DestinationId, *message.SourceId)
}

func (client *Client) sendDeviceAuthChallenge() bool {
	deviceAuthMessage := &channel.DeviceAuthMessage{
		Challenge: &channel.AuthChallenge{},
	}

	payloadBinary, err := proto.Marshal(deviceAuthMessage)
	if err != nil {
		client.log.Error("failed to encode device auth challenge", "err", err)
		return false
	}

	namespace := deviceAuthNamespace
	payloadType := channel.CastMessage_BINARY
	protocolVersion := channel.CastMessage_CASTV2_1_0
	sourceId := "sender-0"
	destinationId := "receiver-0"
	message := channel.CastMessage{
		DestinationId:   &destinationId,
		Namespace:       &namespace,
		PayloadBinary:   payloadBinary,
		PayloadType:     &payloadType,
		ProtocolVersion: &protocolVersion,
		SourceId:        &sourceId,
	}

	return client.castChannel.Send(&message)
}
