package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	applogger "CraftFire/backend/internal/logger"
)

//go:embed all:frontend/dist
var assets embed.FS

// main 是 CraftFire 应用程序的入口点。
// 初始化 Wails 应用并绑定后端服务到前端。
func main() {
	// 初始化日志系统
	applogger.Init("INFO")
	applogger.Info("CraftFire 正在启动...")

	// 创建应用实例
	app := NewApp()

	// 创建 Wails 应用
	err := wails.Run(&options.App{
		Title:  "CraftFire",
		Width:  1280,
		Height: 720,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 11, G: 18, B: 32, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("启动 CraftFire 时发生错误: %v", err)
	}
}
