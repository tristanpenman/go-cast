package internal

import (
	"context"
	"os"

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

func NewAdvertisement(device *Device, hostname *string, port int) *Advertisement {
	var log = NewLogger("advertisement")

	log.Info("starting mdns...")
	if hostname == nil {
		*hostname, _ = os.Hostname()
	}

	info := map[string]string{
		"cd": "",
		"rm": "",
		"ve": "02",
		"st": "0",
		"rs": "",
		"nf": "1",
		"ic": "",
		"ca": "4101",
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		responder.Respond(ctx)
	}()

	log.Info("started")

	return &Advertisement{
		cancel:    cancel,
		log:       log,
		responder: responder,
	}
}
