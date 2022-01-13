package main

import (
	"flag"
)

func main() {
	var disableChallenge = flag.Bool("disable-challenge", false, "disable auth challenge")
	var hostname = flag.String("hostname", "", "receiver address or hostname")
	var port = flag.Uint("port", 8009, "receiver port")

	flag.Parse()

	if *hostname == "" {
		flag.PrintDefaults()
		return
	}

	logger.Info("args",
		"disable-challenge", disableChallenge,
		"hostname", *hostname,
		"port", port)

	startClient(hostname, port, !*disableChallenge)
}
