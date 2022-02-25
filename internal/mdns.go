package internal

import (
	"github.com/hashicorp/mdns"
	"os"
)

func StartAdvertisement(hostname *string, port int) {
	var log = NewLogger("mdns")

	log.Info("starting mdns...")
	if hostname == nil {
		*hostname, _ = os.Hostname()
	}
	info := []string{"test"}

	// TODO: Error handling
	service, err := mdns.NewMDNSService("go-cast", "_googlecast._tcp", "", *hostname, port, nil, info)
	if err != nil {
		log.Warn("failed to create mdns service", "err", err)
		return
	}

	_, err = mdns.NewServer(&mdns.Config{
		Zone: service,
	})

	if err != nil {
		log.Warn("failed to create mdns server", "err", err)
		return
	}

	log.Info("started")
}
