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
	buffer []byte
	//currentFrameId int
	decode       func([]byte, int)
	highestSeq   uint16
	log          hclog.Logger
	nextSeq      int
	newFrameId   int
	packetsQueue map[uint16]*rtp.Packet
	prevFrameId  int
	receiverSsrc uint32
	sendRtcp     func([]byte, net.Addr)
	senderSsrc   uint32
}

func (stream *Stream) enqueuePacket(packet *rtp.Packet) {
	stream.packetsQueue[packet.SequenceNumber] = packet
	stream.log.Info("enqueued packet", "sequenceNumber", packet.SequenceNumber)

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
			stream.log.Info("dequeueing packet", "sequenceNumber", seq)
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
		"prevFrameId", stream.prevFrameId,
		"packetId", packetId,
		"maxPacketId", maxPacketId)

	if frameId != stream.prevFrameId {
		stream.decode(stream.buffer, frameId)
		stream.buffer = make([]byte, 0)
		stream.sendPSFB(addr)
		stream.prevFrameId = frameId
	}

	offset := 6
	if hasRef {
		payloadReader.Seek(int64(offset), io.SeekStart)
		refId, _ := payloadReader.ReadByte()
		stream.log.Info(fmt.Sprintf("ref id: %d", refId))
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
			stream.log.Info("ignored adaptive latency extension")
		} else {
			stream.log.Info("ignoring unknown extension type", "dataType", dataType, "dataSize", dataSize)
		}

		offset += dataSize + 2
	}

	stream.buffer = append(stream.buffer, packet.Payload[offset:]...)
}

func (stream *Stream) handleRtcpPackets(packets []rtcp.Packet, addr net.Addr) {
	stream.log.Info("received rtcp packets", "len", len(packets))

	for _, packet := range packets {
		switch p := packet.(type) {
		case *rtcp.SenderReport:
			stream.log.Info("received sender report")
			stream.sendExtendedReport(addr, p.NTPTime)
			stream.sendReceiverReport(addr, p.NTPTime)
			break
		default:
			stream.log.Info("skipping rtcp packet")
			break
		}
	}
}

func (stream *Stream) sendPSFB(addr net.Addr) {
	feedback := CastFeedback{
		ReceiverSSRC:        stream.receiverSsrc,
		SenderSSRC:          stream.senderSsrc,
		CkPtFrameId:         uint8(stream.prevFrameId),
		LossFields:          0,
		CurrentPlayoutDelay: 0,
	}

	stream.log.Info("psfb", "psfb", feedback)

	bytes, _ := feedback.Marshal()

	stream.sendRtcp(bytes, addr)
}

func (stream *Stream) sendExtendedReport(addr net.Addr, time uint64) {
	var reports []rtcp.ReportBlock

	reports = append(reports, &rtcp.ReceiverReferenceTimeReportBlock{
		XRHeader: rtcp.XRHeader{
			BlockType:    rtcp.ReceiverReferenceTimeReportBlockType,
			TypeSpecific: 0,
			BlockLength:  2,
		},
		NTPTimestamp: time,
	})

	report := rtcp.ExtendedReport{
		SenderSSRC: stream.receiverSsrc,
		Reports:    reports,
	}

	bytes, err := report.Marshal()
	if err != nil {
		return
	}

	stream.sendRtcp(bytes, addr)
}

func (stream *Stream) sendReceiverReport(addr net.Addr, time uint64) {
	reports := make([]rtcp.ReceptionReport, 1)
	reports[0].SSRC = stream.senderSsrc
	reports[0].LastSenderReport = ExtractMiddleBits(time)
	reports[0].LastSequenceNumber = uint32(stream.highestSeq)

	report := rtcp.ReceiverReport{
		SSRC:              stream.receiverSsrc,
		Reports:           reports,
		ProfileExtensions: nil,
	}

	bytes, err := report.Marshal()
	if err != nil {
		return
	}

	stream.sendRtcp(bytes, addr)
}

func NewStream(decode func([]byte, int), log hclog.Logger, sendRtcp func([]byte, net.Addr), receiverSsrc uint32, senderSsrc uint32) *Stream {
	return &Stream{
		buffer: make([]byte, 0),
		//currentFrameId: 0,
		decode:       decode,
		highestSeq:   0,
		log:          log,
		nextSeq:      -1,
		newFrameId:   -1,
		packetsQueue: make(map[uint16]*rtp.Packet),
		prevFrameId:  0,
		receiverSsrc: receiverSsrc,
		sendRtcp:     sendRtcp,
		senderSsrc:   senderSsrc,
	}
}
