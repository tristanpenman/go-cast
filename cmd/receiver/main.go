package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	var certManifest = flag.String("cert-manifest", "", "path to certificate manifest")
	var clientIp = flag.String("client-prefix", "", "optional client prefix, to limit connections")
	var iface = flag.String("iface", "", "interface to listen on (optional)")
	var port = flag.Uint("port", 8009, "port to listen on")

	flag.Parse()

	if *certManifest == "" {
		flag.PrintDefaults()
		return
	}

	logger.Info("args", "cert-manifest", *certManifest, "iface", *iface, "port", *port)

	// read manifest from disk
	data, err := ioutil.ReadFile(*certManifest)
	if err != nil {
		logger.Error("Failed to read cert manifest", "err", err)
		return
	}

	// convert new-line characters so that JSON parses correctly
	s := string(data)
	t := strings.ReplaceAll(s, "\n", "\\n")

	// parse manifest
	var manifest map[string]string
	err = json.Unmarshal([]byte(t), &manifest)
	if err != nil {
		logger.Error("Failed to parse device-auth manifest", "err", err)
		return
	}

	if logger.IsDebug() {
		logger.Debug("Manifest contents")
		for key, value := range manifest {
			logger.Debug(fmt.Sprintf("%s: %s", key, value))
		}
	}

	startServer(manifest, clientIp, iface, port)
}
