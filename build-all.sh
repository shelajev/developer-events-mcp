#!/bin/bash
# Build binaries for all platforms

set -e

VERSION="1.0.0"
BINARY_NAME="developer-events-mcp"

echo "Building Developer Events MCP Server v${VERSION}"
echo ""

# Build for Linux x86_64
echo "Building for Linux (x86_64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "bin/${BINARY_NAME}-linux-amd64" main.go

# Build for Linux ARM64
echo "Building for Linux (ARM64)..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o "bin/${BINARY_NAME}-linux-arm64" main.go

# Build for macOS x86_64
echo "Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "bin/${BINARY_NAME}-darwin-amd64" main.go

# Build for macOS ARM64 (Apple Silicon)
echo "Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "bin/${BINARY_NAME}-darwin-arm64" main.go

# Build for Windows
echo "Building for Windows (x86_64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "bin/${BINARY_NAME}-windows-amd64.exe" main.go

echo ""
echo "✅ Build complete! Binaries:"
echo ""
ls -lh bin/
echo ""
echo "Total size:"
du -sh bin/