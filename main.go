package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"case/internal/app"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	appService := app.New()

	wailsApp := application.New(application.Options{
		Name:        "case",
		Description: "Desktop media organizer powered by Stash",
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	appService.SetApplication(wailsApp)

	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "case",
		Width:  1280,
		Height: 800,
		URL:    "/",
	})

	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}
