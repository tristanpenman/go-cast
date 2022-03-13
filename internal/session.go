package internal

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/cast"
)

type Session struct {
	AppId       string
	DisplayName string
	SessionId   string
	StatusText  string

	// implementation
	device      *Device
	log         hclog.Logger
	packetConn  net.PacketConn
	transportId string
}

func (session *Session) GetPort() int {
	return GetPort(session.packetConn.LocalAddr())
}

type webrtcMessage struct {
	SeqNum uint32 `json:"seqNum"`
	Type   string `json:"type"`
}

type SupportedStream struct {
	AesIvMask string `json:"aesIvMask"`
	AesKey    string `json:"aesKey"`
	Index     int    `json:"index"`
	Ssrc      int    `json:"ssrc"`
	Type      string `json:"type"`
}

type Offer struct {
	CastMode          string            `json:"castMode"`
	ReceiverGetStatus bool              `json:"receiverGetStatus"`
	SupportedStreams  []SupportedStream `json:"supportedStreams"`
}

type webrtcOfferMessage struct {
	*webrtcMessage

	Offer Offer `json:"offer"`
}

type Audio struct {
}

type Video struct {
	MaxDimensions *Dimensions `json:"maxDimensions"`
	MinDimensions *Dimensions `json:"minDimensions"`
}

type Constraints struct {
	Audio *Audio `json:"audio"`
	Video *Video `json:"video"`
}

type Dimensions struct {
	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	FrameRate string `json:"frameRate"`
}

type Display struct {
	Dimensions  Dimensions `json:"dimensions"`
	AspectRatio string     `json:"aspectRatio"`
	Scaling     string     `json:"scaling"`
}

type Answer struct {
	CastMode          string      `json:"castMode"`
	UdpPort           int         `json:"udpPort"`
	SendIndexes       []int       `json:"sendIndexes"`
	Ssrcs             []int       `json:"ssrcs"`
	Constraints       Constraints `json:"constraints"`
	Display           Display     `json:"display"`
	ReceiverGetStatus bool        `json:"receiverGetStatus"`
}

type webrtcAnswerMessage struct {
	*webrtcMessage

	Answer Answer `json:"answer"`
}

func (session *Session) handleGenericMessage(castMessage *cast.CastMessage) {
	if *castMessage.PayloadType == cast.CastMessage_BINARY {
		session.log.Warn("ignoring message from unimplemented namespace",
			"namespace", *castMessage.Namespace,
			"payloadType", "STRING",
			"payloadUtf8", *castMessage.PayloadUtf8)
	} else {
		session.log.Warn("ignoring message from unimplemented namespace",
			"namespace", *castMessage.Namespace,
			"payloadType", "BINARY")
	}
}

func (session *Session) handleWebrtcOffer(castMessage *cast.CastMessage) {
	var request webrtcOfferMessage
	err := json.Unmarshal([]byte(*castMessage.PayloadUtf8), &request)
	if err != nil {
		session.log.Error("failed to unmarshall webrtc offer", "err", err)
		return
	}

	var sendIndexes []int
	var ssrcs []int

	for _, supportedStream := range request.Offer.SupportedStreams {
		if supportedStream.Type == "video_source" {
			sendIndexes = append(sendIndexes, supportedStream.Index)
			ssrcs = append(ssrcs, supportedStream.Ssrc)
		}
	}

	response := webrtcAnswerMessage{
		webrtcMessage: &webrtcMessage{
			Type: "ANSWER",
		},
		Answer: Answer{
			CastMode:    request.Offer.CastMode,
			UdpPort:     session.GetPort(),
			SendIndexes: sendIndexes,
			Ssrcs:       ssrcs,
			Constraints: Constraints{
				Video: &Video{
					MaxDimensions: &Dimensions{
						Width:     640,
						Height:    360,
						FrameRate: "60",
					},
				},
			},
			Display: Display{
				Dimensions: Dimensions{
					Width:     640,
					Height:    360,
					FrameRate: "60",
				},
				AspectRatio: "16:9",
				Scaling:     "sender",
			},
			ReceiverGetStatus: request.Offer.ReceiverGetStatus,
		},
	}

	bytes, err := json.Marshal(&response)
	if err != nil {
		session.log.Error("failed to marshall webrtc answer", "err", err)
		return
	}

	payloadUtf8 := string(bytes)
	session.device.sendUtf8(webrtcNamespace, &payloadUtf8, *castMessage.DestinationId, *castMessage.SourceId)
}

func (session *Session) handleWebrtcMessage(castMessage *cast.CastMessage) {
	var request webrtcMessage
	err := json.Unmarshal([]byte(*castMessage.PayloadUtf8), &request)
	if err != nil {
		session.log.Error("failed to unmarshall webrtc message", "err", err)
		return
	}

	switch request.Type {
	case "OFFER":
		session.handleWebrtcOffer(castMessage)
		break
	default:
		session.log.Error("unrecognised webrtc request type", "type", request.Type)
		break
	}
}

func (session *Session) HandleCastMessage(castMessage *cast.CastMessage) {
	switch *castMessage.Namespace {
	case debugNamespace:
	case mediaNamespace:
	case remotingNamespace:
		break
	case webrtcNamespace:
		session.handleWebrtcMessage(castMessage)
		break
	default:

	}
}

func (session *Session) Namespaces() []string {
	namespaces := make([]string, 4)

	namespaces[0] = debugNamespace
	namespaces[1] = mediaNamespace
	namespaces[2] = remotingNamespace
	namespaces[3] = webrtcNamespace

	return namespaces
}

func (session *Session) TransportId() string {
	return session.transportId
}

func NewSession(appId string, device *Device, displayName string, sessionId string, transportId string) *Session {
	log := NewLogger(fmt.Sprintf("session (%s)", sessionId))

	packetConn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil
	}

	go func() {
		bytes := make([]byte, 1500)
		for {
			count, peer, err := packetConn.ReadFrom(bytes)
			if err != nil {
				log.Error(fmt.Sprintf("error while reading from socket: %d", err))
				packetConn.Close()
			}

			log.Info("read %d from %s", count, peer.Network())
		}
	}()

	return &Session{
		AppId:       appId,
		DisplayName: displayName,
		SessionId:   sessionId,
		StatusText:  "",

		// implementation
		device:      device,
		log:         log,
		packetConn:  packetConn,
		transportId: transportId,
	}
}