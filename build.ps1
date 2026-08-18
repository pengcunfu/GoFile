param(
    [switch]$Web,
    [switch]$Desktop
)

$ErrorActionPreference = 'Stop'

if ($Web -or -not ($Web -or $Desktop)) {
    Write-Host 'Building Web version...' -ForegroundColor Cyan
    go build -o GoFile.exe .
    Write-Host 'Web build complete: GoFile.exe' -ForegroundColor Green
}

if ($Desktop -or -not ($Web -or $Desktop)) {
    Write-Host 'Building Desktop version...' -ForegroundColor Cyan
    go build -tags "desktop production" -ldflags "-H windowsgui" -o GoFile-Desktop.exe .
    Write-Host 'Desktop build complete: GoFile-Desktop.exe' -ForegroundColor Green
}
