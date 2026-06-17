package internal

import (
	// third-party
	"github.com/hashicorp/go-hclog"
)

func NewLogger(name string) hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Color: hclog.AutoColor,
		Level: hclog.Info,
		Name:  name,
	})
}
