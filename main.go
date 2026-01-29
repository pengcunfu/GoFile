package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	uploadDir   = "./uploads"
	defaultPort = 3000
)

// FileInfo 文件信息结构
type FileInfo struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
}

func main() {
	// 确保上传目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		fmt.Printf("创建上传目录失败: %v\n", err)
		return
	}

	// 设置静态文件服务
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// API 路由
	http.HandleFunc("/api/files", listFilesHandler)
	http.HandleFunc("/api/upload", uploadHandler)
	http.HandleFunc("/api/download/", downloadHandler)
	http.HandleFunc("/api/delete/", deleteHandler)

	// 获取局域网 IP 地址
	lanIP := getLANIP()

	// 尝试从默认端口开始启动服务器
	port := defaultPort
	for {
		addr := fmt.Sprintf(":%d", port)

		// 尝试启动服务器
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			// 成功绑定端口
			listener.Close()

			fmt.Printf("服务器启动成功！\n")
			fmt.Printf("文件存储目录: %s\n", uploadDir)
			fmt.Printf("本地访问: http://localhost:%d\n", port)
			if lanIP != "" {
				fmt.Printf("局域网访问: http://%s:%d\n", lanIP, port)
			}
			fmt.Printf("启动时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

			// 启动 HTTP 服务器
			if err := http.ListenAndServe(addr, nil); err != nil {
				fmt.Printf("服务器运行失败: %v\n", err)
			}
			return
		}

		// 端口被占用，尝试下一个端口
		fmt.Printf("端口 %d 被占用，尝试端口 %d...\n", port, port+1)
		port++
		if port > defaultPort+100 {
			fmt.Printf("无法找到可用端口（已尝试 %d-%d）\n", defaultPort, port)
			return
		}
	}
}

// getLANIP 获取局域网 IP 地址
func getLANIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		// 检查 IP 地址类型
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()
				// 过滤掉本机回环地址和链路本地地址
				if !strings.HasPrefix(ip, "127.") && !strings.HasPrefix(ip, "169.254.") {
					return ip
				}
			}
		}
	}
	return ""
}

// listFilesHandler 处理文件列表请求
func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	files, err := os.ReadDir(uploadDir)
	if err != nil {
		http.Error(w, "读取文件列表失败", http.StatusInternalServerError)
		return
	}

	// 初始化为空切片，避免返回 null
	fileList := make([]FileInfo, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		fileList = append(fileList, FileInfo{
			Name:     file.Name(),
			Size:     info.Size(),
			Modified: info.ModTime().Unix(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileList)
}

// uploadHandler 处理文件上传
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析表单，最大 100MB
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "解析表单失败", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "获取文件失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 创建目标文件
	filename := filepath.Join(uploadDir, handler.Filename)
	dst, err := os.Create(filename)
	if err != nil {
		http.Error(w, "创建文件失败", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "保存文件失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "文件上传成功")
}

// downloadHandler 处理文件下载
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 中提取文件名
	filename := strings.TrimPrefix(r.URL.Path, "/api/download/")
	if filename == "" {
		http.Error(w, "文件名不能为空", http.StatusBadRequest)
		return
	}

	filepath := filepath.Join(uploadDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, filepath)
}

// deleteHandler 处理文件删除
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 中提取文件名
	filename := strings.TrimPrefix(r.URL.Path, "/api/delete/")
	if filename == "" {
		http.Error(w, "文件名不能为空", http.StatusBadRequest)
		return
	}

	filepath := filepath.Join(uploadDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 删除文件
	if err := os.Remove(filepath); err != nil {
		http.Error(w, "删除文件失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "文件删除成功")
}
