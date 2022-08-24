package main

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

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

func beginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func findCertManifest(timestamp string) *string {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		log.Error("failed to parse timestamp", err)
		return nil
	}

	unixTime := time.Unix(i, 0).UTC()
	log.Info("original time: " + unixTime.String())

	bod := beginningOfDay(unixTime)
	log.Info("beginning of day: " + bod.String())

	filename := fmt.Sprintf("certs-%04d%02d%02d.json", bod.Year(), bod.Month(), bod.Day())
	result := path.Join(*certManifestDir, filename)
	return &result
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

	certManifestPath := findCertManifest(timestamp)
	if certManifestPath == nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	fileBytes, err := ioutil.ReadFile(*certManifestPath)
	if err != nil {
		log.Error(err.Error())
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
