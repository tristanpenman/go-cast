package internal

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/mdns"

	"github.com/tristanpenman/go-cast/internal/cast"
)

func sendDeviceAuthResponse(castChannel *cast.CastChannel) bool {
	payloadBytes, err := proto.Marshal(&cast.DeviceAuthMessage{
		Response: &cast.AuthResponse{},
	})

	if err != nil {
		Logger.Error("failed to encode device auth response", "err", err)
		return false
	}

	payloadType := cast.CastMessage_BINARY
	return castChannel.Send(&cast.CastMessage{
		Namespace:     &cast.DeviceAuthNamespace,
		PayloadBinary: payloadBytes,
		PayloadType:   &payloadType,
	})
}

func relayCastMessage(conn net.Conn, castMessage *cast.CastMessage, relay *net.Conn) {
	Logger.Info("relay cast message")
}

func handleCastMessage(conn net.Conn, castMessage *cast.CastMessage) {
	Logger.Info("handle cast message")
}

func handleClient(clientConn net.Conn, relayConn *net.Conn) {
	defer func() {
		_ = clientConn.Close()
		Logger.Info("connection closed")
	}()

	castChannel := cast.CreateCastChannel(clientConn, Logger)

	for {
		select {
		case castMessage, ok := <-castChannel.Messages:
			if castMessage != nil {
				Logger.Info("received", "message", castMessage)
				if *castMessage.Namespace == cast.DeviceAuthNamespace {
					sendDeviceAuthResponse(&castChannel)
				} else if relayConn != nil {
					relayCastMessage(clientConn, castMessage, relayConn)
				} else {
					handleCastMessage(clientConn, castMessage)
				}
			}
			if !ok {
				Logger.Info("channel closed")
				return
			}
		}
	}
}

func startAdvertisement(hostname *string, port int) {
	Logger.Info("starting mdns...")
	if hostname == nil {
		*hostname, _ = os.Hostname()
	}
	info := []string{"test"}

	// TODO: Error handling
	service, err := mdns.NewMDNSService("go-cast", "_googlecast._tcp", "", *hostname, port, nil, info)
	if err != nil {
		Logger.Warn("failed to create mdns service", "err", err)
		return
	}

	_, err = mdns.NewServer(&mdns.Config{
		Zone: service,
	})

	if err != nil {
		Logger.Warn("failed to create mdns server", "err", err)
		return
	}

	Logger.Info("started")
}

func StartServer(
	manifest map[string]string,
	clientPrefix *string,
	enableMdns bool,
	iface *string,
	hostname *string,
	port int,
	relayHost *string,
	relayPort int,
) {
	cert, err := tls.X509KeyPair([]byte(manifest["pu"]), []byte(manifest["pr"]))
	if err != nil {
		Logger.Error("failed to load X509 keypair", "err", err)
		return
	}

	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	addr := fmt.Sprintf("%s:%d", *iface, port)
	listener, err := tls.Listen("tcp", addr, cfg)
	if err != nil {
		Logger.Error("failed to listen", "err", err)
		return
	}

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			Logger.Error("failed to stop listening", "err", err)
		}
	}(listener)

	Logger.Info("listening")

	if enableMdns {
		startAdvertisement(hostname, port)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			Logger.Error("server accept failed", "err", err)
			break
		}

		if clientPrefix == nil || strings.HasPrefix(conn.RemoteAddr().String(), *clientPrefix) {
			Logger.Info("accepted connection", "remote addr", conn.RemoteAddr())
			go handleClient(conn, nil)
		} else {
			Logger.Debug("ignored connection", "remote addr", conn.RemoteAddr())
			_ = conn.Close()
		}
	}
}
