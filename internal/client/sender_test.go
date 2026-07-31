package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestNativeYouTubeMatchesUniversalAppID(t *testing.T) {
	sender := testSender()
	payload := `{"requestId":2,"responseType":"RECEIVER_STATUS","status":{"applications":[{"appId":"2C6A6E3D","appType":"ANDROID_TV","displayName":"YouTube","universalAppId":"233637DE","sessionId":"session-1","transportId":"transport-1"}]}}`
	sender.handleReceiverMessage(receiverMessage(payload))

	if transportID := sender.SessionTransportID("233637DE"); transportID != "transport-1" {
		t.Fatalf("universal YouTube ID returned transport ID %q", transportID)
	}
	status := sender.Status()
	if status == nil || len(status.Applications) != 1 || status.Applications[0].AppType != "ANDROID_TV" {
		t.Fatalf("unexpected native YouTube status: %+v", status)
	}
}

func TestIdleScreenIsNotTreatedAsRunning(t *testing.T) {
	sender := testSender()
	payload := `{"requestId":3,"responseType":"RECEIVER_STATUS","status":{"applications":[{"appId":"233637DE","displayName":"YouTube","isIdleScreen":true,"sessionId":"idle-session","transportId":"idle-transport"}]}}`
	sender.handleReceiverMessage(receiverMessage(payload))

	if transportID := sender.TransportID("233637DE"); transportID != "" {
		t.Fatalf("idle screen returned transport ID %q", transportID)
	}
	if sender.appRunningLocked("233637DE") {
		t.Fatal("idle screen was treated as a running app")
	}
	if transportID := sender.SessionTransportID("233637DE"); transportID != "idle-transport" {
		t.Fatalf("idle app session returned transport ID %q", transportID)
	}
	transportID, err := sender.WaitForAppTransport("233637DE", time.Second)
	if err != nil {
		t.Fatalf("wait for idle app transport: %v", err)
	}
	if transportID != "idle-transport" {
		t.Fatalf("wait returned transport ID %q", transportID)
	}
}

func TestHandleYouTubeScreenStatus(t *testing.T) {
	sender := testSender()
	payload := `{"type":"mdxSessionStatus","data":{"screenId":"screen-123"}}`
	namespace := youtubeNamespace
	sender.handleYouTubeMessage(&channel.CastMessage{Namespace: &namespace, PayloadUtf8: &payload})

	if sender.youtubeScreenID != "screen-123" {
		t.Fatalf("unexpected YouTube screen ID %q", sender.youtubeScreenID)
	}
}

func TestPlayYouTubeViaLounge(t *testing.T) {
	requestNumber := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestNumber++
		if err := request.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		switch requestNumber {
		case 1:
			if request.URL.Path != "/api/lounge/pairing/get_lounge_token_batch" || request.Form.Get("screen_ids") != "screen-123" {
				t.Errorf("unexpected token request: %s %+v", request.URL.Path, request.Form)
			}
			_, _ = fmt.Fprint(response, `{"screens":[{"loungeToken":"lounge-token"}]}`)
		case 2:
			if request.URL.Path != "/api/lounge/bc/bind" || request.URL.Query().Get("RID") != "0" || request.URL.Query().Get("TYPE") != "bind" {
				t.Errorf("unexpected bind request: %s %+v", request.URL.Path, request.URL.Query())
			}
			if request.Header.Get("X-YouTube-LoungeId-Token") != "lounge-token" || request.Form.Get("device") != "REMOTE_CONTROL" || request.Form.Get("loungeIdToken") != "lounge-token" {
				t.Errorf("unexpected bind headers or form: %+v %+v", request.Header, request.Form)
			}
			_, _ = fmt.Fprint(response, `[[0,["c","sid-123","",8]],[1,["S","gsession-123"]]]`)
		case 3:
			if request.URL.Query().Get("SID") != "sid-123" || request.URL.Query().Get("gsessionid") != "gsession-123" || request.URL.Query().Get("TYPE") != "bind" {
				t.Errorf("unexpected session request: %+v", request.URL.Query())
			}
			if request.Form.Get("req0__sc") != "setPlaylist" || request.Form.Get("req0_videoId") != "video-123" || request.Form.Get("ofs") != "0" {
				t.Errorf("unexpected playlist request: %+v", request.Form)
			}
			if _, malformed := request.Form["req0__videoId"]; malformed {
				t.Errorf("playlist request contains malformed video field: %+v", request.Form)
			}
		default:
			t.Errorf("unexpected extra request %d", requestNumber)
		}
	}))
	defer server.Close()

	if err := playYouTubeViaLounge(context.Background(), server.Client(), server.URL, "screen-123", "video-123"); err != nil {
		t.Fatal(err)
	}
	if requestNumber != 3 {
		t.Fatalf("expected three YouTube requests, got %d", requestNumber)
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
