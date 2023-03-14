package internal

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/proto"

	"github.com/tristanpenman/go-cast/internal/channel"
)

type Client struct {
	castChannel  CastChannel
	conn         net.Conn
	deviceAuthWg *sync.WaitGroup
	log          hclog.Logger
	//Incoming     chan *channel.CastMessage
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

func (client *Client) verifyDeviceAuthResponse(payloadBytes []byte) {
	client.log.Info(string(payloadBytes))

	if client.deviceAuthWg != nil {
		client.deviceAuthWg.Done()
		client.deviceAuthWg = nil
	}
}

func NewClient(hostname string, port uint, authChallenge bool, wg *sync.WaitGroup) *Client {
	var log = NewLogger("client")

	addr := fmt.Sprintf("%s:%d", hostname, port)
	log.Info(fmt.Sprintf("addr: %s", addr))

	config := tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		log.Error("client: dial error", "err", err)
		return nil
	}

	log.Info("Connected")

	castChannel := CreateCastChannel(conn, log)

	client := Client{
		castChannel: castChannel,
		conn:        conn,
		log:         log,
		//Incoming:    make(chan *channel.CastMessage),
	}

	if authChallenge {
		var deviceAuthWg sync.WaitGroup
		client.sendDeviceAuthChallenge(&deviceAuthWg)
	}

	go func() {
		for {
			select {
			case castMessage, ok := <-castChannel.Messages:
				if ok {
					if castMessage != nil {
						if log.IsDebug() {
							log.Debug("received message", "content", castMessage)
						} else {
							log.Info("received message", "namespace", *castMessage.Namespace)
						}
					}

					//client.Incoming <- castMessage

					if *castMessage.Namespace == deviceAuthNamespace {
						client.verifyDeviceAuthResponse(castMessage.PayloadBinary)
					}
				} else {
					log.Info("channel closed")
					_ = conn.Close()
					if wg != nil {
						wg.Done()
					}
					return
				}
			}
		}
	}()

	return &client
}

func (client *Client) SendMessage(castMessage *channel.CastMessage) {
	client.castChannel.Send(castMessage)
}
