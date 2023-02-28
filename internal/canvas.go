package internal

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/hashicorp/go-hclog"
)

type Canvas struct {
	log    hclog.Logger
	window *glfw.Window
}

func (canvas *Canvas) Start() {

}

func NewCanvas() *Canvas {
	return &Canvas{
		log: NewLogger("canvas"),
	}
}
