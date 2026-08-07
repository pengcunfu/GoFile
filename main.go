//go:build !desktop

package main

import (
	"fmt"
	"time"
)

// main Web 版入口：直接运行 HTTP 服务，通过浏览器访问
func main() {
	// 确保上传目录存在
	if err := ensureUploadDir(); err != nil {
		fmt.Printf("创建上传目录失败: %v\n", err)
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

	// 获取局域网 IP 地址
	lanIP := getLANIP()

	fmt.Printf("服务器启动成功！\n")
	fmt.Printf("文件存储目录: %s\n", uploadDir)
	fmt.Printf("本地访问: http://localhost:%d/\n", port)
	if lanIP != "" {
		fmt.Printf("局域网访问: http://%s:%d/\n", lanIP, port)
	}
	fmt.Printf("启动时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// 等待服务器退出（正常运行时会一直阻塞）
	if err := <-errCh; err != nil {
		fmt.Printf("服务器运行失败: %v\n", err)
	}
}
