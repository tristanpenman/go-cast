package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	. "github.com/tristanpenman/go-cast/internal"
	"github.com/tristanpenman/go-cast/internal/channel"
)

const (
	senderID            = "sender-0"
	receiverID          = "receiver-0"
	receiverNamespace   = "urn:x-cast:com.google.cast.receiver"
	connectionNamespace = "urn:x-cast:com.google.cast.tp.connection"
	heartbeatNamespace  = "urn:x-cast:com.google.cast.tp.heartbeat"
	mirroringAppID      = "0F5096E8"
)

var log = NewLogger("main")

type receiverMessage struct {
	RequestID int    `json:"requestId"`
	Type      string `json:"type"`
}

type launchRequest struct {
	*receiverMessage
	AppID string `json:"appId"`
}

type statusApplication struct {
	AppID       string `json:"appId"`
	TransportID string `json:"transportId"`
}

type statusPayload struct {
	Applications []statusApplication `json:"applications"`
}

type receiverStatus struct {
	*receiverMessage
	Status statusPayload `json:"status"`
}

type sender struct {
	client    *Client
	requestID int
}

func (s *sender) nextRequestID() int {
	s.requestID++
	return s.requestID
}

func newUTF8CastMessage(namespace, sourceID, destinationID, payload string) *channel.CastMessage {
	payloadType := channel.CastMessage_STRING
	protocolVersion := channel.CastMessage_CASTV2_1_0

	return &channel.CastMessage{
		DestinationId:   &destinationID,
		Namespace:       &namespace,
		PayloadType:     &payloadType,
		PayloadUtf8:     &payload,
		ProtocolVersion: &protocolVersion,
		SourceId:        &sourceID,
	}
}

func (s *sender) sendConnection(sourceID, destinationID string) {
	payload := `{"type":"CONNECT"}`
	s.client.SendMessage(newUTF8CastMessage(connectionNamespace, sourceID, destinationID, payload))
}

func (s *sender) requestStatus() {
	request := receiverMessage{RequestID: s.nextRequestID(), Type: "GET_STATUS"}
	payloadBytes, _ := json.Marshal(request)
	s.client.SendMessage(newUTF8CastMessage(receiverNamespace, senderID, receiverID, string(payloadBytes)))
}

func (s *sender) launchApp(appID string) {
	request := launchRequest{
		receiverMessage: &receiverMessage{RequestID: s.nextRequestID(), Type: "LAUNCH"},
		AppID:           appID,
	}
	payloadBytes, _ := json.Marshal(request)
	s.client.SendMessage(newUTF8CastMessage(receiverNamespace, senderID, receiverID, string(payloadBytes)))
}

func (s *sender) waitForSession(appID string, timeout time.Duration) (string, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return "", fmt.Errorf("timed out waiting for %s session", appID)
		case castMessage, ok := <-s.client.Incoming:
			if !ok {
				return "", errors.New("connection closed")
			}
			if castMessage == nil || castMessage.Namespace == nil {
				continue
			}

			switch *castMessage.Namespace {
			case heartbeatNamespace:
				if castMessage.PayloadUtf8 != nil && *castMessage.PayloadUtf8 == `{"type":"PING"}` {
					pong := `{"type":"PONG"}`
					s.client.SendMessage(newUTF8CastMessage(heartbeatNamespace, senderID, receiverID, pong))
				}
			case receiverNamespace:
				if castMessage.PayloadUtf8 == nil {
					continue
				}
				var status receiverStatus
				if err := json.Unmarshal([]byte(*castMessage.PayloadUtf8), &status); err != nil {
					log.Warn("failed to parse receiver payload", "err", err)
					continue
				}
				if status.Type != "RECEIVER_STATUS" {
					continue
				}
				for _, app := range status.Status.Applications {
					if app.AppID == appID {
						return app.TransportID, nil
					}
				}
			}
		}
	}
}

func main() {
	var disableChallenge = flag.Bool("disable-challenge", false, "disable auth challenge")
	var hostname = flag.String("hostname", "", "receiver address or hostname")
	var port = flag.Uint("port", 8009, "receiver port")
	var appID = flag.String("app-id", "", "Chromecast app ID to launch")
	var videoPath = flag.String("video-path", "", "path to local video file; launches Chromecast mirroring app")

	flag.Parse()

	if *hostname == "" {
		flag.PrintDefaults()
		return
	}

	effectiveAppID := *appID
	if *videoPath != "" {
		if _, err := os.Stat(*videoPath); err != nil {
			log.Error("video path is invalid", "path", *videoPath, "err", err)
			return
		}
		effectiveAppID = mirroringAppID
	}

	if effectiveAppID == "" {
		log.Error("either --app-id or --video-path must be provided")
		return
	}

	log.Info("args",
		"disable-challenge", *disableChallenge,
		"hostname", *hostname,
		"port", *port,
		"app-id", effectiveAppID,
		"video-path", *videoPath)

	var wg sync.WaitGroup
	wg.Add(1)

	client := NewClient(*hostname, *port, !*disableChallenge, &wg)
	if client == nil {
		return
	}

	s := sender{client: client}
	s.sendConnection(senderID, receiverID)
	s.requestStatus()
	s.launchApp(effectiveAppID)

	transportID, err := s.waitForSession(effectiveAppID, 10*time.Second)
	if err != nil {
		log.Error("failed to launch app", "err", err)
		_ = client.Close()
		wg.Wait()
		return
	}

	s.sendConnection(senderID, transportID)
	log.Info("app launched", "app-id", effectiveAppID, "transport-id", transportID)

	if *videoPath != "" {
		log.Warn("video-path support currently launches the Chromecast mirroring app only; stream upload is not implemented yet", "video-path", *videoPath)
	}

	_ = client.Close()
	wg.Wait()
}
