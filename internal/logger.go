package internal

import (
	"github.com/hashicorp/go-hclog"
)

func NewLogger(name string) hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Name:  name,
		Level: hclog.Info,
	})
}
