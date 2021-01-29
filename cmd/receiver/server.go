package main

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"

	"github.com/tristanpenman/go-cast/internal/message"
)

func handleClient(conn net.Conn) {
	defer conn.Close()

	lenBytes := make([]byte, 4)
	msgBytes := make([]byte, 512)

	for {
		logger.Info("server: connection waiting")

		n, err := conn.Read(lenBytes)
		if err != nil {
			logger.Error("server: connection read error", "err", err)
			break
		}

		if n != 4 {
			logger.Error("Failed to read length; too short")
			break
		}

		if logger.IsDebug() {
			lenInt := binary.BigEndian.Uint32(lenBytes)
			logger.Debug(fmt.Sprintf("Message length: %d", lenInt))
		}

		// TODO: Make this handle split header and body packets properly
		n, err = conn.Read(msgBytes)
		if err != nil {
			logger.Error("server: connection read error", "err", err)
			break
		}

		var castMessage message.CastMessage
		err = proto.Unmarshal(msgBytes[:n], &castMessage)
		if err != nil {
			logger.Warn("Failed to parse cast message", "err", err)
			continue
		}

		logger.Info("Received message", "namespace", *castMessage.Namespace)
	}

	logger.Info("server: connection closed")
}

func startServer(manifest map[string]string, iface string, port uint) {
	cert, err := tls.X509KeyPair([]byte(manifest["pu"]), []byte(manifest["pr"]))
	if err != nil {
		logger.Error("Failed to load X509 keypair", "err", err)
		return
	}

	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	addr := fmt.Sprintf("%s:%d", iface, port)
	listener, err := tls.Listen("tcp", addr, cfg)
	if err != nil {
		logger.Error("Failed to listen", "err", err)
		return
	}

	logger.Info("listening...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Info("server: accept failed", "err", err)
			break
		}

		logger.Info("server: accepted connection", "remote addr", conn.RemoteAddr())

		go handleClient(conn)
	}
}
