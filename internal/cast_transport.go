package internal

import "github.com/tristanpenman/go-cast/internal/cast"

type CastTransport interface {
	HandleCastMessage(message *cast.CastMessage)
	TransportId() string
}
