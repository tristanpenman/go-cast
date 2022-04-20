package internal

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/channel"
)

type Client struct {
	castChannel CastChannel
	conn        net.Conn
	log         hclog.Logger
	Incoming    chan *channel.CastMessage
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
		Incoming:    make(chan *channel.CastMessage),
	}

	if authChallenge {
		client.sendDeviceAuthChallenge()
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

					client.Incoming <- castMessage
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
