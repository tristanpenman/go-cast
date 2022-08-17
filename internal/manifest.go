package internal

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/grantae/certinfo"
	"github.com/hashicorp/go-hclog"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

func DownloadManifest(log hclog.Logger, certService string, certServiceSalt string) map[string]string {
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

func PrintCertificate(bytes []byte) {
	block, rest := pem.Decode(bytes)
	if block == nil || len(rest) > 0 {
		fmt.Println("Error: failed to decode pem data")
		return
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		fmt.Printf("Error: failed to parse certificate (%s)", err)
		return
	}

	result, err := certinfo.CertificateText(cert)
	if err != nil {
		fmt.Printf("Error: failed to generate certificate text (%s)", err)
		return
	}

	fmt.Println(result)
}

func PrintManifest(manifest map[string]string) {
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("Peer Certificate (pu)")
	fmt.Println("--------------------------------------------------------------------------------")
	PrintCertificate([]byte(manifest["pu"]))

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("Device Certificate (cpu)")
	fmt.Println("--------------------------------------------------------------------------------")
	PrintCertificate([]byte(manifest["cpu"]))

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("Intermediate CA Certificate (ica)")
	fmt.Println("--------------------------------------------------------------------------------")
	PrintCertificate([]byte(manifest["ica"]))
}

func ReadManifest(log hclog.Logger, certManifest string, fixNewlines bool) map[string]string {
	data, err := ioutil.ReadFile(certManifest)
	if err != nil {
		log.Error("failed to read certificate manifest file from disk", "err", err)
		return nil
	}

	// convert new-line characters so that JSON parses correctly
	var s = string(data)
	s = strings.TrimSuffix(s, "\n")
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
