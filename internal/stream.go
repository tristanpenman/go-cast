package internal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	// third-party
	"github.com/hashicorp/go-hclog"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
)

type Stream struct {
	actualFrameId int
	buffer        []byte
	ckptFrameId   int
	decode        func([]byte, int)
	highestSeq    uint16
	log           hclog.Logger
	nextSeq       int
	newFrameId    int
	ntpTime       uint64
	packetsQueue  map[uint16]*rtp.Packet
	receiverSsrc  uint32
	rtpTime       uint32
	sendRtcp      func([]byte, net.Addr)
	senderSsrc    uint32
}

func (stream *Stream) enqueuePacket(packet *rtp.Packet) {
	stream.packetsQueue[packet.SequenceNumber] = packet
	stream.log.Debug("enqueued packet", "sequenceNumber", packet.SequenceNumber)

	if packet.SequenceNumber > stream.highestSeq {
		stream.highestSeq = packet.SequenceNumber
	}
}

func (stream *Stream) nextPacket() *rtp.Packet {
	for seq := range stream.packetsQueue {
		if int(seq) == stream.nextSeq || stream.nextSeq == -1 {
			packet := stream.packetsQueue[seq]
			delete(stream.packetsQueue, seq)
			stream.nextSeq = int(seq + 1)
			stream.log.Debug("next packet", "sequenceNumber", seq)
			return packet
		}
	}

	return nil
}

func (stream *Stream) handleDataPacket(packet *rtp.Packet, addr net.Addr) {
	bits := int(packet.Payload[0])
	keyframe := (bits & 0x80) != 0
	hasRef := (bits & 0x40) != 0
	numExt := bits & 0x3f
	frameId := int(packet.Payload[1])

	//	payload:     packet.Payload,
	//	seq:         packet.SequenceNumber,
	//	ssrc:        packet.SSRC,

	payloadReader := bytes.NewReader(packet.Payload)

	packetIdBytes := make([]byte, 2)
	payloadReader.Seek(2, io.SeekStart)
	binary.Read(payloadReader, binary.BigEndian, packetIdBytes)
	packetId := binary.BigEndian.Uint16(packetIdBytes)

	maxPacketIdBytes := make([]byte, 4)
	payloadReader.Seek(4, io.SeekStart)
	binary.Read(payloadReader, binary.BigEndian, maxPacketIdBytes)
	maxPacketId := binary.BigEndian.Uint32(maxPacketIdBytes)

	stream.log.Info("frame",
		"keyframe", keyframe,
		"hasRef", hasRef,
		"numExt", numExt,
		"frameId", frameId,
		"ckptFrameId", stream.ckptFrameId,
		"packetId", packetId,
		"maxPacketId", maxPacketId)

	offset := 6
	if hasRef {
		payloadReader.Seek(int64(offset), io.SeekStart)
		refId, _ := payloadReader.ReadByte()
		stream.log.Debug(fmt.Sprintf("ref id: %d", refId))
		offset++
	}

	for i := 0; i < numExt; i++ {
		payloadReader.Seek(int64(offset), io.SeekStart)
		typeAndSizeBytes := make([]byte, 2)
		binary.Read(payloadReader, binary.BigEndian, typeAndSizeBytes)

		typeAndSize := int(binary.BigEndian.Uint16(typeAndSizeBytes))
		dataType := typeAndSize >> 10
		dataSize := typeAndSize & 0x3FF

		if dataType == 1 {
			stream.log.Debug("ignored adaptive latency extension")
		} else {
			stream.log.Info("ignoring unknown extension type", "dataType", dataType, "dataSize", dataSize)
		}

		offset += dataSize + 2
	}

	stream.buffer = append(stream.buffer, packet.Payload[offset:]...)

	if packet.Marker {
		// send payload-specific feedback
		stream.log.Info("sending psfb")
		extReportBytes := stream.prepareExtendedReport(stream.ntpTime)
		psfbBytes := stream.preparePSFB()
		payload := append(extReportBytes, psfbBytes...)
		stream.sendRtcp(payload, addr)

		// decode frame
		stream.log.Info("decoding frame", "actualFrameId", stream.actualFrameId, "ckptFrameId", stream.ckptFrameId)
		stream.decode(stream.buffer, stream.actualFrameId)
		stream.actualFrameId++
		stream.buffer = make([]byte, 0)
		stream.ckptFrameId = frameId
	}
}

func (stream *Stream) handleRtcpPackets(packets []rtcp.Packet, addr net.Addr) {
	stream.log.Info("received rtcp packets", "len", len(packets))

	for _, packet := range packets {
		switch p := packet.(type) {
		case *rtcp.SenderReport:
			stream.log.Info("received sender report")

			// update sender timestamps
			stream.ntpTime = p.NTPTime
			stream.rtpTime = p.RTPTime

			// respond with a receiver report
			go func() {
				stream.log.Info("sending receiver report")
				extReportBytes := stream.prepareExtendedReport(stream.ntpTime)
				recvReportBytes := stream.prepareReceiverReport(stream.rtpTime)
				payload := append(extReportBytes, recvReportBytes...)
				stream.sendRtcp(payload, addr)
			}()

			break
		default:
			stream.log.Info("skipping rtcp packet")
			break
		}
	}
}

func (stream *Stream) preparePSFB() []byte {
	feedback := CastFeedback{
		ReceiverSSRC:        stream.receiverSsrc,
		SenderSSRC:          stream.senderSsrc,
		CkPtFrameId:         uint8(stream.ckptFrameId),
		LossFields:          0,
		CurrentPlayoutDelay: 400,
	}

	stream.log.Debug("psfb", "psfb", feedback)

	payload, err := feedback.Marshal()
	if err != nil {
		stream.log.Warn("failed to prepare psfb: " + err.Error())
		return nil
	}

	return payload
}

func (stream *Stream) prepareExtendedReport(time uint64) []byte {
	var reports []rtcp.ReportBlock

	reports = append(reports, &rtcp.ReceiverReferenceTimeReportBlock{
		NTPTimestamp: time,
	})

	report := rtcp.ExtendedReport{
		SenderSSRC: stream.receiverSsrc,
		Reports:    reports,
	}

	payload, err := report.Marshal()
	if err != nil {
		stream.log.Warn("failed to prepare ext report: " + err.Error())
		return nil
	}

	return payload
}

func (stream *Stream) prepareReceiverReport(rtpTime uint32) []byte {
	reports := make([]rtcp.ReceptionReport, 1)
	reports[0].SSRC = stream.senderSsrc
	reports[0].LastSenderReport = rtpTime
	reports[0].Delay = 200
	reports[0].LastSequenceNumber = uint32(stream.highestSeq)

	report := rtcp.ReceiverReport{
		SSRC:              stream.receiverSsrc,
		Reports:           reports,
		ProfileExtensions: nil,
	}

	payload, err := report.Marshal()
	if err != nil {
		stream.log.Warn("failed to prepare recv report: " + err.Error())
		return nil
	}

	return payload
}

func NewStream(decode func([]byte, int), log hclog.Logger, sendRtcp func([]byte, net.Addr), receiverSsrc uint32, senderSsrc uint32) *Stream {
	return &Stream{
		actualFrameId: 0,
		buffer:        make([]byte, 0),
		ckptFrameId:   255,
		decode:        decode,
		highestSeq:    0,
		log:           log,
		nextSeq:       -1,
		newFrameId:    -1,
		ntpTime:       0,
		packetsQueue:  make(map[uint16]*rtp.Packet),
		receiverSsrc:  receiverSsrc,
		rtpTime:       0,
		sendRtcp:      sendRtcp,
		senderSsrc:    senderSsrc,
	}
}
