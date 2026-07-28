package main

import (
	"errors"
	"testing"
	"time"

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
