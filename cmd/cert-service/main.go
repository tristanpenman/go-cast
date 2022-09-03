package main

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/tristanpenman/go-cast/internal"
)

var log = NewLogger("main")

var certManifestDir *string
var certServiceSalt *string

func validateChecksum(id string, timestamp string, checksum string) bool {
	key := []byte(id + timestamp + *certServiceSalt)
	md5Sum := md5.Sum(key)
	md5SumStr := hex.EncodeToString(md5Sum[:])
	return md5SumStr == checksum
}

func writeCompressed(writer http.ResponseWriter, fileBytes []byte) {
	gzipWriter := gzip.NewWriter(writer)
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/octet-stream")
	_, err := gzipWriter.Write(fileBytes)
	if err != nil {
		log.Error("")
		return
	}

	err = gzipWriter.Close()
	if err != nil {
		log.Error("")
		return
	}
}

func getRoot(writer http.ResponseWriter, request *http.Request) {
	log.Info("received request")

	query := request.URL.Query()
	if !query.Has("a") || !query.Has("b") || !query.Has("c") {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	id := query.Get("a")
	timestamp := query.Get("b")
	checksum := query.Get("c")

	if !validateChecksum(id, timestamp, checksum) {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	certManifestPath, err := MakeCertManifestPath(*certManifestDir, timestamp)
	if err != nil {
		log.Error("failed to make cert manifest path: " + err.Error())
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	fileBytes, err := ioutil.ReadFile(*certManifestPath)
	if err != nil {
		log.Error("failed to read cert manifest: " + err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeCompressed(writer, fileBytes)
}

func main() {
	certManifestDir = flag.String("cert-manifest-dir", "", "path to a directory containing timestamped cert manifest files")
	certServiceSalt = flag.String("cert-service-salt", "", "salt for generating cert service hash")

	flag.Parse()

	if *certManifestDir == "" || *certServiceSalt == "" {
		flag.PrintDefaults()
		return
	}

	if stat, err := os.Stat(*certManifestDir); os.IsNotExist(err) || !stat.IsDir() {
		log.Error("cert-manifest-dir must be a valid directory")
		return
	}

	http.HandleFunc("/", getRoot)
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		log.Error("failed to listen and serve", err)
	}
}
