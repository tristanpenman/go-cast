package main

import (
	"flag"
	"github.com/google/uuid"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	. "github.com/tristanpenman/go-cast/internal"
)

var log = NewLogger("main")

func resolveManifest(certManifest string, certManifestDir string, certService string, certServiceSalt string, fixNewlines bool) map[string]string {
	if certManifest != "" {
		log.Info("attempting to read manifest from file: " + certManifest)

		manifest, err := ReadManifest(log, certManifest, fixNewlines)
		if err == nil {
			return manifest
		}

		log.Warn("failed to read manifest: " + err.Error())
	}

	if certManifestDir != "" {
		log.Info("attempting to find manifest in directory: " + certManifestDir)

		path, err := MakeCertManifestPath(certManifestDir, strconv.FormatInt(time.Now().Unix(), 10))
		if err != nil {
			log.Error("failed to make cert manifest path: " + err.Error())
		}

		if path != nil {
			log.Info("attempting to read manifest from file: " + *path)

			manifest, err := ReadManifest(log, *path, fixNewlines)
			if err == nil {
				return manifest
			}

			log.Warn("failed to read manifest: " + err.Error())
		}
	}

	if certService != "" {
		log.Info("attempting to download manifest from cert service: " + certService)

		manifest, err := DownloadManifest(log, certService, certServiceSalt)
		if err == nil {
			return manifest
		}

		log.Warn("failed to download manifest")
	}

	return nil
}

func main() {
	var certManifest = flag.String("cert-manifest", "", "path to a cert manifest file")
	var certManifestDir = flag.String("cert-manifest-dir", "", "path to a directory containing cert manifests")
	var certService = flag.String("cert-service", "", "base URL for certificate service")
	var certServiceSalt = flag.String("cert-service-salt", "", "salt for generating cert service hash")
	var clientPrefix = flag.String("client-prefix", "", "optional client prefix, to limit connections")
	var deviceModel = flag.String("device-model", "go-cast", "device model")
	var enableMdns = flag.Bool("enable-mdns", false, "advertise service using mDNS")
	var fixNewlines = flag.Bool("fix-newlines", false, "fix newline characters in manifest file")
	var friendlyName = flag.String("friendly-name", "GoCast Receiver", "friendly name")
	var iface = flag.String("iface", "", "interface to listen on (optional)")
	var port = flag.Int("port", 8009, "port to listen on")

	flag.Parse()

	if *certManifest == "" && *certManifestDir == "" && *certService == "" {
		flag.PrintDefaults()
		return
	}

	log.Info("args",
		"cert-manifest", *certManifest,
		"cert-manifest-dir", *certManifestDir,
		"cert-service", *certService,
		"cert-service-salt", *certServiceSalt,
		"client-prefix", *clientPrefix,
		"device-model", *deviceModel,
		"enable-mdns", *enableMdns,
		"fix-newlines", *fixNewlines,
		"friendly-name", *friendlyName,
		"iface", *iface,
		"port", *port,
	)

	manifest := resolveManifest(*certManifest, *certManifestDir, *certService, *certServiceSalt, *fixNewlines)
	if manifest == nil {
		log.Error("failed to load manifest from any sources")
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	id := uuid.New().String()
	udn := id
	device := NewDevice(*deviceModel, *friendlyName, id, udn)

	server := NewServer(device, manifest, clientPrefix, iface, *port, &wg)
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
