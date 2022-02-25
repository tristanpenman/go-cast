package internal

import (
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"net"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/tristanpenman/go-cast/internal/cast"
)

type Client struct {
	conn        net.Conn
	castChannel cast.CastChannel
	log         hclog.Logger
}

func (client *Client) sendDeviceAuthChallenge() bool {
	payloadBytes, err := proto.Marshal(&cast.DeviceAuthMessage{
		Challenge: &cast.AuthChallenge{},
	})

	if err != nil {
		client.log.Error("failed to encode device auth challenge", "err", err)
		return false
	}

	payloadType := cast.CastMessage_BINARY
	protocolVersion := cast.CastMessage_CASTV2_1_0
	sourceId := "sender-0"
	destinationId := "receiver-0"
	message := cast.CastMessage{
		DestinationId:   &destinationId,
		Namespace:       &cast.DeviceAuthNamespace,
		PayloadBinary:   payloadBytes,
		PayloadType:     &payloadType,
		ProtocolVersion: &protocolVersion,
		SourceId:        &sourceId,
	}

	return client.castChannel.Send(&message)
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

	castChannel := cast.CreateCastChannel(conn, log)

	client := Client{
		castChannel: castChannel,
		conn:        conn,
		log:         log,
	}

	if authChallenge {
		client.sendDeviceAuthChallenge()
	}

	go func() {
		for {
			select {
			case castMessage, ok := <-castChannel.Messages:
				if castMessage != nil {
					if log.IsDebug() {
						log.Debug("received message", "content", castMessage)
					} else {
						log.Info("received message", "namespace", *castMessage.Namespace)
					}
				}
				if !ok {
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

func (client *Client) SendMessage(castMessage *cast.CastMessage) {
	client.castChannel.Send(castMessage)
}
