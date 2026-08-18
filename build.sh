#!/usr/bin/env bash
set -e

echo "Building Web version..."
go build -o FireShare .
echo "Web build complete: FireShare"

echo "Building Desktop version..."
go build -tags "desktop production" -o FireShare-Desktop .
echo "Desktop build complete: FireShare-Desktop"
