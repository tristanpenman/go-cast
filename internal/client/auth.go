package client

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/tristanpenman/go-cast/internal/channel"
)

// These are Chromium's production Cast trust anchors. Device certificates must
// chain to one of these roots; roots supplied by the receiver are not trusted.
// Sources:
// https://chromium.googlesource.com/chromium/src/+/refs/tags/91.0.4472.77/components/cast_certificate/cast_root_ca_cert_der-inc.h
// https://chromium.googlesource.com/chromium/src/+/refs/tags/91.0.4472.77/components/cast_certificate/eureka_root_ca_der-inc.h
const castRootCAPEM = `-----BEGIN CERTIFICATE-----
MIIDxTCCAq2gAwIBAgIBAjANBgkqhkiG9w0BAQUFADB1MQswCQYDVQQGEwJVUzET
MBEGA1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwNTW91bnRhaW4gVmlldzETMBEG
A1UECgwKR29vZ2xlIEluYzENMAsGA1UECwwEQ2FzdDEVMBMGA1UEAwwMQ2FzdCBS
b290IENBMB4XDTE0MDQwMjE3MzQyNloXDTM0MDMyODE3MzQyNlowdTELMAkGA1UE
BhMCVVMxEzARBgNVBAgMCkNhbGlmb3JuaWExFjAUBgNVBAcMDU1vdW50YWluIFZp
ZXcxEzARBgNVBAoMCkdvb2dsZSBJbmMxDTALBgNVBAsMBENhc3QxFTATBgNVBAMM
DENhc3QgUm9vdCBDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALrZ
ZZ3aOdPBd/bU0K6PWAhoOUqV7XDP/XkIqarl6binLaBnR4qeyc9wswWHaRHscJiX
w+bDw+u9xrA9/E/BXjif2s9zMAZbeTfBXoyHR5SaQZIq1pXEcVwnXQixgMaSvRvj
QZeh7HWfVZ4+n48cx2VkB9OzlqEEn5HE3gp7bNnIwHgxoBlCqeiD48788c7CLiRG
lQkZysBGsuUButdP87/2aa2ZBPqgBzkO5t9RRwfA5KlcS5TFL7OgMH/nlWuyrzIN
8YzVbct7R6cIq8sno03PSlrxBdH4YsUQKnRpquZLlvub2GPkWGbTrYpu/3te+aVW
Hi2CMVvw4iTmQUofrhMCAwEAAaNgMF4wDwYDVR0TBAgwBgEB/wIBAjAdBgNVHQ4E
FgQUfJoefd95VLzXzF7KmYZFeWV0KBkwHwYDVR0jBBgwFoAUfJoefd95VLzXzF7K
mYZFeWV0KBkwCwYDVR0PBAQDAgEGMA0GCSqGSIb3DQEBBQUAA4IBAQCA9Fr7PSgZ
USDX1PsSl0pl8lg1kncwavHXtlEaf5rNx3sDQq1VagCv8OEGwr1reHXb/kERU0o5
u5o6xlk0Lywz47LWXH/deOtxWznag5DFMeI/I+/a6ystd17ew0PSyWtZgsrV7fqh
ZFvL8Q0aYuGc6KcYcPBfF5b47Ybbrh3gzz5dLu4WbZUrPP2X8wVaJGhNObb45Fi6
9eAmeFHFW11OCeVsR4t6Wi6JU+bMNlsmPPhyQwKC0ivN8NOj7BM+UtWDPQfcHUNl
ejMCAaPOt9ZgUTsJwiOKMv6YGWBik4XNNEbb1SMPedp3ACoCbYNYzgN3NeGjIJPC
SqKkRhx1LB9N
-----END CERTIFICATE-----
`

const eurekaRootCAPEM = `-----BEGIN CERTIFICATE-----
MIIDwzCCAqugAwIBAgIBATANBgkqhkiG9w0BAQUFADB8MQswCQYDVQQGEwJVUzET
MBEGA1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwNTW91bnRhaW4gVmlldzETMBEG
A1UECgwKR29vZ2xlIEluYzESMBAGA1UECwwJR29vZ2xlIFRWMRcwFQYDVQQDDA5F
dXJla2EgUm9vdCBDQTAeFw0xMjEyMTcyMjM5MzNaFw0zMjEyMTIyMjM5MzNaMHwx
CzAJBgNVBAYTAlVTMRMwEQYDVQQIDApDYWxpZm9ybmlhMRYwFAYDVQQHDA1Nb3Vu
dGFpbiBWaWV3MRMwEQYDVQQKDApHb29nbGUgSW5jMRIwEAYDVQQLDAlHb29nbGUg
VFYxFzAVBgNVBAMMDkV1cmVrYSBSb290IENBMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAuRHQ6hLcMuHfXDNrGXMdnZ7QOXa/pYQJpv1ubencjzZO6YgC
vZ/06ET9TPWaAlZqRypjbFhFzHxmJNx5ecMqpLKLoPeitc0Gftu+7AyG8g0kYHSE
yikjhALYp+078ewmR1TjsS3mZA/2csXpmFIXwPzyLCDIQPhHyTKeO5exi/WYJHBj
ZhnBUugEBT1fjbzYS693mG8feNG2UCdN5OwUaWcfWK+poBEmPJQyB3/X6Wkfrj9P
Y4qPidbyGXhcIY6xtlfYwOHufW7d8ToKavG6//mDL9y1pCAXYzbvyGIZzFbOsuox
iUt4WMG/AxOZ4BLyiKqblNrddnkXHjTRCsQHRQIDAQABo1AwTjAdBgNVHQ4EFgQU
RE4qR1jYuUiR9k/OdKkdMpqNjekwHwYDVR0jBBgwFoAURE4qR1jYuUiR9k/OdKkd
MpqNjekwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQUFAAOCAQEAP8gmoG5cBUB5
oZipM95odIXurrccM1mwEd6f9E/T61EJfUd+blGF9FTNg5glsbqwV+yT2xLi7FFJ
epZzm8iWbYWM0+E8+jLiWAx3bYcMNAGqMKl24MDn214b6RAwpOAJSSa5WM1aB+VQ
dd6aO/ZTfrFTXkUnTxfjCDOyUAq79PwllyneQXUw+nc4qmWKc0/qEXvrfBdgJw68
PnZS2IvtGvjrN7sR/a5wFwr+4K0Gsx9pinIEwsAzC9YvY0wzERS4YjaIxQNlARmj
7wC7bw6S/zQcodYx0Fxen5l9x8q9fHIL9Fylfm4EqNKZLFEBFP6iSPB+voQNtNPi
8w593ov1Mw==
-----END CERTIFICATE-----
`

func verifyDeviceAuthResponse(payload, peerCertificate, senderNonce []byte, now time.Time) error {
	var authMessage channel.DeviceAuthMessage
	if err := proto.Unmarshal(payload, &authMessage); err != nil {
		return fmt.Errorf("decode device-auth response: %w", err)
	}
	if authError := authMessage.GetError(); authError != nil {
		return fmt.Errorf("receiver reported device-auth error: %s", authError.GetErrorType())
	}
	response := authMessage.GetResponse()
	if response == nil {
		return errors.New("device-auth message contains no response")
	}
	if !bytes.Equal(response.GetSenderNonce(), senderNonce) {
		return errors.New("device-auth sender nonce mismatch")
	}
	if len(peerCertificate) == 0 {
		return errors.New("TLS peer certificate is missing")
	}
	peer, err := x509.ParseCertificate(peerCertificate)
	if err != nil {
		return fmt.Errorf("parse TLS peer certificate: %w", err)
	}
	if now.Before(peer.NotBefore) {
		return errors.New("TLS peer certificate is not valid yet")
	}
	if now.After(peer.NotAfter) {
		return errors.New("TLS peer certificate has expired")
	}

	deviceCertificate, err := x509.ParseCertificate(response.GetClientAuthCertificate())
	if err != nil {
		return fmt.Errorf("parse device-auth certificate: %w", err)
	}
	if err := verifyDeviceCertificate(deviceCertificate, response.GetIntermediateCertificate(), now); err != nil {
		return err
	}
	// The certificate chain and challenge signature provide device identity.
	// Cast's protobuf CRL requires its own pinned CRL root and parser; add that
	// separately before treating revoked-device detection as enforced here.

	publicKey, ok := deviceCertificate.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("unsupported device-auth public key type %T", deviceCertificate.PublicKey)
	}
	signatureInput := append(append([]byte(nil), senderNonce...), peerCertificate...)
	hash, digest, err := deviceAuthDigest(response.GetHashAlgorithm(), signatureInput)
	if err != nil {
		return err
	}
	switch response.GetSignatureAlgorithm() {
	case channel.SignatureAlgorithm_RSASSA_PKCS1v15:
		err = rsa.VerifyPKCS1v15(publicKey, hash, digest, response.GetSignature())
	case channel.SignatureAlgorithm_RSASSA_PSS:
		err = rsa.VerifyPSS(publicKey, hash, digest, response.GetSignature(), nil)
	default:
		return fmt.Errorf("unsupported device-auth signature algorithm %s", response.GetSignatureAlgorithm())
	}
	if err != nil {
		return fmt.Errorf("verify device-auth signature: %w", err)
	}
	return nil
}

func verifyDeviceCertificate(deviceCertificate *x509.Certificate, intermediateDER [][]byte, now time.Time) error {
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM([]byte(castRootCAPEM)) || !roots.AppendCertsFromPEM([]byte(eurekaRootCAPEM)) {
		return errors.New("load built-in Cast trust anchors")
	}
	intermediates := x509.NewCertPool()
	for _, certificateDER := range intermediateDER {
		certificate, err := x509.ParseCertificate(certificateDER)
		if err != nil {
			return fmt.Errorf("parse device-auth intermediate certificate: %w", err)
		}
		intermediates.AddCert(certificate)
	}
	_, err := deviceCertificate.Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   now,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	if err != nil {
		return fmt.Errorf("verify device-auth certificate chain: %w", err)
	}
	return nil
}

func deviceAuthDigest(algorithm channel.HashAlgorithm, input []byte) (crypto.Hash, []byte, error) {
	switch algorithm {
	case channel.HashAlgorithm_SHA1:
		digest := sha1.Sum(input)
		return crypto.SHA1, digest[:], nil
	case channel.HashAlgorithm_SHA256:
		digest := sha256.Sum256(input)
		return crypto.SHA256, digest[:], nil
	default:
		return 0, nil, fmt.Errorf("unsupported device-auth hash algorithm %s", algorithm)
	}
}

func parsePEMCertificate(certificatePEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certificatePEM))
	if block == nil {
		return nil, errors.New("decode certificate PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}
