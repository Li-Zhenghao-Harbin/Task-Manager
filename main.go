package main

import (
	"context"
	"embed"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "TaskManager",
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 680,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 244, G: 248, B: 255, A: 1},
		Logger:           logger.NewFileLogger("TaskManager-wails.log"),
		LogLevel:         logger.DEBUG,
		LogLevelProduction: logger.DEBUG,
		OnStartup:        app.startup,
		OnDomReady: func(ctx context.Context) {
			go func() {
				time.Sleep(250 * time.Millisecond)
				runtime.WindowMaximise(ctx)
			}()
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
