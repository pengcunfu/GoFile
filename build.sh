#!/usr/bin/env bash
set -e

echo "构建 Web 版..."
go build -o GoFile .
echo "Web 版构建完成: GoFile"

echo "构建桌面版..."
go build -tags "desktop production" -o GoFile-Desktop .
echo "桌面版构建完成: GoFile-Desktop"
