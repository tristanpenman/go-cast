package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	// third-party
	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/common"
)

type Server struct {
	clientConnections []*ClientConnection
	listener          net.Listener
	interfaceNames    []string
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
	listenHost, err := resolveListenHost(iface)
	if err != nil {
		return nil, fmt.Errorf("resolve Cast listener interface: %w", err)
	}
	addr := net.JoinHostPort(listenHost, strconv.Itoa(port))
	listener, err := tls.Listen("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("listen for Cast connections: %w", err)
	}
	interfaceNames, err := listenerInterfaceNames(listener.Addr())
	if err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("resolve Cast listener network interface: %w", err)
	}

	log.Info("listening", "addr", listener.Addr(), "interfaces", interfaceNames)

	server := Server{
		clientConnections: make([]*ClientConnection, 0),
		listener:          listener,
		interfaceNames:    interfaceNames,
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

// InterfaceNames returns the network interfaces on which the listener accepts
// Cast connections. An empty result means the listener is bound to all local
// interfaces.
func (server *Server) InterfaceNames() []string {
	return append([]string(nil), server.interfaceNames...)
}

// StopListening stops the server from accepting new connections.
func (server *Server) StopListening() error {
	if err := server.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("stop listening: %w", err)
	}
	return nil
}

func resolveListenHost(interfaceValue *string) (string, error) {
	if interfaceValue == nil || strings.TrimSpace(*interfaceValue) == "" {
		return "", nil
	}
	value := strings.TrimSpace(*interfaceValue)
	if ip := net.ParseIP(value); ip != nil {
		return ip.String(), nil
	}

	iface, err := net.InterfaceByName(value)
	if err != nil {
		// Preserve support for hostnames while allowing --iface to use the
		// interface-name behavior described by the flag.
		return value, nil
	}
	ip, err := preferredInterfaceIP(iface)
	if err != nil {
		return "", err
	}
	if ip.To4() == nil && ip.IsLinkLocalUnicast() {
		return ip.String() + "%" + iface.Name, nil
	}
	return ip.String(), nil
}

func preferredInterfaceIP(iface *net.Interface) (net.IP, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("list addresses for interface %s: %w", iface.Name, err)
	}
	var ipv6 net.IP
	for _, addr := range addrs {
		ip := ipFromAddr(addr)
		if ip == nil {
			continue
		}
		if ip.To4() != nil {
			return ip, nil
		}
		if ipv6 == nil {
			ipv6 = ip
		}
	}
	if ipv6 != nil {
		return ipv6, nil
	}
	return nil, fmt.Errorf("interface %s has no IP address", iface.Name)
}

func listenerInterfaceNames(addr net.Addr) ([]string, error) {
	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("unexpected listener address type %T", addr)
	}
	if tcpAddr.IP == nil || tcpAddr.IP.IsUnspecified() {
		return nil, nil
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list network interfaces: %w", err)
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, candidate := range addrs {
			if ip := ipFromAddr(candidate); ip != nil && ip.Equal(tcpAddr.IP) {
				return []string{iface.Name}, nil
			}
		}
	}
	return nil, fmt.Errorf("no network interface owns listener address %s", tcpAddr.IP)
}

func ipFromAddr(addr net.Addr) net.IP {
	switch value := addr.(type) {
	case *net.IPNet:
		return value.IP
	case *net.IPAddr:
		return value.IP
	default:
		return nil
	}
}
