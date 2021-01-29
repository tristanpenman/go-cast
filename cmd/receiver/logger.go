package main

import (
	"github.com/hashicorp/go-hclog"
)

var logger = hclog.New(&hclog.LoggerOptions{
	Name:  "receiver",
	Level: hclog.LevelFromString("INFO"),
})
