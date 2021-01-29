package main

import (
	"github.com/hashicorp/go-hclog"
)

var logger = hclog.New(&hclog.LoggerOptions{
	Name:  "sender",
	Level: hclog.LevelFromString("INFO"),
})
