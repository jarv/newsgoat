#!/bin/bash
set -e

VERSION="$1"

# Build with CGO enabled and version information
LDFLAGS="-w -X github.com/jarv/newsgoat/internal/version.GitHash=v${VERSION}"

# Linux (CGO requires gcc for cross-compilation)
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc go build -ldflags "$LDFLAGS" -o newsgoat-linux-amd64 .
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -ldflags "$LDFLAGS" -o newsgoat-linux-arm64 .

# macOS (requires osxcross or native build)
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o newsgoat-darwin-amd64 .
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o newsgoat-darwin-arm64 .

# Windows (requires mingw-w64)
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags "$LDFLAGS" -o newsgoat-windows-amd64.exe .

# Create checksums
sha256sum newsgoat-* > checksums.txt
