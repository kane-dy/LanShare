package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// 使用 Go 1.16+ 的 //go:embed 指令，将构建好的 Vue 3 前端静态资源打包进 Go 可执行文件内部
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 创建 App 实例（包含传输服务的状态与业务方法）
	app := NewApp()

	// 启动 Wails 桌面应用并配置运行时选项
	err := wails.Run(&options.App{
		Title:     "LanShare", // 软件窗口标题
		Width:     480,        // 窗口默认宽度
		Height:    580,        // 窗口默认高度
		MinWidth:  400,        // 允许调节的最小宽度
		MinHeight: 500,        // 允许调节的最小高度
		// 【核心】禁用窗口缩放
		DisableResize: true,

		// 配置前端静态资源服务器
		AssetServer: &assetserver.Options{
			Assets: assets,
		},

		// 应用启动时的生命周期回调函数
		OnStartup: app.startUp,

		// 核心绑定：在此处注册的结构体方法会自动映射到前端 window.go.main.App 中
		Bind: []interface{}{
			app,
		},
	})

	// 捕获应用运行期间的严重错误
	if err != nil {
		println("Wails 运行异常:", err.Error())
	}
}
