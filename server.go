package main

import (
	"embed"
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

//go:embed static/index.html
var staticFiles embed.FS

const (
	// 用户数据：文档目录 / FNSoftware（提供商） / FileShare（程序） / uploads
	providerName = "FNSoftware"
	appName      = "FileShare"
	uploadsName  = "uploads"

	defaultPort = 3000
	// maxPortScan 端口扫描范围
	maxPortScan = 100
)

// uploadDir 上传文件实际存储路径，在 ensureUploadDir 中初始化
var uploadDir string

// FileInfo 文件信息结构
type FileInfo struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
}

// newMux 注册所有路由，Web 版与桌面版共用同一套处理逻辑
func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	// 设置路由
	mux.HandleFunc("/", indexHandler)

	// API 路由
	mux.HandleFunc("/api/files", listFilesHandler)
	mux.HandleFunc("/api/upload", uploadHandler)
	mux.HandleFunc("/api/download/", downloadHandler)
	mux.HandleFunc("/api/delete/", deleteHandler)

	return mux
}

// userDocumentsDir 返回当前用户的「文档」目录
func userDocumentsDir() (string, error) {
	if xdg := os.Getenv("XDG_DOCUMENTS_DIR"); xdg != "" {
		return xdg, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Documents"), nil
}

// resolveUploadDir 解析为 Documents/FNSoftware/FileShare/uploads
func resolveUploadDir() (string, error) {
	docs, err := userDocumentsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(docs, providerName, appName, uploadsName), nil
}

// ensureUploadDir 解析用户数据路径；若 FNSoftware / FileShare / uploads 不存在则自动创建
func ensureUploadDir() error {
	dir, err := resolveUploadDir()
	if err != nil {
		return fmt.Errorf("解析用户数据目录失败: %w", err)
	}
	uploadDir = dir
	// MkdirAll 会创建整条路径中缺失的每一级目录
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return fmt.Errorf("创建上传目录失败 (%s): %w", uploadDir, err)
	}
	return nil
}

// findAvailablePort 尝试从默认端口开始查找可用端口
func findAvailablePort(verbose bool) (int, error) {
	for port := defaultPort; port <= defaultPort+maxPortScan; port++ {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			listener.Close()
			return port, nil
		}
		if verbose {
			fmt.Printf("端口 %d 被占用，尝试端口 %d...\n", port, port+1)
		}
	}
	return 0, fmt.Errorf("无法找到可用端口（已尝试 %d-%d）", defaultPort, defaultPort+maxPortScan)
}

// startServer 在指定端口启动 HTTP 服务并立即返回，服务在后台运行。
// 返回的 error 通道会在服务异常退出时收到错误，正常关闭时通道被关闭。
func startServer(mux http.Handler, port int) (net.Listener, <-chan error, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, nil, err
	}

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		err := srv.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	return listener, errCh, nil
}

// indexHandler 处理首页请求
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// 只处理根路径
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 读取内嵌的 HTML 文件
	content, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "页面加载失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
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

	if err := ensureUploadDir(); err != nil {
		http.Error(w, "创建上传目录失败", http.StatusInternalServerError)
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

	// 上传前再次确保目录存在（防止运行中被删除）
	if err := ensureUploadDir(); err != nil {
		http.Error(w, "创建上传目录失败", http.StatusInternalServerError)
		return
	}

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
