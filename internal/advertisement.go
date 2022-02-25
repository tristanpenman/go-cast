package internal

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/mdns"
)

type Advertisement struct {
	log    hclog.Logger
	server *mdns.Server
}

func NewAdvertisement(hostname *string, port int) *Advertisement {
	var log = NewLogger("advertisement")

	log.Info("starting mdns...")
	if hostname == nil {
		*hostname, _ = os.Hostname()
	}

	info := []string{"test"}
	service, err := mdns.NewMDNSService("go-cast", "_googlecast._tcp", "", *hostname, port, nil, info)
	if err != nil {
		log.Error("failed to create mdns service", "err", err)
		return nil
	}

	var server *mdns.Server
	server, err = mdns.NewServer(&mdns.Config{
		Zone: service,
	})

	if err != nil {
		log.Error("failed to create mdns server", "err", err)
		return nil
	}

	log.Info("started")

	return &Advertisement{
		log:    log,
		server: server,
	}
}
