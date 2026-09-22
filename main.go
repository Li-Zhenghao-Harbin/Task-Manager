package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(appOptions(app))

	if err != nil {
		println("Error:", err.Error())
	}
}

// appOptions holds the window configuration. It is a function rather than an inline literal so the
// startup behaviour can be pinned down by a test.
//
// WindowStartState: options.Maximised makes the window appear already maximised: Wails creates the
// window hidden and calls ShowWindow(SW_MAXIMIZE) on it, so the first frame the user sees is the
// maximised one. Maximising *later* cannot avoid the visible transition — the OnDomReady +
// runtime.WindowMaximise approach used before always showed the window at Width x Height first and
// then grew it, which is the animation this replaced. Width/Height stay as the size the window
// falls back to when it is un-maximised.
func appOptions(app *App) *options.App {
	return &options.App{
		Title:            "Task Manager",
		Width:            1280,
		Height:           800,
		MinWidth:         1024,
		MinHeight:        680,
		WindowStartState: options.Maximised,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 244, G: 248, B: 255, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	}
}
