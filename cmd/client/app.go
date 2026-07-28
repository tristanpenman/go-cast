package main

import (
	"time"

	"github.com/tristanpenman/go-cast/internal/discovery"
)

// App contains the backend methods exposed to the Wails frontend.
type App struct {
	discover func(time.Duration) ([]discovery.Device, error)
}

func NewApp() *App {
	return &App{discover: discovery.Discover}
}

// DiscoverDevices performs a bounded mDNS scan for Google Cast receivers.
func (a *App) DiscoverDevices() ([]discovery.Device, error) {
	return a.discover(5 * time.Second)
}
