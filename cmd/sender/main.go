package main

import (
	"flag"
)

func main() {
	var hostname = flag.String("hostname", "localhost", "receiver address or hostname")
	var port = flag.Uint("port", 8009, "receiver port")

	startClient(hostname, port)
}
