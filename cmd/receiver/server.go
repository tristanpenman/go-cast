package main

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"

	"github.com/tristanpenman/go-cast/internal/cast"
)

func sendDeviceAuthResponse(castChannel *cast.CastChannel) bool {
	payloadBytes, err := proto.Marshal(&cast.DeviceAuthMessage{
		Response: &cast.AuthResponse{},
	})

	if err != nil {
		logger.Error("failed to encode device auth response", "err", err)
		return false
	}

	payloadType := cast.CastMessage_BINARY
	return castChannel.Send(&cast.CastMessage{
		Namespace: &cast.DeviceAuthNamespace,
		PayloadBinary: payloadBytes,
		PayloadType: &payloadType,
	})
}

func handleClient(conn net.Conn) {
	defer func() {
		_ = conn.Close()
		logger.Info("connection closed")
	}()

	castChannel := cast.CreateCastChannel(conn, logger)

	for {
		select {
		case castMessage, ok := <-castChannel.Messages:
			if castMessage != nil {
				logger.Info("received", "message", castMessage)
				if *castMessage.Namespace == cast.DeviceAuthNamespace {
					sendDeviceAuthResponse(&castChannel)
				}
			}
			if !ok {
				logger.Info("channel closed")
				return
			}
		}
	}
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
