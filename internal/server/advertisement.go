package server

import (
	"context"
	"fmt"

	"github.com/brutella/dnssd"
	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/common"
)

type Advertisement struct {
	cancel    context.CancelFunc
	log       hclog.Logger
	responder dnssd.Responder
}

func (advertisement *Advertisement) Stop() {
	advertisement.cancel()
	advertisement.log.Info("stopped")
}

// NewAdvertisement starts advertising a Cast device over mDNS.
func NewAdvertisement(device *Device, port int) (*Advertisement, error) {
	var log = common.NewLogger("advertisement")

	log.Info("starting mdns...")

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

	cfg := dnssd.Config{
		Name:   "GoCast",
		Type:   "_googlecast._tcp",
		Domain: "local",
		Host:   "",
		Port:   port,
		Text:   info,
	}

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

	log.Info("starting")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		err := responder.Respond(ctx)
		if err != nil {
			log.Error("failed to start responder", "err", err)
		}
	}()

	return &Advertisement{
		cancel:    cancel,
		log:       log,
		responder: responder,
	}, nil
}
