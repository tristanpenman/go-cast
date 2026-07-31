package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	// third-party
	"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/proto"

	// internal
	"github.com/tristanpenman/go-cast/internal/channel"
	"github.com/tristanpenman/go-cast/internal/common"
	"github.com/tristanpenman/go-cast/internal/transport"
)

type ServerConnection struct {
	castChannel     transport.CastChannel
	conn            net.Conn
	peerCertificate []byte
	authResult      chan error
	authOnce        sync.Once
	Incoming        chan *channel.CastMessage
	log             hclog.Logger
}

func (client *ServerConnection) sendDeviceAuthChallenge() error {
	deviceAuthMessage := &channel.DeviceAuthMessage{
		Challenge: &channel.AuthChallenge{},
	}

	payloadBinary, err := proto.Marshal(deviceAuthMessage)
	if err != nil {
		return fmt.Errorf("encode device-auth challenge: %w", err)
	}

	namespace := common.DeviceAuthNamespace
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

	if !client.castChannel.Send(&message) {
		return errors.New("send device-auth challenge")
	}
	return nil
}

func (client *ServerConnection) completeDeviceAuth(err error) {
	client.authOnce.Do(func() {
		if client.authResult != nil {
			client.authResult <- err
		}
	})
}

// NewClient connects to a Cast receiver and starts its inbound message loop.
func NewClient(hostname string, port uint, authChallenge bool, wg *sync.WaitGroup) (*ServerConnection, error) {
	var log = common.NewLogger("client")

	addr := fmt.Sprintf("%s:%d", hostname, port)
	log.Info(fmt.Sprintf("addr: %s", addr))

	config := tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		return nil, fmt.Errorf("connect to receiver: %w", err)
	}
	connectionState := conn.ConnectionState()
	if len(connectionState.PeerCertificates) == 0 {
		_ = conn.Close()
		return nil, errors.New("connect to receiver: TLS peer did not provide a certificate")
	}

	log.Info("Connected")

	castChannel := transport.NewCastChannel(conn, log)

	client := ServerConnection{
		castChannel:     castChannel,
		conn:            conn,
		peerCertificate: connectionState.PeerCertificates[0].Raw,
		Incoming:        make(chan *channel.CastMessage, 64),
		log:             log,
	}

	if authChallenge {
		client.authResult = make(chan error, 1)
	}

	go func() {
		for castMessage := range castChannel.Messages {
			if castMessage != nil {
				if castMessage.PayloadUtf8 != nil {
					log.Info("received message",
						"namespace", castMessage.GetNamespace(),
						"source", castMessage.GetSourceId(),
						"destination", castMessage.GetDestinationId(),
						"payload", castMessage.GetPayloadUtf8())
				} else {
					// Device-auth payloads contain certificates and signatures. Their
					// size is useful diagnostically; dumping the binary content is not.
					log.Info("received binary message",
						"namespace", castMessage.GetNamespace(),
						"source", castMessage.GetSourceId(),
						"destination", castMessage.GetDestinationId(),
						"bytes", len(castMessage.GetPayloadBinary()))
				}

				if castMessage.GetNamespace() == common.DeviceAuthNamespace {
					err := verifyDeviceAuthResponse(castMessage.PayloadBinary, client.peerCertificate, nil, time.Now())
					client.completeDeviceAuth(err)
					continue
				}
			}

			client.Incoming <- castMessage
		}

		log.Info("channel closed")
		client.completeDeviceAuth(errors.New("connection closed during device authentication"))
		close(client.Incoming)
		_ = conn.Close()
		if wg != nil {
			wg.Done()
		}
	}()

	if authChallenge {
		if err := client.sendDeviceAuthChallenge(); err != nil {
			_ = client.Close()
			return nil, err
		}
		select {
		case err := <-client.authResult:
			if err != nil {
				_ = client.Close()
				return nil, fmt.Errorf("authenticate receiver: %w", err)
			}
			log.Info("device authentication succeeded")
		case <-time.After(5 * time.Second):
			_ = client.Close()
			return nil, errors.New("authenticate receiver: timed out waiting for device-auth response")
		}
	}

	return &client, nil
}

func (client *ServerConnection) SendMessage(castMessage *channel.CastMessage) {
	if castMessage == nil {
		return
	}
	if castMessage.PayloadUtf8 != nil {
		client.log.Info("sending message",
			"namespace", castMessage.GetNamespace(),
			"source", castMessage.GetSourceId(),
			"destination", castMessage.GetDestinationId(),
			"payload", castMessage.GetPayloadUtf8())
	} else {
		client.log.Info("sending binary message",
			"namespace", castMessage.GetNamespace(),
			"source", castMessage.GetSourceId(),
			"destination", castMessage.GetDestinationId(),
			"bytes", len(castMessage.GetPayloadBinary()))
	}
	client.castChannel.Send(castMessage)
}

func (client *ServerConnection) Close() error {
	return client.conn.Close()
}
