//go:build windows

package main

import "golang.org/x/sys/windows"

// userDocumentsDir 通过 Windows Known Folder API 动态获取「文档」目录
// （支持 OneDrive 重定向、用户自定义位置等，而非写死 %USERPROFILE%\Documents）
func userDocumentsDir() (string, error) {
	return windows.KnownFolderPath(windows.FOLDERID_Documents, windows.KF_FLAG_DEFAULT)
}
