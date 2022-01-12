package main

import (
	"flag"
)

func main() {
	var hostname = flag.String("hostname", "", "receiver address or hostname")
	var port = flag.Uint("port", 8009, "receiver port")
	var disableChallenge = flag.Bool("disable-challenge", false, "disable auth challenge")

	flag.Parse()

	if *hostname == "" {
		flag.PrintDefaults()
		return
	}

	startClient(hostname, port, !*disableChallenge)
}
