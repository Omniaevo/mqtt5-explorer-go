package main

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

//go:embed wails.json
var WailsJSON []byte

func main() {
	app := NewApp()
	name := "MQTT5 Explorer"

	var config struct {
    Info struct {
      ProductVersion string `json:"productVersion"`
    } `json:"info"`
  }
  json.Unmarshal(WailsJSON, &config)

  version := config.Info.ProductVersion

	err := wails.Run(&options.App{
		Title:     name,
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
		  Icon: icon,
			ProgramName: name,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			About: &mac.AboutInfo{
        Title:   name,
        Message: fmt.Sprintf("Version: %s", version),
        Icon:    icon,
      },
		},
		Windows: &windows.Options{
		  WindowClassName: name,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
