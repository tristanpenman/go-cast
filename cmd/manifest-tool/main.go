package main

import "flag"

import . "github.com/tristanpenman/go-cast/internal"

var log = NewLogger("main")

func main() {
	var certManifest = flag.String("cert-manifest", "", "path to certificate manifest")
	var certService = flag.String("cert-service", "", "base URL for certificate service")
	var certServiceSalt = flag.String("cert-service-salt", "", "salt for generating cert service hash")
	var fixNewlines = flag.Bool("fix-newlines", false, "fix newline characters in manifest file")
	var printManifest = flag.Bool("print-manifest", false, "print manifest details to terminal")

	flag.Parse()

	if *certManifest == "" && *certService == "" {
		flag.PrintDefaults()
		return
	}

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

	if *printManifest {
		log.Info("print manifest")
		PrintManifest(manifest)
	}
}
