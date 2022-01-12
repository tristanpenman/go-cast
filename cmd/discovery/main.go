package main

import (
	"fmt"
	"time"

	"github.com/hashicorp/mdns"
)

func main() {
	entries := make(chan *mdns.ServiceEntry, 4)
	defer close(entries)

	go func() {
		for entry := range entries {
			fmt.Printf("got new entry: %s %s %d %v\n", entry.Name, entry.AddrV4, entry.Port, entry.InfoFields)
		}
	}()

	params := mdns.DefaultParams("_googlecast._tcp")
	params.DisableIPv6 = true
	params.Entries = entries
	params.Timeout = 10 * time.Second

	err := mdns.Query(params)
	if err != nil {
		fmt.Printf("error performing mdns lookup: %s", err)
	}
}
