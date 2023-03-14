package internal

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"net"
	"os"
	"path"

	// third-party
	"github.com/hashicorp/go-hclog"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/xlab/libvpx-go/vpx"

	// internal
	"github.com/tristanpenman/go-cast/internal/channel"
)

type Session struct {
	AppId       string
	DisplayName string
	SessionId   string
	StatusText  string

	// implementation
	device      *Device
	frameCount  int
	log         hclog.Logger
	packetConn  net.PacketConn
	streams     map[uint32]*Stream
	stop        chan struct{}
	stopping    bool
	transportId string
	vpxCtx      *vpx.CodecCtx
	vpxIface    *vpx.CodecIface
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
	Ssrc      uint32 `json:"ssrc"`
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
	CastMode             string   `json:"castMode"`
	ReceiverGetStatus    bool     `json:"receiverGetStatus"`
	ReceiverRtcpEventLog []int    `json:"receiverRtcpEventLog"`
	SendIndexes          []int    `json:"sendIndexes"`
	Ssrcs                []uint32 `json:"ssrcs"`
	UdpPort              int      `json:"udpPort"`
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
	var ssrcs []uint32

	for _, supportedStream := range request.Offer.SupportedStreams {
		if supportedStream.Type == "video_source" {
			receiverRtcpEventLog = append(receiverRtcpEventLog, supportedStream.Index)
			sendIndexes = append(sendIndexes, supportedStream.Index)
			ssrcs = append(ssrcs, supportedStream.Ssrc)

			key, _ := hex.DecodeString(supportedStream.AesKey)
			iv, _ := hex.DecodeString(supportedStream.AesIvMask)

			decrypter := NewDecrypter(key, iv)
			ssrc := supportedStream.Ssrc

			decode := func(buffer []byte, nextFrameId int) {
				plaintext := make([]byte, len(buffer))
				session.log.Info(fmt.Sprintf("decrypting %d bytes", len(buffer)), "frame id", nextFrameId)
				decrypter.Decrypt(buffer, plaintext)
				session.decodeBuffer(plaintext)
				decrypter.Reset(nextFrameId)
			}

			sendRtcp := func(buffer []byte, addr net.Addr) {
				session.packetConn.WriteTo(buffer, addr)
			}

			logger := NewLogger(fmt.Sprintf("stream (%d)", supportedStream.Ssrc))
			session.streams[ssrc] = NewStream(decode, logger, sendRtcp, supportedStream.Ssrc)
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

func (session *Session) Start() {
	go func() {
		select {
		case <-session.stop:
			session.packetConn.Close()
		}
	}()

	session.log.Info("listening on port", "port", GetPort(session.packetConn.LocalAddr()))

	go func() {
		data := make([]byte, 200000)

		for {
			count, addr, err := session.packetConn.ReadFrom(data)
			if session.stopping {
				session.log.Info("stopping udp listener")
				break
			} else if err != nil {
				session.log.Error(fmt.Sprintf("error while reading from socket: %s", err))
				break
			}

			session.log.Info(fmt.Sprintf("read %d bytes", count))

			packet := &rtp.Packet{}
			err = packet.Unmarshal(data[:count])
			if err != nil {
				session.log.Warn("error while unmarshalling rtp")
				return
			}

			if packet.PayloadType == 72 {
				// rtcp
				rtcpPackets, err := rtcp.Unmarshal(data[:count])
				if err != nil {
					session.log.Warn("error while unmarshalling rtcp")
					return
				}

				if len(rtcpPackets) == 0 {
					session.log.Warn("expected at least one rtcp packet")
					continue
				}

				// TODO
				stream := session.streams[rtcpPackets[0].DestinationSSRC()[0]]
				if stream == nil {
					session.log.Warn("stream not found", "ssrc", packet.SSRC, "seq", packet.SequenceNumber, "type", packet.PayloadType)
					continue
				}

				stream.handleRtcpPackets(rtcpPackets, addr)
			} else {
				// data
				stream := session.streams[packet.SSRC]
				if stream == nil {
					session.log.Warn("stream not found", "ssrc", packet.SSRC, "seq", packet.SequenceNumber, "type", packet.PayloadType)
					continue
				}

				stream.enqueuePacket(packet)
				nextPacket := stream.nextPacket()
				for nextPacket != nil {
					stream.handleDataPacket(nextPacket)
					nextPacket = stream.nextPacket()
				}
			}
		}

		session.packetConn.Close()
	}()
}

func (session *Session) Stop() {
	session.stopping = true
	close(session.stop)
}

func (session *Session) TransportId() string {
	return session.transportId
}

func (session *Session) decodeBuffer(payload []byte) {
	err := vpx.Error(vpx.CodecDecode(session.vpxCtx, string(payload), uint32(len(payload)), nil, 0))
	if err != nil {
		session.log.Error("failed to decode buffer: " + err.Error())
		return
	}

	var iter vpx.CodecIter
	image := vpx.CodecGetFrame(session.vpxCtx, &iter)
	if image != nil {
		image.Deref()
		session.frameCount++

		session.log.Info("image", "format", image.Fmt)

		session.device.DisplayImage(image.ImageRGBA())

		jpegBuffer := new(bytes.Buffer)
		if err = jpeg.Encode(jpegBuffer, image.ImageYCbCr(), nil); err != nil {
			session.log.Error("failed to encode jpeg: " + err.Error())
			return
		}

		jpegPath := path.Join("tmp", fmt.Sprintf("%d%s", session.frameCount, ".jpg"))
		fo, err := os.Create(jpegPath)
		if err != nil {
			session.log.Error("failed to create image: " + err.Error())
			return
		}

		if _, err := fo.Write(jpegBuffer.Bytes()); err != nil {
			session.log.Error("failed to write jpeg: " + err.Error())
			return
		}

		err = fo.Close()
		if err != nil {
			session.log.Warn("failed to close file: " + err.Error())
			return
		}
	}
}

func NewSession(appId string, clientId int, device *Device, displayName string, sessionId string, transportId string) *Session {
	log := NewLogger(fmt.Sprintf("session (%d) [%s]", clientId, sessionId))

	packetConn, err := net.ListenPacket("udp", ":50000")
	if err != nil {
		return nil
	}

	stop := make(chan struct{})

	vpxCtx := vpx.NewCodecCtx()
	vpxIface := vpx.DecoderIfaceVP8()

	err = vpx.Error(vpx.CodecDecInitVer(vpxCtx, vpxIface, nil, 0, vpx.DecoderABIVersion))
	if err != nil {
		log.Error(err.Error())
	}

	session := Session{
		AppId:       appId,
		DisplayName: displayName,
		SessionId:   sessionId,
		StatusText:  "",

		// internal
		device:      device,
		frameCount:  0,
		log:         log,
		packetConn:  packetConn,
		stop:        stop,
		stopping:    false,
		streams:     make(map[uint32]*Stream),
		transportId: transportId,
		vpxCtx:      vpxCtx,
		vpxIface:    vpxIface,
	}

	return &session
}
