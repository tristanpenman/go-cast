package discovery

import (
	"net"
	"testing"

	"github.com/hashicorp/mdns"
)

func TestDeviceFromEntry(t *testing.T) {
	entry := &mdns.ServiceEntry{
		Name:       "Living Room._googlecast._tcp.local.",
		Host:       "chromecast.local.",
		AddrV4:     net.ParseIP("192.0.2.10"),
		Port:       8009,
		InfoFields: []string{"id=device-1", "fn=Living Room", "md=Chromecast Ultra"},
	}

	got := deviceFromEntry(entry)
	if got.ID != "device-1" || got.Name != "Living Room" || got.Model != "Chromecast Ultra" {
		t.Fatalf("unexpected metadata: %+v", got)
	}
	if got.Host != "192.0.2.10" || got.Port != 8009 {
		t.Fatalf("unexpected address: %+v", got)
	}
	if got.ConnectionState != "discovered" {
		t.Fatalf("unexpected connection state: %q", got.ConnectionState)
	}
}

func TestDeviceFromEntryFallsBackToHostAndServiceName(t *testing.T) {
	entry := &mdns.ServiceEntry{
		Name: "Office TV._googlecast._tcp.local.",
		Host: "office-tv.local.",
		Port: 8009,
	}

	got := deviceFromEntry(entry)
	if got.Name != "Office TV" {
		t.Fatalf("unexpected name: %q", got.Name)
	}
	if got.Host != "office-tv.local" {
		t.Fatalf("unexpected host: %q", got.Host)
	}
}

func TestDiscoverRejectsInvalidTimeout(t *testing.T) {
	if _, err := Discover(0); err == nil {
		t.Fatal("expected an error")
	}
}
