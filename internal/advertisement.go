package internal

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/mdns"
)

type Advertisement struct {
	device Device
	log    hclog.Logger
	server *mdns.Server
}

func NewAdvertisement(device *Device, hostname *string, port int) *Advertisement {
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
		"ic=/setup/icon.png",
		"ca=4101",
		fmt.Sprintf("md=%s", device.DeviceModel),
		fmt.Sprintf("id=%s", device.Id),
		fmt.Sprintf("fn=%s", device.FriendlyName),
	}

	service, err := mdns.NewMDNSService(device.Id, "_googlecast._tcp", "", *hostname, port, nil, info)
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
