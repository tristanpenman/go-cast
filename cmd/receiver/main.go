package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	. "github.com/tristanpenman/go-cast/internal"
)

func main() {
	var certManifest = flag.String("cert-manifest", "", "path to certificate manifest")
	var clientPrefix = flag.String("client-prefix", "", "optional client prefix, to limit connections")
	var enableMdns = flag.Bool("enable-mdns", false, "advertise service using mDNS")
	var fixNewlines = flag.Bool("fix-newlines", false, "fix newline characters in manifest file")
	var hostname = flag.String("hostname", "", "override default OS hostname (optional)")
	var iface = flag.String("iface", "", "interface to listen on (optional)")
	var port = flag.Int("port", 8009, "port to listen on")
	var relayHost = flag.String("relay-host", "", "relay to another Chromecast receiver (optional)")
	var relayPort = flag.Int("relay-port", 8009, "port to relay to (optional)")

	flag.Parse()

	if *certManifest == "" {
		flag.PrintDefaults()
		return
	}

	Logger.Info("args",
		"cert-manifest", *certManifest,
		"client-prefix", *clientPrefix,
		"enable-mdns", *enableMdns,
		"fix-newlines", *fixNewlines,
		"hostname", *hostname,
		"iface", *iface,
		"port", *port)

	// read manifest from disk
	data, err := ioutil.ReadFile(*certManifest)
	if err != nil {
		Logger.Error("failed to read certificate manifest file from disk", "err", err)
		return
	}

	// convert new-line characters so that JSON parses correctly
	var s = string(data)
	if *fixNewlines {
		s = strings.ReplaceAll(s, "\n", "\\n")
	}

	// parse manifest
	var manifest map[string]string
	err = json.Unmarshal([]byte(s), &manifest)
	if err != nil {
		Logger.Error("failed to parse certificate manifest file", "err", err)
		return
	}

	if Logger.IsDebug() {
		Logger.Debug("manifest contents")
		for key, value := range manifest {
			Logger.Debug(fmt.Sprintf("%s: %s", key, value))
		}
	}

	StartServer(manifest, clientPrefix, *enableMdns, iface, hostname, *port, relayHost, *relayPort)
}
