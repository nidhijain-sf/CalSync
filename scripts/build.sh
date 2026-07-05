#!/bin/bash
# Build both Mac and Windows binaries and copy to dist/

set -e

cd "$(dirname "$0")/.."

echo "Building Mac binary..."
GOOS=darwin GOARCH=amd64 go build -o dist/mac/SyncApp .
echo "  -> dist/mac/SyncApp"

echo "Building Windows binary..."
GOOS=windows GOARCH=amd64 go build -o dist/windows/SyncApp.exe .
echo "  -> dist/windows/SyncApp.exe"

echo "Syncing templates..."
cp templates/index.html dist/mac/templates/index.html
cp templates/index.html dist/windows/templates/index.html
echo "  -> dist/mac/templates/index.html"
echo "  -> dist/windows/templates/index.html"

echo ""
echo "Build complete."
