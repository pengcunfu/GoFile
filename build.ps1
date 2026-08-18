param(
    [switch]$Web,
    [switch]$Desktop
)

$ErrorActionPreference = 'Stop'

if ($Web -or -not ($Web -or $Desktop)) {
    Write-Host 'Building Web version...' -ForegroundColor Cyan
    go build -o FireShare.exe .
    Write-Host 'Web build complete: FireShare.exe' -ForegroundColor Green
}

if ($Desktop -or -not ($Web -or $Desktop)) {
    Write-Host 'Building Desktop version...' -ForegroundColor Cyan
    go build -tags "desktop production" -ldflags "-H windowsgui" -o FireShare-Desktop.exe .
    Write-Host 'Desktop build complete: FireShare-Desktop.exe' -ForegroundColor Green
}
