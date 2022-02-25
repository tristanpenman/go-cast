package main

import (
	"time"

	"github.com/hashicorp/mdns"

	. "github.com/tristanpenman/go-cast/internal"
)

var log = NewLogger("main")

func main() {
	entries := make(chan *mdns.ServiceEntry, 4)
	defer close(entries)

	go func() {
		for entry := range entries {
			log.Info("got new entry", "name", entry.Name, "addr", entry.AddrV4, "port", entry.Port, "info", entry.InfoFields)
		}
	}()

	params := mdns.DefaultParams("_googlecast._tcp")
	params.DisableIPv6 = true
	params.Entries = entries
	params.Timeout = 10 * time.Second

	err := mdns.Query(params)
	if err != nil {
		log.Error("error performing mdns lookup", "err", err)
	}
}
