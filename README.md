# 📁 局域网文件传输助手

一个简单高效的基于 Go 的局域网文件传输工具，支持文件上传、下载和列表管理。

项目提供两种使用形态，共享同一套代码与界面：

- 🌐 **Web 版**：启动 HTTP 服务，用任意浏览器（含手机）访问
- 🖥️ **桌面版**：原生桌面窗口（Windows / macOS / Linux），同时保留局域网 HTTP 服务供其他设备访问

## ✨ 功能特性

- 📤 **文件上传** - 支持拖拽上传和点击选择，可批量上传
- 📥 **文件下载** - 一键下载任意文件
- 📋 **文件列表** - 实时显示服务器上的所有文件
- 🗑️ **文件删除** - 支持在线删除不需要的文件
- 🎨 **美观界面** - 现代化的响应式设计，支持移动端
- 🌐 **局域网传输** - 在同一局域网内快速传输文件
- ⚡ **高速传输** - 基于原生 Go 实现，传输速度快
- 🖥️ **跨平台桌面版** - 基于 Wails + 系统 WebView，Windows / macOS / Linux 均可构建

## 🚀 快速开始

### 环境要求

- Go 1.22 或更高版本
- 现代浏览器（Chrome、Firefox、Edge 等）

桌面版额外要求：

| 平台 | 依赖 |
| --- | --- |
| Windows | WebView2 运行时（Win10/11 自带，无需 CGO） |
| macOS | Xcode Command Line Tools（CGO） |
| Linux | gcc、`libgtk-3-dev`、`libwebkit2gtk-4.0-dev`（CGO） |

### 运行 Web 版

```bash
go run .
```

启动后访问：

- 本机访问：`http://localhost:3000/`
- 局域网访问：`http://你的IP地址:3000/`（端口被占用时自动递增）

### 运行桌面版

**Windows：**

```bash
go build -tags "desktop production" -ldflags "-H windowsgui" -o GoFile-Desktop.exe .
GoFile-Desktop.exe
```

**macOS / Linux：**

```bash
go build -tags "desktop production" -o GoFile-Desktop .
./GoFile-Desktop
```

桌面版启动后会打开原生窗口，同时后台仍会运行局域网 HTTP 服务，其他设备依然可以通过浏览器上传/下载文件。

## 🔨 构建脚本

项目内置了两个构建脚本，可一次构建 Web 版和桌面版：

- **Windows**：`.\build.ps1`
- **macOS / Linux**：`./build.sh`

构建产物：

```text
GoFile.exe / GoFile          # Web 版（终端运行，输出访问地址）
GoFile-Desktop.exe / GoFile-Desktop  # 桌面版（原生窗口）
```

## 📖 使用说明

### 上传文件

1. **点击上传** - 点击"选择文件"按钮，选择要上传的文件
2. **拖拽上传** - 直接将文件拖拽到上传区域
3. 支持同时选择多个文件进行上传

### 下载文件

在文件列表中找到要下载的文件，点击"⬇️ 下载"按钮即可。

### 删除文件

点击文件右侧的"🗑️ 删除"按钮，确认后即可删除文件。

### 刷新列表

点击"🔄 刷新"按钮重新加载文件列表。

## 📂 项目结构

```text
FileShare.Go/
├── main.go           # Web 版入口（!desktop 构建标签）
├── desktop.go        # 桌面版入口（desktop 构建标签，基于 Wails）
├── server.go         # 共享的 HTTP 服务与全部 API 处理逻辑
├── static/           # 静态文件目录
│   └── index.html    # 前端页面（Web 版与桌面版共用）
├── uploads/          # 上传文件存储目录（自动创建）
├── build.ps1         # Windows 构建脚本
├── build.sh          # macOS / Linux 构建脚本
└── README.md         # 项目说明文档
```

## ⚙️ 配置说明

### 修改端口

编辑 [server.go](server.go) 文件：

```go
const defaultPort = 3000  // 修改为你想要的端口
```

### 修改上传目录

编辑 [server.go](server.go) 文件：

```go
const uploadDir = "./uploads"  // 修改为你想要的目录
```

### 修改文件大小限制

编辑 [server.go](server.go) 中的上传处理函数：

```go
err := r.ParseMultipartForm(100 << 20)  // 100 << 20 表示 100MB
```

## 🛡️ 安全注意事项

1. **局域网使用** - 本工具设计用于局域网文件传输，不建议直接暴露到公网
2. **文件验证** - 当前版本不包含文件类型验证，请谨慎上传和下载文件
3. **数据备份** - 上传的文件存储在 `uploads` 目录，请定期备份重要文件
4. **访问控制** - 如需添加访问控制，请自行实现身份验证功能

## 🎣 API 接口

### 获取文件列表

```text
GET /api/files
```

### 上传文件

```text
POST /api/upload
Content-Type: multipart/form-data
```

### 下载文件

```text
GET /api/download/{filename}
```

### 删除文件

```text
DELETE /api/delete/{filename}
```

## 💻 技术栈

- **后端**：Go 标准库 `net/http`
- **前端**：原生 HTML + CSS + JavaScript
- **桌面壳**：[Wails v2](https://wails.io) + 系统 WebView（WebView2 / WKWebView / WebKitGTK）
- **样式**：自定义 CSS3（无第三方依赖）

## ❓ 常见问题

**Q: 无法访问服务？**

- 检查防火墙设置
- 确认端口是否被占用（程序会自动尝试下一个端口）
- 查看本机 IP 地址是否正确

**Q: 上传大文件失败？**

- 修改 [server.go](server.go) 中的文件大小限制
- 检查磁盘空间是否充足

**Q: 桌面版启动报错？**

- Windows：确认已安装 WebView2 运行时（Win10/11 通常自带）
- Linux：确认已安装 `libwebkit2gtk-4.0-dev` 等依赖，并使用 `-tags "desktop production"` 构建
- macOS：确认已安装 Xcode Command Line Tools

**Q: 中文文件名乱码？**

- 确保浏览器编码设置为 UTF-8
- 本程序已对中文文件名进行优化

## 📜 许可

MIT License

**享受高速的局域网文件传输体验！🚀**
