package internal

import (
	"context"

	// third-party
	"github.com/brutella/dnssd"
	"github.com/hashicorp/go-hclog"
)

type Advertisement struct {
	cancel    context.CancelFunc
	device    Device
	log       hclog.Logger
	responder dnssd.Responder
}

func (advertisement *Advertisement) Stop() {
	advertisement.cancel()
	advertisement.log.Info("stopped")
}

func NewAdvertisement(device *Device, port int) *Advertisement {
	var log = NewLogger("advertisement")

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
		log.Error("failed to create service", "err", err)
		return nil
	}

	responder, err := dnssd.NewResponder()
	if err != nil {
		log.Error("failed to create responder", "err", err)
		return nil
	}

	_, err = responder.Add(service)
	if err != nil {
		log.Error("failed to add service to responder", "err", err)
		return nil
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
	}
}
