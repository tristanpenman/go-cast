package internal

import (
	"github.com/hashicorp/go-hclog"
)

var Logger = hclog.New(&hclog.LoggerOptions{
	Name:  "go-cast",
	Level: hclog.LevelFromString("INFO"),
})
