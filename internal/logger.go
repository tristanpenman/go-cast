package internal

import (
	"github.com/hashicorp/go-hclog"
)

func NewLogger(name string) hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Color: hclog.AutoColor,
		Level: hclog.Info,
		Name:  name,
	})
}
