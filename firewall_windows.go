//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

// hiddenCommand 创建不弹出黑框的子进程（桌面版 GUI 下执行 netsh 时必需）
func hiddenCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
	return cmd
}

// checkFirewallPort 检测 Windows 防火墙是否可能拦截指定 TCP 入站端口（只读，不改规则、不提权）
func checkFirewallPort(port int) FirewallCheckResult {
	profilesOn, err := windowsFirewallProfilesEnabled()
	if err != nil {
		return FirewallCheckResult{
			Status:  "unknown",
			Message: fmt.Sprintf("无法检测防火墙状态（可手动确认 TCP %d 入站是否放行）", port),
		}
	}
	if !profilesOn {
		return FirewallCheckResult{
			Status:  "ok",
			Message: fmt.Sprintf("Windows 防火墙未启用相关配置文件，TCP %d 通常不会被本机防火墙拦截", port),
		}
	}

	portAllowed, appAllowed, err := windowsInboundAllows(port, currentExecutable())
	if err != nil {
		return FirewallCheckResult{
			Status:  "unknown",
			Message: fmt.Sprintf("无法读取防火墙规则（可手动确认 TCP %d 入站是否放行）", port),
		}
	}
	if portAllowed || appAllowed {
		return FirewallCheckResult{
			Status:  "ok",
			Message: fmt.Sprintf("检测到允许规则，TCP %d 入站在本机防火墙上应可访问", port),
		}
	}
	return FirewallCheckResult{
		Status:  "may_block",
		Message: fmt.Sprintf("Windows 防火墙已开启，且未发现 TCP %d（或本程序）的入站允许规则，局域网/公网访问可能被拦截", port),
	}
}

func windowsFirewallProfilesEnabled() (bool, error) {
	out, err := hiddenCommand("netsh", "advfirewall", "show", "allprofiles").CombinedOutput()
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if !(strings.HasPrefix(line, "状态") || strings.HasPrefix(line, "State")) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		last := fields[len(fields)-1]
		if equalsAny(last, "启用", "ON") {
			return true, nil
		}
	}
	return false, nil
}

func windowsInboundAllows(port int, exe string) (portAllowed, appAllowed bool, err error) {
	out, err := hiddenCommand("netsh", "advfirewall", "firewall", "show", "rule", "name=all", "dir=in", "verbose").CombinedOutput()
	if err != nil {
		return false, false, err
	}

	exeNorm := normalizePath(exe)
	for _, block := range splitNetshRuleBlocks(string(out)) {
		fields := parseNetshBlock(block)
		enabled := firstNonEmpty(fields["已启用"], fields["Enabled"])
		if !equalsAny(enabled, "是", "Yes") {
			continue
		}
		action := firstNonEmpty(fields["操作"], fields["Action"])
		if !equalsAny(action, "允许", "Allow") {
			continue
		}
		proto := firstNonEmpty(fields["协议"], fields["Protocol"])
		localPort := firstNonEmpty(fields["本地端口"], fields["LocalPort"])
		program := firstNonEmpty(fields["程序"], fields["Program"])

		if equalsAny(proto, "TCP", "任何", "Any") && portMatches(localPort, port) {
			portAllowed = true
		}
		if exeNorm != "" && program != "" && normalizePath(program) == exeNorm &&
			equalsAny(proto, "TCP", "任何", "Any") {
			appAllowed = true
		}
		if portAllowed && appAllowed {
			break
		}
	}
	return portAllowed, appAllowed, nil
}

func splitNetshRuleBlocks(text string) []string {
	var starts []int
	for _, m := range []string{"规则名称:", "Rule Name:"} {
		offset := 0
		for {
			i := strings.Index(text[offset:], m)
			if i < 0 {
				break
			}
			starts = append(starts, offset+i)
			offset += i + len(m)
		}
	}
	if len(starts) == 0 {
		return nil
	}
	for i := 1; i < len(starts); i++ {
		for j := i; j > 0 && starts[j-1] > starts[j]; j-- {
			starts[j-1], starts[j] = starts[j], starts[j-1]
		}
	}

	blocks := make([]string, 0, len(starts))
	for i, start := range starts {
		end := len(text)
		if i+1 < len(starts) {
			end = starts[i+1]
		}
		blocks = append(blocks, text[start:end])
	}
	return blocks
}

func parseNetshBlock(block string) map[string]string {
	fields := make(map[string]string)
	for _, line := range strings.Split(block, "\n") {
		key, val := splitNetshField(line)
		if key != "" {
			fields[key] = val
		}
	}
	return fields
}

func splitNetshField(line string) (key, val string) {
	line = strings.TrimSpace(strings.ReplaceAll(line, "\u3000", " "))
	if line == "" || strings.HasPrefix(line, "---") {
		return "", ""
	}
	i := strings.Index(line, ":")
	if i <= 0 {
		return "", ""
	}
	key = strings.TrimSpace(line[:i])
	val = strings.TrimSpace(line[i+1:])
	return key, val
}

func portMatches(localPort string, port int) bool {
	localPort = strings.TrimSpace(strings.ReplaceAll(localPort, "\u3000", " "))
	if localPort == "" {
		return false
	}
	if equalsAny(localPort, "任何", "Any") {
		return true
	}
	want := strconv.Itoa(port)
	for _, p := range strings.Split(localPort, ",") {
		if strings.TrimSpace(p) == want {
			return true
		}
	}
	return false
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if equalsAny(p, "任何", "Any", "System") {
		return strings.ToLower(p)
	}
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	return strings.ToLower(filepath.Clean(p))
}

func equalsAny(v string, candidates ...string) bool {
	v = strings.TrimSpace(v)
	for _, c := range candidates {
		if strings.EqualFold(v, c) {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
