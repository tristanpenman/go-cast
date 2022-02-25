package main

import (
	"flag"
	"sync"

	. "github.com/tristanpenman/go-cast/internal"
)

var log = NewLogger("main")

func main() {
	var disableChallenge = flag.Bool("disable-challenge", false, "disable auth challenge")
	var hostname = flag.String("hostname", "", "receiver address or hostname")
	var port = flag.Uint("port", 8009, "receiver port")

	flag.Parse()

	if *hostname == "" {
		flag.PrintDefaults()
		return
	}

	log.Info("args",
		"disable-challenge", *disableChallenge,
		"hostname", *hostname,
		"port", *port)

	var wg sync.WaitGroup
	wg.Add(1)

	client := NewClient(*hostname, *port, !*disableChallenge, &wg)
	if client == nil {
		return
	}

	wg.Wait()
}
