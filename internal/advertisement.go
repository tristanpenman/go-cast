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

	info := []string{
		"cd=",
		"rm=",
		"ve=2",
		"st=0",
		"rs=",
		"nf=1",
		"md=go-cast",
		"id=a98a257b4a3dd84392a34bd0",
		"ic=/setup/icon.png",
		"fn=go-cast",
		"ca=4101",
	}

	service, err := mdns.NewMDNSService("test", "_googlecast._tcp", "", "", port, nil, info)
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
