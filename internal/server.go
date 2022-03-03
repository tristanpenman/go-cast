package internal

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
)

type Server struct {
	clientConnections []*ClientConnection
	listener          net.Listener
	log               hclog.Logger
	nextClientId      int
	wg                *sync.WaitGroup
}

func NewServer(
	device Device,
	manifest map[string]string,
	clientPrefix *string,
	iface *string,
	hostname *string,
	port int,
	relayHost string,
	relayPort uint,
	relayAuthChallenge bool,
	wg *sync.WaitGroup,
) *Server {
	var log = NewLogger("server")

	cert, err := tls.X509KeyPair([]byte(manifest["pu"]), []byte(manifest["pr"]))
	if err != nil {
		log.Error("failed to load X509 keypair", "err", err)
		return nil
	}

	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	addr := fmt.Sprintf("%s:%d", *iface, port)
	listener, err := tls.Listen("tcp", addr, cfg)
	if err != nil {
		log.Error("failed to listen", "err", err)
		return nil
	}

	log.Info("listening")

	// if relaying, attempt to connect to target
	var relayClient *Client
	if relayHost != "" {
		relayClient = NewClient(relayHost, relayPort, relayAuthChallenge, nil)
		if relayClient == nil {
			log.Error("failed to connect to target")
			return nil
		}
	}

	server := Server{
		clientConnections: make([]*ClientConnection, 0),
		listener:          listener,
		log:               log,
		nextClientId:      0,
		wg:                wg,
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Error("server accept failed", "err", err)
				break
			}

			if clientPrefix == nil || strings.HasPrefix(conn.RemoteAddr().String(), *clientPrefix) {
				log.Info("accepted connection", "remote addr", conn.RemoteAddr())
				id := server.nextClientId
				clientConnection := NewClientConnection(device, conn, id, manifest, relayClient)
				server.nextClientId++
				server.clientConnections = append(server.clientConnections, clientConnection)
			} else {
				log.Debug("ignored connection", "remote addr", conn.RemoteAddr())
				_ = conn.Close()
			}
		}
	}()

	return &server
}

func (server *Server) StopListening() {
	err := server.listener.Close()
	if err != nil {
		server.log.Error("failed to stop listening", "err", err)
	}

	server.wg.Done()
}
