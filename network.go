package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

const cloudPortTip = "若运行在云服务器（阿里云 / 腾讯云 / AWS 等），除本机防火墙外，还需在云厂商控制台的安全组 / 防火墙中放行该 TCP 端口，外网才能访问。"

// listenPort 当前 HTTP 服务监听端口，启动后写入
var listenPort int

var (
	networkStatusMu sync.RWMutex
	networkStatus   NetworkStatus
)

// NetworkStatus 网络与防火墙检测结果
type NetworkStatus struct {
	Port            int    `json:"port"`
	LocalURL        string `json:"localUrl"`
	LanIP           string `json:"lanIp"`
	LanURL          string `json:"lanUrl"`
	FirewallStatus  string `json:"firewallStatus"`  // checking | ok | may_block | unknown
	FirewallMessage string `json:"firewallMessage"`
	CloudTip        string `json:"cloudTip"`
}

// refreshNetworkStatus 立即填充访问地址，并异步检测防火墙（不阻塞启动）
func refreshNetworkStatus(port int) NetworkStatus {
	listenPort = port
	lanIP := getLANIP()

	status := NetworkStatus{
		Port:            port,
		LocalURL:        fmt.Sprintf("http://localhost:%d/", port),
		LanIP:           lanIP,
		CloudTip:        cloudPortTip,
		FirewallStatus:  "checking",
		FirewallMessage: fmt.Sprintf("正在检测 TCP %d 是否可能被本机防火墙拦截…", port),
	}
	if lanIP != "" {
		status.LanURL = fmt.Sprintf("http://%s:%d/", lanIP, port)
	}

	networkStatusMu.Lock()
	networkStatus = status
	networkStatusMu.Unlock()

	go func(p int) {
		fw := checkFirewallPort(p)
		networkStatusMu.Lock()
		if networkStatus.Port == p {
			networkStatus.FirewallStatus = fw.Status
			networkStatus.FirewallMessage = fw.Message
		}
		networkStatusMu.Unlock()
		fmt.Printf("防火墙检测: %s\n", fw.Message)
	}(port)

	return status
}

func getNetworkStatus() NetworkStatus {
	networkStatusMu.RLock()
	defer networkStatusMu.RUnlock()
	return networkStatus
}

// printNetworkStatus 在控制台输出访问地址与云提示（防火墙结果稍后异步打印）
func printNetworkStatus(status NetworkStatus) {
	fmt.Printf("本地访问: %s\n", status.LocalURL)
	if status.LanURL != "" {
		fmt.Printf("局域网/公网访问: %s\n", status.LanURL)
	} else {
		fmt.Println("局域网/公网访问: （未检测到可用网卡 IP）")
	}
	fmt.Printf("防火墙检测: %s\n", status.FirewallMessage)
	fmt.Printf("提示: %s\n", status.CloudTip)
}

func networkStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := getNetworkStatus()
	if status.Port == 0 && listenPort > 0 {
		status = refreshNetworkStatus(listenPort)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(status)
}

// FirewallCheckResult 防火墙端口检测结果
type FirewallCheckResult struct {
	Status  string // ok | may_block | unknown
	Message string
}

func currentExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return exe
}
