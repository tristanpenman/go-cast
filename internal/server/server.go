package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"

	// third-party
	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/common"
)

type Server struct {
	clientConnections []*ClientConnection
	listener          net.Listener
	log               hclog.Logger
	nextClientId      int
}

// NewServer starts a TLS listener for Cast client connections.
func NewServer(
	device *Device,
	manifest map[string]string,
	clientPrefix *string,
	iface *string,
	port int,
) (*Server, error) {
	var log = common.NewLogger("server")

	cert, err := tls.X509KeyPair([]byte(manifest["pu"]), []byte(manifest["pr"]))
	if err != nil {
		return nil, fmt.Errorf("load X509 keypair: %w", err)
	}

	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	addr := fmt.Sprintf("%s:%d", *iface, port)
	listener, err := tls.Listen("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("listen for Cast connections: %w", err)
	}

	log.Info("listening")

	server := Server{
		clientConnections: make([]*ClientConnection, 0),
		listener:          listener,
		log:               log,
		nextClientId:      0,
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					break
				}
				log.Error("server accept failed", "err", err)
				continue
			}

			if clientPrefix == nil || strings.HasPrefix(conn.RemoteAddr().String(), *clientPrefix) {
				log.Info("accepted connection", "remote addr", conn.RemoteAddr())
				id := server.nextClientId
				clientConnection := NewClientConnection(device, conn, id, manifest)
				server.nextClientId++
				server.clientConnections = append(server.clientConnections, clientConnection)
			} else {
				log.Debug("ignored connection", "remote addr", conn.RemoteAddr())
				_ = conn.Close()
			}
		}
	}()

	return &server, nil
}

// StopListening stops the server from accepting new connections.
func (server *Server) StopListening() error {
	if err := server.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("stop listening: %w", err)
	}
	return nil
}
