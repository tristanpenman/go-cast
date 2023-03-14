package internal

import "net"

func ExtractMiddleBits(num uint64) uint32 {
	// Shift the 64-bit integer to the right by 16 bits to discard the lower 16 bits
	// and move the desired bits to the lower 32 bits of the result.
	// Then use a bitwise AND operation with 0xFFFFFFFF to mask off the upper 32 bits
	// and obtain only the middle 32 bits.
	return uint32((num >> 16) & 0xFFFFFFFF)
}

func GetPort(addr net.Addr) int {
	switch addr.(type) {
	case *net.UDPAddr:
		return addr.(*net.UDPAddr).Port
	case *net.TCPAddr:
		return addr.(*net.TCPAddr).Port
	}

	return -1
}
