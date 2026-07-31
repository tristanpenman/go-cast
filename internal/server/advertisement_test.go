package server

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
)

func TestAdvertisementStopWaitsForResponder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	advertisement := &Advertisement{
		cancel: cancel,
		done:   done,
		log:    hclog.NewNullLogger(),
	}

	responderFinished := make(chan struct{})
	go func() {
		<-ctx.Done()
		time.Sleep(25 * time.Millisecond)
		close(responderFinished)
		close(done)
	}()

	advertisement.Stop()
	select {
	case <-responderFinished:
	default:
		t.Fatal("Stop returned before the responder finished")
	}

	// Stop must remain safe if cleanup is requested more than once.
	advertisement.Stop()
}

func TestAdvertisementConfigUsesListenerInterfaces(t *testing.T) {
	device := &Device{DeviceModel: "go-cast", FriendlyName: "Living Room", Id: "device-id"}
	interfaces := []string{"en0"}
	config := advertisementConfig(device, 8009, interfaces)

	if len(config.Ifaces) != 1 || config.Ifaces[0] != "en0" {
		t.Fatalf("unexpected advertisement interfaces: %v", config.Ifaces)
	}

	interfaces[0] = "utun0"
	if config.Ifaces[0] != "en0" {
		t.Fatal("advertisement config retained caller's mutable interface slice")
	}
}
