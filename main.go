// In go-static-cms/main.go
package main

import (
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "Go Static CMS",
		Width:  1280,
		Height: 800,
		// Since we are using a live web server, we don't use the AssetServer.
		// We tell Wails to load its content from our internal server's URL.
		StartHidden: false, // Start the window visible
		Bind: []interface{}{
			app,
		},
		AssetServer: &assetserver.Options{
			Handler: createWebServerMux(app),
		},
		LogLevel: logger.DEBUG,
		Logger:   logger.NewDefaultLogger(),
	})

	if err != nil {
		log.Fatal(err)
	}
}
