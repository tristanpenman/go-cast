package internal

import (
	// internal
	"github.com/tristanpenman/go-cast/pkg/channel"
)

type CastTransport interface {
	HandleCastMessage(message *channel.CastMessage)
	TransportId() string
}
