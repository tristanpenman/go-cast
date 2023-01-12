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
	var useSha256 = flag.Bool("use-sha-256", false, "use SHA256 for signature verification")
	var verifySignature = flag.Bool("verify-signature", false, "verify signature")

	flag.Parse()

	if *certManifest == "" && *certService == "" {
		flag.PrintDefaults()
		return
	}

	var manifest map[string]string
	var err error

	if *certManifest != "" {
		manifest, err = ReadManifest(log, *certManifest, *fixNewlines)
		if err != nil {
			log.Error("failed to read manifest: " + err.Error())
			return
		}
	} else if *certService != "" {
		if *certServiceSalt == "" {
			log.Error("missing cert service salt")
			return
		}
		manifest, err = DownloadManifest(log, *certService, *certServiceSalt)
		if err != nil {
			log.Error("failed to download manifest: " + err.Error())
		}
	}

	if *printManifest {
		PrintManifest(manifest)
	}

	if *verifySignature {
		VerifySignature(manifest, *useSha256)
	}
}
