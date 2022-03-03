package cast

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-hclog"
)

type CastChannel struct {
	conn     net.Conn
	logger   hclog.Logger
	Messages chan *CastMessage
}

func CreateCastChannel(conn net.Conn, logger hclog.Logger) CastChannel {
	messages := make(chan *CastMessage)

	go func() {
		for {
			lenBytes := make([]byte, 4)
			n, err := conn.Read(lenBytes)
			if err != nil {
				logger.Error("failed to read length", "err", err)
				break
			}

			if n != 4 {
				logger.Error("failed to read length; too short")
				break
			}

			lenInt := binary.BigEndian.Uint32(lenBytes)
			if logger.IsDebug() {
				logger.Debug(fmt.Sprintf("Message length: %d", lenInt))
			}

			// TODO: Make this handle split header and body packets properly
			msgBytes := make([]byte, lenInt)
			n, err = conn.Read(msgBytes)
			if err != nil {
				logger.Error("failed to read message", "err", err)
				break
			}

			if uint32(n) != lenInt {
				logger.Error("read unexpected number of bytes", "expected", lenInt, "actual", n)
				break
			}

			if logger.IsDebug() {
				logger.Debug(fmt.Sprintf("Read: %d", n))
			}

			var castMessage CastMessage
			err = proto.Unmarshal(msgBytes[:n], &castMessage)
			if err != nil {
				logger.Error("failed to parse message", "err", err)
				break
			}

			if logger.IsDebug() {
				logger.Debug("Received message", "namespace", *castMessage.Namespace)
			}

			messages <- &castMessage
		}

		close(messages)
	}()

	return CastChannel{
		conn:     conn,
		logger:   logger,
		Messages: messages,
	}
}

func (castChannel *CastChannel) Send(castMessage *CastMessage) bool {
	msgBytes, err := proto.Marshal(castMessage)
	if err != nil {
		castChannel.logger.Error("failed to encode binary cast message", "err", err)
		return false
	}

	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(msgBytes)))
	if _, err := castChannel.conn.Write(lenBytes); err != nil {
		castChannel.logger.Error("failed to send cast message header", "err", err)
		return false
	}

	if _, err := castChannel.conn.Write(msgBytes); err != nil {
		castChannel.logger.Error("failed to send cast message payload", "err", err)
		return false
	}

	return true
}
