package client

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	// third-party
	"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/proto"

	// internal
	"github.com/tristanpenman/go-cast/internal/channel"
	"github.com/tristanpenman/go-cast/internal/common"
	"github.com/tristanpenman/go-cast/internal/transport"
)

type Client struct {
	castChannel  transport.CastChannel
	conn         net.Conn
	deviceAuthWg *sync.WaitGroup
	Incoming     chan *channel.CastMessage
	log          hclog.Logger
}

func (client *Client) sendDeviceAuthChallenge(deviceAuthWg *sync.WaitGroup) bool {
	client.deviceAuthWg = deviceAuthWg

	deviceAuthMessage := &channel.DeviceAuthMessage{
		Challenge: &channel.AuthChallenge{},
	}

	payloadBinary, err := proto.Marshal(deviceAuthMessage)
	if err != nil {
		client.log.Error("failed to encode device auth challenge", "err", err)
		return false
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

	return client.castChannel.Send(&message)
}

func (client *Client) verifyDeviceAuthResponse(payloadBytes []byte) {
	client.log.Info(string(payloadBytes))

	if client.deviceAuthWg != nil {
		client.deviceAuthWg.Done()
		client.deviceAuthWg = nil
	}
}

// NewClient connects to a Cast receiver and starts its inbound message loop.
func NewClient(hostname string, port uint, authChallenge bool, wg *sync.WaitGroup) (*Client, error) {
	var log = common.NewLogger("client")

	addr := fmt.Sprintf("%s:%d", hostname, port)
	log.Info(fmt.Sprintf("addr: %s", addr))

	config := tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		return nil, fmt.Errorf("connect to receiver: %w", err)
	}

	log.Info("Connected")

	castChannel := transport.NewCastChannel(conn, log)

	client := Client{
		castChannel: castChannel,
		conn:        conn,
		Incoming:    make(chan *channel.CastMessage, 64),
		log:         log,
	}

	if authChallenge {
		var deviceAuthWg sync.WaitGroup
		client.sendDeviceAuthChallenge(&deviceAuthWg)
	}

	go func() {
		for castMessage := range castChannel.Messages {
			if castMessage != nil {
				if log.IsDebug() {
					log.Debug("received message", "content", castMessage)
				} else {
					log.Info("received message", "namespace", *castMessage.Namespace)
				}

				if *castMessage.Namespace == common.DeviceAuthNamespace {
					client.verifyDeviceAuthResponse(castMessage.PayloadBinary)
					continue
				}
			}

			client.Incoming <- castMessage
		}

		log.Info("channel closed")
		close(client.Incoming)
		_ = conn.Close()
		if wg != nil {
			wg.Done()
		}
	}()

	return &client, nil
}

func (client *Client) SendMessage(castMessage *channel.CastMessage) {
	client.castChannel.Send(castMessage)
}

func (client *Client) Close() error {
	return client.conn.Close()
}
