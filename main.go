//go:build !desktop

package main

import (
	"fmt"
	"time"
)

// main Web 版入口：直接运行 HTTP 服务，通过浏览器访问
func main() {
	// 启动时确保数据目录存在：Documents/FNSoftware/FireShare/uploads
	if err := ensureDataDir(); err != nil {
		fmt.Printf("创建数据目录失败: %v\n", err)
		return
	}

	mux := newMux()

	// 尝试从默认端口开始启动服务器
	port, err := findAvailablePort(true)
	if err != nil {
		fmt.Println(err)
		return
	}

	listener, errCh, err := startServer(mux, port)
	if err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
		return
	}
	defer listener.Close()

	status := refreshNetworkStatus(port)

	fmt.Printf("服务器启动成功！\n")
	fmt.Printf("数据目录: %s\n", dataDir)
	fmt.Printf("文件存储目录: %s\n", uploadDir)
	printNetworkStatus(status)
	fmt.Printf("启动时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// 等待服务器退出（正常运行时会一直阻塞）
	if err := <-errCh; err != nil {
		fmt.Printf("服务器运行失败: %v\n", err)
	}
}
