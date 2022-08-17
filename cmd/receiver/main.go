package main

import (
	"flag"
	"github.com/google/uuid"
	"os"
	"os/signal"
	"sync"
	"syscall"

	. "github.com/tristanpenman/go-cast/internal"
)

var log = NewLogger("main")

func main() {
	var certManifest = flag.String("cert-manifest", "", "path to certificate manifest")
	var certService = flag.String("cert-service", "", "base URL for certificate service")
	var certServiceSalt = flag.String("cert-service-salt", "", "salt for generating cert service hash")
	var clientPrefix = flag.String("client-prefix", "", "optional client prefix, to limit connections")
	var deviceModel = flag.String("device-model", "go-cast", "device model")
	var enableMdns = flag.Bool("enable-mdns", false, "advertise service using mDNS")
	var fixNewlines = flag.Bool("fix-newlines", false, "fix newline characters in manifest file")
	var friendlyName = flag.String("friendly-name", "GoCast Receiver", "friendly name")
	var iface = flag.String("iface", "", "interface to listen on (optional)")
	var port = flag.Int("port", 8009, "port to listen on")
	var relayAuthChallenge = flag.Bool("relay-auth-challenge", false, "send auth challenge when relaying")
	var relayHost = flag.String("relay-host", "", "relay to another Chromecast receiver (optional)")
	var relayPort = flag.Uint("relay-port", 8009, "port to relay to (optional)")

	flag.Parse()

	if *certManifest == "" && *certService == "" {
		flag.PrintDefaults()
		return
	}

	log.Info("args",
		"cert-manifest", *certManifest,
		"cert-service", *certService,
		"cert-service-salt", *certServiceSalt,
		"client-prefix", *clientPrefix,
		"device-model", *deviceModel,
		"enable-mdns", *enableMdns,
		"fix-newlines", *fixNewlines,
		"friendly-name", *friendlyName,
		"iface", *iface,
		"port", *port,
		"relay-auth-challenge", *relayAuthChallenge,
		"relay-host", *relayHost,
		"relay-port", *relayPort,
	)

	var manifest map[string]string
	if *certManifest != "" {
		manifest = ReadManifest(log, *certManifest, *fixNewlines)
	} else if *certService != "" {
		if *certServiceSalt == "" {
			log.Error("missing cert service salt")
			return
		}
		manifest = DownloadManifest(log, *certService, *certServiceSalt)
	}

	if manifest == nil {
		log.Error("failed to load manifest")
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	id := uuid.New().String()
	udn := id
	device := NewDevice(*deviceModel, *friendlyName, id, udn)

	server := NewServer(device, manifest, clientPrefix, iface, *port, *relayHost, *relayPort, *relayAuthChallenge, &wg)
	if server == nil {
		return
	}

	var advertisement *Advertisement
	if *enableMdns {
		advertisement = NewAdvertisement(device, *port)
		if advertisement == nil {
			log.Error("failed to advertise receiver")
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("interrupted")
		if advertisement != nil {
			advertisement.Stop()
		}
		server.StopListening()
		os.Exit(0)
	}()

	wg.Wait()
}
