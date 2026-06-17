package internal

import (
	"encoding/binary"
	"errors"
	"fmt"

	// third-party
	"github.com/pion/rtcp"
)

// Cast Feedback Message:
//
//  0                   1                   2                   3
//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                       SSRC of Receiver                        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                        SSRC of Sender                         |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |               Unique identifier 'C' 'A' 'S' 'T'               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | CkPt Frame ID | # Loss Fields | Current Playout Delay (msec)  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

type CastFeedback struct {
	ReceiverSSRC        uint32
	SenderSSRC          uint32
	CkPtFrameId         uint8
	LossFields          uint8
	CurrentPlayoutDelay uint16
}

const (
	headerLength = 4
	bodyLength   = 4 // TODO: loss increases size
	ssrcLength   = 4
)

func (feedback *CastFeedback) Marshal() ([]byte, error) {
	rawPacket := make([]byte, feedback.len())
	packetBody := rawPacket[headerLength:]

	binary.BigEndian.PutUint32(packetBody, feedback.ReceiverSSRC)
	binary.BigEndian.PutUint32(packetBody[ssrcLength:], feedback.SenderSSRC)

	packetBody[ssrcLength*2+0] = 'C'
	packetBody[ssrcLength*2+1] = 'A'
	packetBody[ssrcLength*2+2] = 'S'
	packetBody[ssrcLength*2+3] = 'T'

	packetBody[ssrcLength*2+4] = feedback.CkPtFrameId
	packetBody[ssrcLength*2+5] = feedback.LossFields

	binary.BigEndian.PutUint16(packetBody[ssrcLength*2+6:], feedback.CurrentPlayoutDelay)

	h := feedback.Header()
	hData, err := h.Marshal()
	if err != nil {
		return nil, err
	}
	copy(rawPacket, hData)

	return rawPacket, nil
}

func (feedback *CastFeedback) Unmarshal(rawPacket []byte) error {
	if len(rawPacket) < (headerLength + bodyLength) {
		return errors.New("rtcp: packet too short")
	}

	var h rtcp.Header
	if err := h.Unmarshal(rawPacket); err != nil {
		return err
	}

	if h.Type != rtcp.TypePayloadSpecificFeedback || h.Count != rtcp.FormatREMB {
		return errors.New("rtcp: wrong packet type")
	}

	feedback.ReceiverSSRC = binary.BigEndian.Uint32(rawPacket[headerLength:])
	feedback.SenderSSRC = binary.BigEndian.Uint32(rawPacket[headerLength+ssrcLength:])

	return nil
}

func (feedback *CastFeedback) Header() rtcp.Header {
	return rtcp.Header{
		Count:  rtcp.FormatREMB,
		Type:   rtcp.TypePayloadSpecificFeedback,
		Length: bodyLength,
	}
}

func (feedback *CastFeedback) len() int {
	return headerLength + 16
}

func (feedback *CastFeedback) String() string {
	return fmt.Sprintf("CastFeedback %x %x", feedback.ReceiverSSRC, feedback.SenderSSRC)
}
