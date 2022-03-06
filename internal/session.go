package internal

import (
	"fmt"
	"net"

	"github.com/hashicorp/go-hclog"
)

type Session struct {
	id         string
	log        hclog.Logger
	packetConn net.PacketConn
}

func (session *Session) GetPort() int {
	return GetPort(session.packetConn.LocalAddr())
}

func NewSession(id string) *Session {
	log := NewLogger(fmt.Sprintf("session (%s)", id))

	packetConn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil
	}

	go func() {
		bytes := make([]byte, 1500)
		for {
			count, peer, err := packetConn.ReadFrom(bytes)
			if err != nil {
				log.Error(fmt.Sprintf("error while reading from socket: %d", err))
				packetConn.Close()
			}

			log.Info("read %d from %s", count, peer.Network())
		}
	}()

	return &Session{
		id:         id,
		log:        log,
		packetConn: packetConn,
	}
}
