//go:build desktop

package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// main 桌面版入口：内嵌 WebView 窗口 + 后台局域网 HTTP 服务
func main() {
	// 启动时确保数据目录存在：Documents/FNSoftware/FireShare/uploads
	if err := ensureDataDir(); err != nil {
		fmt.Printf("创建数据目录失败: %v\n", err)
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

	status := refreshNetworkStatus(port)

	fmt.Printf("桌面版已启动\n")
	fmt.Printf("数据目录: %s\n", dataDir)
	fmt.Printf("文件存储目录: %s\n", uploadDir)
	printNetworkStatus(status)

	// 重要：不要把 API 挂在 Wails AssetServer 上。
	// WebView2 通过 AssetServer 转发 multipart 上传时会丢失文件二进制内容，导致保存为 0KB。
	// 桌面窗口直接跳转到本机 HTTP 服务，与浏览器走同一完整请求路径。
	err = wails.Run(&options.App{
		Title:                    "FireShare",
		Width:                    1100,
		Height:                   760,
		MinWidth:                 800,
		MinHeight:                600,
		BackgroundColour:         options.NewRGB(245, 247, 250),
		EnableDefaultContextMenu: true,
		AssetServer: &assetserver.Options{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				target := fmt.Sprintf("http://127.0.0.1:%d%s", port, r.URL.RequestURI())
				http.Redirect(w, r, target, http.StatusFound)
			}),
		},
		OnShutdown: func(ctx context.Context) {
			_ = listener.Close()
		},
	})
	if err != nil {
		fmt.Printf("桌面版启动失败: %v\n", err)
		_ = listener.Close()
	}
}
