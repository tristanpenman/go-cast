package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/google/uuid"

	. "github.com/tristanpenman/go-cast/internal"
)

var log = NewLogger("main")

func loadManifest(certManifest string, fixNewlines bool) map[string]string {
	data, err := ioutil.ReadFile(certManifest)
	if err != nil {
		log.Error("failed to read certificate manifest file from disk", "err", err)
		return nil
	}

	// convert new-line characters so that JSON parses correctly
	var s = string(data)
	if fixNewlines {
		s = strings.ReplaceAll(s, "\n", "\\n")
	}

	var manifest map[string]string
	err = json.Unmarshal([]byte(s), &manifest)
	if err != nil {
		log.Error("failed to parse certificate manifest file", "err", err)
		return nil
	}

	if log.IsDebug() {
		log.Debug("manifest contents")
		for key, value := range manifest {
			log.Debug(fmt.Sprintf("%s: %s", key, value))
		}
	}

	return manifest
}

func main() {
	var certManifest = flag.String("cert-manifest", "", "path to certificate manifest")
	var clientPrefix = flag.String("client-prefix", "", "optional client prefix, to limit connections")
	var deviceModel = flag.String("device-model", "go-cast", "device model")
	var enableMdns = flag.Bool("enable-mdns", false, "advertise service using mDNS")
	var fixNewlines = flag.Bool("fix-newlines", false, "fix newline characters in manifest file")
	var friendlyName = flag.String("friendly-name", "GoCast Receiver", "friendly name")
	var hostname = flag.String("hostname", "", "override default OS hostname (optional)")
	var iface = flag.String("iface", "", "interface to listen on (optional)")
	var port = flag.Int("port", 8009, "port to listen on")
	var relayAuthChallenge = flag.Bool("relay-auth-challenge", false, "send auth challenge when relaying")
	var relayHost = flag.String("relay-host", "", "relay to another Chromecast receiver (optional)")
	var relayPort = flag.Uint("relay-port", 8009, "port to relay to (optional)")

	flag.Parse()

	if *certManifest == "" {
		flag.PrintDefaults()
		return
	}

	log.Info("args",
		"cert-manifest", *certManifest,
		"client-prefix", *clientPrefix,
		"device-model", *deviceModel,
		"enable-mdns", *enableMdns,
		"fix-newlines", *fixNewlines,
		"friendly-name", *friendlyName,
		"hostname", *hostname,
		"iface", *iface,
		"port", *port,
		"relay-auth-challenge", *relayAuthChallenge,
		"relay-host", *relayHost,
		"relay-port", *relayPort,
	)

	manifest := loadManifest(*certManifest, *fixNewlines)

	var wg sync.WaitGroup
	wg.Add(1)

	id := uuid.New().String()
	udn := id
	device := NewDevice(*deviceModel, *friendlyName, id, udn)

	server := NewServer(device, manifest, clientPrefix, iface, hostname, *port, *relayHost, *relayPort, *relayAuthChallenge, &wg)
	if server == nil {
		return
	}

	if *enableMdns {
		advertisement := NewAdvertisement(device, hostname, *port)
		if advertisement == nil {
			log.Error("failed to advertise receiver")
		}
	}

	wg.Wait()
}
