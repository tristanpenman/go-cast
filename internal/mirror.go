package internal

import (
	"net"
)

type Mirror struct {
	packetConn net.PacketConn
}

func (mirror *Mirror) GetPort() int {
	return GetPort(mirror.packetConn.LocalAddr())
}

func NewMirror() *Mirror {
	packetConn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil
	}

	return &Mirror{
		packetConn: packetConn,
	}
}
