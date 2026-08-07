//go:build desktop

package main

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// main 桌面版入口：内嵌 WebView 窗口 + 后台局域网 HTTP 服务
func main() {
	// 确保上传目录存在
	if err := ensureUploadDir(); err != nil {
		fmt.Printf("创建上传目录失败: %v\n", err)
		return
	}

	// 注册路由：桌面窗口与局域网 HTTP 服务共用同一套处理逻辑
	mux := newMux()

	// 启动局域网 HTTP 服务，其他设备仍可通过浏览器访问
	port, err := findAvailablePort(false)
	if err != nil {
		fmt.Println(err)
		return
	}
	listener, _, err := startServer(mux, port)
	if err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
		return
	}

	// 获取局域网 IP 地址
	lanIP := getLANIP()
	fmt.Printf("桌面版已启动\n")
	fmt.Printf("文件存储目录: %s\n", uploadDir)
	fmt.Printf("本机浏览器访问: http://localhost:%d/\n", port)
	if lanIP != "" {
		fmt.Printf("局域网访问: http://%s:%d/\n", lanIP, port)
	}

	// 桌面窗口内嵌静态页面（与 Web 版同一份 index.html）
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		fmt.Printf("加载页面资源失败: %v\n", err)
		listener.Close()
		return
	}

	err = wails.Run(&options.App{
		Title:                    "局域网文件传输助手",
		Width:                    1100,
		Height:                   760,
		MinWidth:                 800,
		MinHeight:                600,
		BackgroundColour:         options.NewRGB(245, 247, 250),
		EnableDefaultContextMenu: true,
		AssetServer: &assetserver.Options{
			Assets:  staticFS,
			Handler: mux,
		},
		OnShutdown: func(ctx context.Context) {
			_ = listener.Close()
		},
	})
	if err != nil {
		fmt.Printf("桌面版启动失败: %v\n", err)
	}
}
