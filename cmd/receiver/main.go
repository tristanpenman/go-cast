package main

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"

	. "github.com/tristanpenman/go-cast/internal"
)

var log = NewLogger("main")

func randomId() (int64, error) {
	val, err := rand.Int(rand.Reader, big.NewInt(int64(math.MaxInt64)))
	if err != nil {
		return 0, err
	}
	return val.Int64(), nil
}

func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	resData = resB.Bytes()

	return
}

func downloadManifest(certService string, certServiceSalt string) map[string]string {
	var id, _ = randomId()

	var a = md5.Sum([]byte(strconv.FormatInt(id, 10)))
	var aStr = hex.EncodeToString(a[:])

	var b = time.Now().Unix()
	var bStr = strconv.FormatInt(b, 10)

	var c = md5.Sum([]byte(aStr + bStr + certServiceSalt))
	var cStr = hex.EncodeToString(c[:])

	var url = certService + "?a=" + aStr + "&b=" + bStr + "&c=" + cStr

	log.Error(url)

	resp, err := http.Get(url)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to download %s: %d", url, err))
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to read response body: %d", err))
		return nil
	}

	data, err := gUnzipData(body)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to unzip: %d", err))
		return nil
	}

	var s = string(data)
	s = strings.TrimSuffix(s, "\n")
	s = strings.ReplaceAll(s, "\n", "\\n")

	var manifest map[string]string
	err = json.Unmarshal([]byte(s), &manifest)
	if err != nil {
		log.Error("failed to parse certificate manifest file", "err", err)
		return nil
	}

	return manifest
}

func readManifest(certManifest string, fixNewlines bool) map[string]string {
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
		manifest = readManifest(*certManifest, *fixNewlines)
	} else if *certService != "" {
		if *certServiceSalt == "" {
			log.Error("missing cert service salt")
			return
		}
		manifest = downloadManifest(*certService, *certServiceSalt)
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
