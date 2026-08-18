//go:build !windows

package main

import "fmt"

// checkFirewallPort 非 Windows 平台不做系统防火墙规则探测
func checkFirewallPort(port int) FirewallCheckResult {
	return FirewallCheckResult{
		Status:  "unknown",
		Message: fmt.Sprintf("当前系统未做自动防火墙检测，请确认本机防火墙已放行 TCP %d", port),
	}
}
