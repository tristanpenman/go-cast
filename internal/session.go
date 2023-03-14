package internal

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pion/rtp"
	"image/jpeg"
	"io"
	"net"
	"os"
	"path"

	// third-party
	"github.com/hashicorp/go-hclog"
	"github.com/xlab/libvpx-go/vpx"

	// internal
	"github.com/tristanpenman/go-cast/internal/channel"
)

type packetInfo struct {
	keyframe    bool
	hasRef      bool
	numExt      int
	frameId     int
	packetId    uint16
	maxPacketId uint32
	payload     []byte
	ssrc        uint32
	seq         uint16
}

type Session struct {
	AppId       string
	DisplayName string
	SessionId   string
	StatusText  string

	// implementation
	ciphertext     []byte
	currentFrameId int
	decrypters     map[uint32]*Decrypter
	device         *Device
	frameCount     int
	log            hclog.Logger
	nextSeq        int
	packetConn     net.PacketConn
	prevFrameId    int
	stop           chan struct{}
	stopping       bool
	transportId    string
	videoPackets   map[uint16]*packetInfo
	vpxCtx         *vpx.CodecCtx
	vpxIface       *vpx.CodecIface

	// receiver report
	highestSeq int
	newFrameId int
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

			key, _ := hex.DecodeString(supportedStream.AesKey)
			iv, _ := hex.DecodeString(supportedStream.AesIvMask)

			decrypter := NewDecrypter(key, iv)

			session.log.Info("registering decrypter", "ssrc", supportedStream.Ssrc)

			ssrc := uint32(supportedStream.Ssrc)
			session.decrypters[ssrc] = decrypter
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
			count, _, err := session.packetConn.ReadFrom(data)
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

			switch packet.PayloadType {
			case 72:
				session.handleRtcp(packet)
				break
			case 96:
				session.enqueuePacket(packet)
				packetInfo := session.nextPacket()
				for packetInfo != nil {
					session.handleVideo(packetInfo)
					packetInfo = session.nextPacket()
				}
				break
			default:
				session.log.Warn("unknown payload type", "payloadType", packet.PayloadType)
				break
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

func (session *Session) handleAudio(packet *rtp.Packet) {
	// TODO
	session.log.Info("received audio packet",
		"length", len(packet.Payload))
}

func (session *Session) handleRtcp(packet *rtp.Packet) {
	// TODO
	session.log.Info("received rtcp packet",
		"length", len(packet.Payload))

	// make a fake sender report

}

func (session *Session) enqueuePacket(packet *rtp.Packet) {
	bits := int(packet.Payload[0])

	payloadReader := bytes.NewReader(packet.Payload)

	packetId := make([]byte, 2)
	payloadReader.Seek(2, io.SeekStart)
	binary.Read(payloadReader, binary.BigEndian, packetId)

	maxPacketId := make([]byte, 4)
	payloadReader.Seek(4, io.SeekStart)
	binary.Read(payloadReader, binary.BigEndian, maxPacketId)

	packetInfo := &packetInfo{
		keyframe:    (bits & 0x80) != 0,
		hasRef:      (bits & 0x40) != 0,
		numExt:      bits & 0x3f,
		frameId:     int(packet.Payload[1]),
		packetId:    binary.BigEndian.Uint16(packetId),
		maxPacketId: binary.BigEndian.Uint32(maxPacketId),
		payload:     packet.Payload,
		seq:         packet.SequenceNumber,
		ssrc:        packet.SSRC,
	}

	session.videoPackets[packet.SequenceNumber] = packetInfo
	session.log.Info("enqueued packet", "sequenceNumber", packet.SequenceNumber)
}

func (session *Session) nextPacket() *packetInfo {
	for seq := range session.videoPackets {
		if int(seq) == session.nextSeq || session.nextSeq == -1 {
			packetInfo := session.videoPackets[seq]
			delete(session.videoPackets, seq)
			session.nextSeq = int(seq + 1)
			session.log.Info("dequeueing packet", "sequenceNumber", packetInfo.seq, "packetId", packetInfo.packetId)
			return packetInfo
		}
	}

	return nil
}

func (session *Session) handleVideo(packetInfo *packetInfo) {
	decrypter := session.decrypters[packetInfo.ssrc]

	payloadReader := bytes.NewReader(packetInfo.payload)

	session.log.Info("frame",
		"keyframe", packetInfo.keyframe,
		"hasRef", packetInfo.hasRef,
		"numExt", packetInfo.numExt,
		"frameId", packetInfo.frameId,
		"prevFrameId", session.prevFrameId,
		"packetId", packetInfo.packetId,
		"maxPacketId", packetInfo.maxPacketId)

	if packetInfo.frameId != session.prevFrameId {
		plaintext := make([]byte, len(session.ciphertext))
		session.log.Info(fmt.Sprintf("decrypting %d bytes", len(session.ciphertext)))
		decrypter.Decrypt(session.ciphertext, plaintext)
		session.decodeBuffer(plaintext)
		decrypter.Reset(packetInfo.frameId)
		session.ciphertext = make([]byte, 0)
		session.prevFrameId = packetInfo.frameId
	}

	offset := 6
	if packetInfo.hasRef {
		payloadReader.Seek(int64(offset), io.SeekStart)
		refId, _ := payloadReader.ReadByte()
		session.log.Info(fmt.Sprintf("ref id: %d", refId))
		offset++
	}

	for i := 0; i < packetInfo.numExt; i++ {
		payloadReader.Seek(int64(offset), io.SeekStart)
		typeAndSizeBytes := make([]byte, 2)
		binary.Read(payloadReader, binary.BigEndian, typeAndSizeBytes)

		typeAndSize := int(binary.BigEndian.Uint16(typeAndSizeBytes))
		dataType := typeAndSize >> 10
		dataSize := typeAndSize & 0x3FF

		if dataType == 1 {
			session.log.Info("ignored adaptive latency extension")
		} else {
			session.log.Info("ignoring unknown extension type", "dataType", dataType, "dataSize", dataSize)
		}

		offset += dataSize + 2
	}

	session.ciphertext = append(session.ciphertext, packetInfo.payload[offset:]...)
}

func NewSession(appId string, clientId int, device *Device, displayName string, sessionId string, transportId string) *Session {
	log := NewLogger(fmt.Sprintf("session (%d) [%s]", clientId, sessionId))

	packetConn, err := net.ListenPacket("udp", ":0")
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

		// implementation
		ciphertext:     make([]byte, 0),
		currentFrameId: -1,
		decrypters:     make(map[uint32]*Decrypter),
		device:         device,
		log:            log,
		nextSeq:        -1,
		packetConn:     packetConn,
		prevFrameId:    0,
		stop:           stop,
		stopping:       false,
		transportId:    transportId,
		videoPackets:   make(map[uint16]*packetInfo),
		vpxCtx:         vpxCtx,
		vpxIface:       vpxIface,

		// receiver report
		highestSeq: -1,
		newFrameId: -1,
	}

	return &session
}
