package main

import (
	"errors"
	"testing"
	"time"

	castclient "github.com/tristanpenman/go-cast/internal/client"
	"github.com/tristanpenman/go-cast/internal/discovery"
)

func TestDiscoverDevices(t *testing.T) {
	want := []discovery.Device{{ID: "one", Name: "Living Room"}}
	app := &App{discover: func(timeout time.Duration) ([]discovery.Device, error) {
		if timeout != 5*time.Second {
			t.Fatalf("unexpected timeout: %s", timeout)
		}
		return want, nil
	}}

	got, err := app.DiscoverDevices()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != want[0].ID {
		t.Fatalf("unexpected devices: %+v", got)
	}
}

func TestDiscoverDevicesReturnsDiscoveryError(t *testing.T) {
	app := &App{discover: func(time.Duration) ([]discovery.Device, error) {
		return nil, errors.New("network unavailable")
	}}
	if _, err := app.DiscoverDevices(); err == nil {
		t.Fatal("expected an error")
	}
}

func TestDeviceAppsMergesAvailabilityAndRunningStatus(t *testing.T) {
	availability := map[string]string{
		"233637DE": "APP_AVAILABLE",
		"0F5096E8": "APP_UNAVAILABLE",
		"674A0243": "APP_AVAILABLE",
	}
	status := &castclient.ReceiverStatus{Applications: []castclient.Application{
		{AppID: "233637DE", DisplayName: "YouTube", SessionID: "session-1", StatusText: "Playing"},
		{AppID: "CUSTOM", DisplayName: "Custom receiver", SessionID: "session-2"},
	}}

	got := deviceApps(availability, status)
	if len(got) != 3 {
		t.Fatalf("expected three apps, got %+v", got)
	}
	if got[0].ID != "233637DE" || !got[0].Running || got[0].StatusText != "Playing" {
		t.Fatalf("unexpected YouTube app: %+v", got[0])
	}
	if got[1].ID != "674A0243" || got[1].Running {
		t.Fatalf("unexpected Android mirroring app: %+v", got[1])
	}
	if got[2].ID != "CUSTOM" || !got[2].Running || got[2].Name != "Custom receiver" {
		t.Fatalf("unexpected custom app: %+v", got[2])
	}
}
