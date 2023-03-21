package internal

import "net"

func GetPort(addr net.Addr) int {
	switch addr.(type) {
	case *net.UDPAddr:
		return addr.(*net.UDPAddr).Port
	case *net.TCPAddr:
		return addr.(*net.TCPAddr).Port
	}

	return -1
}
