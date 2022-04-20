package internal

import (
	"github.com/tristanpenman/go-cast/internal/channel"
)

type CastTransport interface {
	HandleCastMessage(message *channel.CastMessage)
	TransportId() string
}
