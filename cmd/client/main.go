package main

import (
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	err := wails.Run(&options.App{
		Title:  "GoCast",
		Width:  920,
		Height: 680,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind:       []interface{}{app},
		OnShutdown: app.shutdown,
	})
	if err != nil {
		fmt.Println("Error:", err)
	}
}
