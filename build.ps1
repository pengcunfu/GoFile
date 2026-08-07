param(
    [switch]$Web,
    [switch]$Desktop
)

$ErrorActionPreference = 'Stop'

if ($Web -or -not ($Web -or $Desktop)) {
    Write-Host '构建 Web 版...' -ForegroundColor Cyan
    go build -o GoFile.exe .
    Write-Host 'Web 版构建完成: GoFile.exe' -ForegroundColor Green
}

if ($Desktop -or -not ($Web -or $Desktop)) {
    Write-Host '构建桌面版...' -ForegroundColor Cyan
    go build -tags "desktop production" -ldflags "-H windowsgui" -o GoFile-Desktop.exe .
    Write-Host '桌面版构建完成: GoFile-Desktop.exe' -ForegroundColor Green
}
