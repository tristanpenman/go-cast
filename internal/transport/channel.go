package transport

import (
	"encoding/binary"
	"io"
	"net"

	// third-party
	"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/proto"

	// internal
	"github.com/tristanpenman/go-cast/internal/channel"
)

type CastChannel struct {
	conn     net.Conn
	log      hclog.Logger
	Messages chan *channel.CastMessage
}

// NewCastChannel creates a framed Cast transport and starts its read loop.
func NewCastChannel(conn net.Conn, log hclog.Logger) CastChannel {
	messages := make(chan *channel.CastMessage, 64)

	go func() {
		for {
			lenBytes := make([]byte, 4)
			n, err := io.ReadFull(conn, lenBytes)
			if err != nil {
				log.Error("failed to read length", "err", err)
				break
			}

			if n != 4 {
				log.Error("failed to read length; too short")
				break
			}

			lenInt := binary.BigEndian.Uint32(lenBytes)
			log.Debug("message length", "bytes", lenInt)

			msgBytes := make([]byte, lenInt)
			n, err = io.ReadFull(conn, msgBytes)
			if err != nil {
				log.Error("failed to read message", "err", err)
				break
			}

			if uint32(n) != lenInt {
				log.Error("read unexpected number of bytes", "expected", lenInt, "actual", n)
				break
			}

			log.Debug("read message", "bytes", n)

			var castMessage channel.CastMessage
			err = proto.Unmarshal(msgBytes[:n], &castMessage)
			if err != nil {
				log.Error("failed to parse message", "err", err)
				break
			}

			log.Debug("received message", "namespace", *castMessage.Namespace)

			messages <- &castMessage
		}

		close(messages)
	}()

	return CastChannel{
		conn:     conn,
		log:      log,
		Messages: messages,
	}
}

// CreateCastChannel is retained for compatibility. New code should use
// NewCastChannel.
func CreateCastChannel(conn net.Conn, log hclog.Logger) CastChannel {
	return NewCastChannel(conn, log)
}

func (castChannel *CastChannel) Send(castMessage *channel.CastMessage) bool {
	msgBytes, err := proto.Marshal(castMessage)
	if err != nil {
		castChannel.log.Error("failed to encode binary cast message", "err", err)
		return false
	}

	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(msgBytes)))
	if err := writeFull(castChannel.conn, lenBytes); err != nil {
		castChannel.log.Error("failed to send cast message header", "err", err)
		return false
	}

	if err := writeFull(castChannel.conn, msgBytes); err != nil {
		castChannel.log.Error("failed to send cast message payload", "err", err)
		return false
	}

	return true
}

func writeFull(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := writer.Write(data)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}
