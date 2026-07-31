package server

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/brutella/dnssd"
	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/common"
)

type Advertisement struct {
	cancel   context.CancelFunc
	done     chan struct{}
	log      hclog.Logger
	stopOnce sync.Once
}

func (advertisement *Advertisement) Stop() {
	advertisement.stopOnce.Do(func() {
		advertisement.cancel()
		<-advertisement.done
		advertisement.log.Info("stopped")
	})
}

// NewAdvertisement starts advertising a Cast device over mDNS.
func NewAdvertisement(device *Device, port int, interfaceNames []string) (*Advertisement, error) {
	var log = common.NewLogger("advertisement")

	log.Info("starting mdns...")

	cfg := advertisementConfig(device, port, interfaceNames)

	service, err := dnssd.NewService(cfg)
	if err != nil {
		return nil, fmt.Errorf("create discovery service: %w", err)
	}

	responder, err := dnssd.NewResponder()
	if err != nil {
		return nil, fmt.Errorf("create discovery responder: %w", err)
	}

	_, err = responder.Add(service)
	if err != nil {
		return nil, fmt.Errorf("add discovery service: %w", err)
	}

	log.Info("starting", "interfaces", interfaceNames)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer cancel()
		defer close(done)
		err := responder.Respond(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Error("failed to start responder", "err", err)
		}
	}()

	return &Advertisement{
		cancel: cancel,
		done:   done,
		log:    log,
	}, nil
}

func advertisementConfig(device *Device, port int, interfaceNames []string) dnssd.Config {
	info := map[string]string{
		"ve": "02",
		"st": "0",
		"nf": "1",
		"ca": "4101",
		"ic": "/setup/icon.png",
		"md": device.DeviceModel,
		"id": device.Id,
		"fn": device.FriendlyName,
	}

	return dnssd.Config{
		Name:   "GoCast",
		Type:   "_googlecast._tcp",
		Domain: "local",
		Host:   "",
		Port:   port,
		Text:   info,
		Ifaces: append([]string(nil), interfaceNames...),
	}
}
