package transport

import (
	// internal
	"github.com/tristanpenman/go-cast/internal/channel"
)

type CastTransport interface {
	HandleCastMessage(message *channel.CastMessage)
	TransportID() string
}
