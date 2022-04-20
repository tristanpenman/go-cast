package internal

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/hashicorp/go-hclog"

	"github.com/tristanpenman/go-cast/internal/channel"
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
	stop        chan struct{}
	stopping    bool
	transportId string
}

func (session *Session) GetPort() int {
	return GetPort(session.packetConn.LocalAddr())
}

type WebrtcMessage struct {
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
	*WebrtcMessage

	Offer Offer `json:"offer"`
}

type Audio struct {
	MaxSampleRate int `json:"maxSampleRate"`
	MaxChannels   int `json:"maxChannels"`
	MinBitRate    int `json:"minBitRate"`
	MaxBitRate    int `json:"maxBitRate"`
	MaxDelay      int `json:"maxDelay"`
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
	CastMode             string `json:"castMode"`
	ReceiverGetStatus    bool   `json:"receiverGetStatus"`
	ReceiverRtcpEventLog []int  `json:"receiverRtcpEventLog"`
	SendIndexes          []int  `json:"sendIndexes"`
	Ssrcs                []int  `json:"ssrcs"`
	UdpPort              int    `json:"udpPort"`
}

type webrtcAnswerMessage struct {
	*WebrtcMessage

	Answer Answer `json:"answer"`
	Result string `json:"result"`
}

func (session *Session) handleGenericMessage(castMessage *channel.CastMessage) {
	if *castMessage.PayloadType == channel.CastMessage_BINARY {
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

func (session *Session) handleWebrtcOffer(castMessage *channel.CastMessage) {
	var request webrtcOfferMessage
	err := json.Unmarshal([]byte(*castMessage.PayloadUtf8), &request)
	if err != nil {
		session.log.Error("failed to unmarshall webrtc offer", "err", err)
		return
	}

	var receiverRtcpEventLog []int
	var sendIndexes []int
	var ssrcs []int

	for _, supportedStream := range request.Offer.SupportedStreams {
		if supportedStream.Type == "video_source" {
			receiverRtcpEventLog = append(receiverRtcpEventLog, supportedStream.Index)
			sendIndexes = append(sendIndexes, supportedStream.Index)
			ssrcs = append(ssrcs, supportedStream.Ssrc)
		}
	}

	response := webrtcAnswerMessage{
		WebrtcMessage: &WebrtcMessage{
			SeqNum: request.SeqNum,
			Type:   "ANSWER",
		},
		Answer: Answer{
			CastMode:             request.Offer.CastMode,
			ReceiverGetStatus:    request.Offer.ReceiverGetStatus,
			ReceiverRtcpEventLog: receiverRtcpEventLog,
			SendIndexes:          sendIndexes,
			Ssrcs:                ssrcs,
			UdpPort:              session.GetPort(),
		},
		Result: "ok",
	}

	bytes, err := json.Marshal(&response)
	if err != nil {
		session.log.Error("failed to marshall webrtc answer", "err", err)
		return
	}

	payloadUtf8 := string(bytes)
	session.device.sendUtf8(webrtcNamespace, &payloadUtf8, *castMessage.DestinationId, *castMessage.SourceId)
}

func (session *Session) handleWebrtcMessage(castMessage *channel.CastMessage) {
	var request WebrtcMessage
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

func (session *Session) HandleCastMessage(castMessage *channel.CastMessage) {
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

func (session *Session) Stop() {
	session.stopping = true
	close(session.stop)
}

func (session *Session) TransportId() string {
	return session.transportId
}

func NewSession(appId string, clientId int, device *Device, displayName string, sessionId string, transportId string) *Session {
	log := NewLogger(fmt.Sprintf("session (%d) [%s]", clientId, sessionId))

	packetConn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil
	}

	stop := make(chan struct{})
	go func() {
		select {
		case <-stop:
			packetConn.Close()
		}
	}()

	session := Session{
		AppId:       appId,
		DisplayName: displayName,
		SessionId:   sessionId,
		StatusText:  "",

		// implementation
		device:      device,
		log:         log,
		packetConn:  packetConn,
		stop:        stop,
		stopping:    false,
		transportId: transportId,
	}

	log.Info("listening on port", "port", GetPort(packetConn.LocalAddr()))

	go func() {
		bytes := make([]byte, 1500)
		for {
			count, peer, err := packetConn.ReadFrom(bytes)
			if session.stopping {
				log.Info("stopping udp listener")
				break
			} else if err != nil {
				log.Error(fmt.Sprintf("error while reading from socket: %s", err))
				break
			}

			log.Info("read %d from %s", count, peer.Network())
		}

		packetConn.Close()
	}()

	return &session
}
