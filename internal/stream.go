package internal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	// third-party
	"github.com/hashicorp/go-hclog"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
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

type Stream struct {
	buffer         []byte
	currentFrameId int
	decode         func([]byte, int)
	highestSeq     int
	log            hclog.Logger
	nextSeq        int
	newFrameId     int
	packetsQueue   map[uint16]*rtp.Packet
	prevFrameId    int
}

func (stream *Stream) enqueuePacket(packet *rtp.Packet) {
	stream.packetsQueue[packet.SequenceNumber] = packet
	stream.log.Info("enqueued packet", "sequenceNumber", packet.SequenceNumber)
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

func (stream *Stream) handleDataPacket(packet *rtp.Packet) {
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

func (stream *Stream) handleRtcpPackets(packets []rtcp.Packet) {
	stream.log.Info("received rtcp packets", "len", len(packets))

	for _, packet := range packets {
		switch packet.(type) {
		case *rtcp.CompoundPacket:
			break
		case *rtcp.PictureLossIndication:
			break
		default:
			break
		}
	}

	// make a fake sender report

}

func NewStream(decode func([]byte, int), log hclog.Logger) *Stream {
	return &Stream{
		buffer:         make([]byte, 0),
		currentFrameId: -1,
		decode:         decode,
		highestSeq:     -1,
		log:            log,
		nextSeq:        -1,
		newFrameId:     -1,
		packetsQueue:   make(map[uint16]*rtp.Packet),
		prevFrameId:    0,
	}
}
