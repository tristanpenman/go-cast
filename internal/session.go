package internal

import (
	"fmt"
	"github.com/tristanpenman/go-cast/internal/cast"
	"net"

	"github.com/hashicorp/go-hclog"
)

type Session struct {
	AppId       string
	DisplayName string
	SessionId   string
	StatusText  string

	// implementation
	log         hclog.Logger
	packetConn  net.PacketConn
	transportId string
}

func (session *Session) GetPort() int {
	return GetPort(session.packetConn.LocalAddr())
}

func (session *Session) HandleCastMessage(message *cast.CastMessage) {
	//TODO implement me
	panic("implement me")
}

func (session *Session) Namespaces() []string {
	namespaces := make([]string, 4)

	namespaces[0] = debugNamespace
	namespaces[1] = mediaNamespace
	namespaces[2] = remotingNamespace
	namespaces[3] = webrtcNamespace

	return namespaces
}

func (session *Session) TransportId() string {
	return session.transportId
}

func NewSession(appId string, displayName string, sessionId string, transportId string) *Session {
	log := NewLogger(fmt.Sprintf("session (%s)", sessionId))

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
		AppId:       appId,
		DisplayName: displayName,
		SessionId:   sessionId,
		StatusText:  "",

		// implementation
		log:         log,
		packetConn:  packetConn,
		transportId: transportId,
	}
}
