package client

import (
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/channel"
	"github.com/tristanpenman/go-cast/internal/common"
)

func TestHandleReceiverAppAvailability(t *testing.T) {
	sender := testSender()
	payload := `{"requestId":1,"responseType":"GET_APP_AVAILABILITY","availability":{"233637DE":"APP_AVAILABLE"}}`
	sender.handleReceiverMessage(receiverMessage(payload))

	got := sender.Availability()
	if got["233637DE"] != "APP_AVAILABLE" {
		t.Fatalf("unexpected availability: %+v", got)
	}
}

func TestHandleReceiverStatusIncludesSessionID(t *testing.T) {
	sender := testSender()
	payload := `{"requestId":2,"responseType":"RECEIVER_STATUS","status":{"applications":[{"appId":"233637DE","displayName":"YouTube","sessionId":"session-1","statusText":"Ready","transportId":"transport-1"}]}}`
	sender.handleReceiverMessage(receiverMessage(payload))

	status := sender.Status()
	if status == nil || len(status.Applications) != 1 {
		t.Fatalf("unexpected status: %+v", status)
	}
	app := status.Applications[0]
	if app.SessionID != "session-1" || app.TransportID != "transport-1" {
		t.Fatalf("unexpected application: %+v", app)
	}
}

func TestVerifyDeviceAuthResponseRejectsMalformedPayload(t *testing.T) {
	if err := verifyDeviceAuthResponse([]byte("not protobuf"), []byte("peer cert"), nil, time.Now()); err == nil {
		t.Fatal("expected malformed device-auth response to fail")
	}
}

func TestBuiltInCastRootIsTrusted(t *testing.T) {
	root, err := parsePEMCertificate(castRootCAPEM)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyDeviceCertificate(root, nil, time.Now()); err != nil {
		t.Fatalf("expected built-in Cast root to be trusted: %v", err)
	}
}

func testSender() *Sender {
	sender := &Sender{log: hclog.NewNullLogger()}
	sender.cond = sync.NewCond(&sender.mu)
	return sender
}

func receiverMessage(payload string) *channel.CastMessage {
	namespace := common.ReceiverNamespace
	return &channel.CastMessage{Namespace: &namespace, PayloadUtf8: &payload}
}
