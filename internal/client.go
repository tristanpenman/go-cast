package internal

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/tristanpenman/go-cast/internal/cast"
)

type Client struct {
	conn        net.Conn
	castChannel cast.CastChannel
}

func sendDeviceAuthChallenge(castChannel *cast.CastChannel) bool {
	payloadBytes, err := proto.Marshal(&cast.DeviceAuthMessage{
		Challenge: &cast.AuthChallenge{},
	})

	if err != nil {
		Logger.Error("failed to encode device auth challenge", "err", err)
		return false
	}

	payloadType := cast.CastMessage_BINARY
	protocolVersion := cast.CastMessage_CASTV2_1_0
	sourceId := "sender-0"
	destinationId := "receiver-0"
	return castChannel.Send(&cast.CastMessage{
		DestinationId:   &destinationId,
		Namespace:       &cast.DeviceAuthNamespace,
		PayloadBinary:   payloadBytes,
		PayloadType:     &payloadType,
		ProtocolVersion: &protocolVersion,
		SourceId:        &sourceId,
	})
}

func NewClient(hostname *string, port *uint, authChallenge bool, wg *sync.WaitGroup) *Client {
	addr := fmt.Sprintf("%s:%d", *hostname, *port)
	Logger.Info(fmt.Sprintf("addr: %s", addr))

	config := tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		Logger.Error("client: dial error", "err", err)
		return nil
	}

	castChannel := cast.CreateCastChannel(conn, Logger)

	if authChallenge {
		sendDeviceAuthChallenge(&castChannel)
	}

	go func() {
		for {
			select {
			case castMessage, ok := <-castChannel.Messages:
				if castMessage != nil {
					if Logger.IsDebug() {
						Logger.Debug("received message", "content", castMessage)
					} else {
						Logger.Info("received message", "namespace", *castMessage.Namespace)
					}
				}
				if !ok {
					Logger.Info("channel closed")
					_ = conn.Close()
					wg.Done()
					return
				}
			}
		}
	}()

	return &Client{
		castChannel: castChannel,
		conn:        conn,
	}
}

func (client *Client) sendMessage(castMessage *cast.CastMessage) {
	client.castChannel.Send(castMessage)
}
