package server

import (
	"net"
	"testing"
)

func TestListenerInterfaceNamesWildcardMeansAllInterfaces(t *testing.T) {
	got, err := listenerInterfaceNames(&net.TCPAddr{IP: net.IPv4zero, Port: 8009})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("wildcard listener returned interfaces: %v", got)
	}
}

func TestServerInterfaceNamesReturnsCopy(t *testing.T) {
	server := &Server{interfaceNames: []string{"en0"}}
	got := server.InterfaceNames()
	got[0] = "utun0"
	if server.interfaceNames[0] != "en0" {
		t.Fatal("InterfaceNames exposed mutable server state")
	}
}

func TestResolveListenHostAcceptsIPAddress(t *testing.T) {
	value := "192.0.2.10"
	got, err := resolveListenHost(&value)
	if err != nil {
		t.Fatal(err)
	}
	if got != value {
		t.Fatalf("listen host %q, want %q", got, value)
	}
}
