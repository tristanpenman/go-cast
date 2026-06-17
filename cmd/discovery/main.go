package main

import (
	"flag"
	"log/slog"
	"time"

	"github.com/tristanpenman/go-cast/internal/discovery"
)

var log = slog.Default()

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "how long to search for Cast devices")
	flag.Parse()

	devices, err := discovery.Discover(*timeout)
	if err != nil {
		log.Error("error performing mDNS lookup", "err", err)
		return
	}
	for _, device := range devices {
		log.Info("found device",
			"id", device.ID,
			"name", device.Name,
			"model", device.Model,
			"host", device.Host,
			"port", device.Port)
	}
}
