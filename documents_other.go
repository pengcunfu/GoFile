//go:build !windows

package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// userDocumentsDir 动态获取当前用户的「文档」目录（Linux / macOS）
func userDocumentsDir() (string, error) {
	if xdg := os.Getenv("XDG_DOCUMENTS_DIR"); xdg != "" {
		return expandUserPath(xdg)
	}

	// Linux：从 ~/.config/user-dirs.dirs 读取 XDG_DOCUMENTS_DIR
	if dir, err := documentsFromUserDirs(); err == nil && dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Documents"), nil
}

func documentsFromUserDirs() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	f, err := os.Open(filepath.Join(home, ".config", "user-dirs.dirs"))
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		const prefix = "XDG_DOCUMENTS_DIR="
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		value := strings.Trim(strings.TrimPrefix(line, prefix), "\"")
		return expandUserPath(value)
	}
	return "", scanner.Err()
}

func expandUserPath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
