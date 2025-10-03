#!/bin/bash
set -e

# NewsGoat installer script
# Usage: curl -sSL https://raw.githubusercontent.com/jarv/newsgoat/main/install.sh | bash

REPO="jarv/newsgoat"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  linux*)
    OS="linux"
    ;;
  darwin*)
    OS="darwin"
    ;;
  *)
    echo "Error: Unsupported operating system: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64)
    ARCH="amd64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

BINARY_NAME="newsgoat-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}"

echo "Installing NewsGoat..."
echo "OS: $OS"
echo "Architecture: $ARCH"
echo "Download URL: $DOWNLOAD_URL"
echo ""

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download binary
echo "Downloading NewsGoat..."
if ! curl -L -o "$TMP_DIR/newsgoat" "$DOWNLOAD_URL"; then
  echo "Error: Failed to download NewsGoat"
  exit 1
fi

# Make executable
chmod +x "$TMP_DIR/newsgoat"

# Install binary
echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_DIR/newsgoat" "$INSTALL_DIR/newsgoat"
else
  echo "Note: Installing to $INSTALL_DIR requires sudo permissions"
  sudo mv "$TMP_DIR/newsgoat" "$INSTALL_DIR/newsgoat"
fi

echo ""
echo "NewsGoat installed successfully!"
echo "Run 'newsgoat' to get started."
