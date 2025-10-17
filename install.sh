#!/bin/bash

# GoDev Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/abneribeiro/godev/main/install.sh | bash

set -e

VERSION="v0.2.0"
REPO="abneribeiro/godev"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="godev"

echo "🚀 GoDev Installer"
echo "   Version: $VERSION"
echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $OS in
    linux)
        OS="linux"
        ;;
    darwin)
        OS="darwin"
        ;;
    *)
        echo "❌ Unsupported OS: $OS"
        exit 1
        ;;
esac

case $ARCH in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

BINARY="godev-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"

echo "📥 Downloading GoDev for ${OS}/${ARCH}..."
echo "   URL: $DOWNLOAD_URL"
echo ""

# Download binary
TMP_FILE="/tmp/${BINARY}"
if command -v curl >/dev/null 2>&1; then
    curl -sSL -o "$TMP_FILE" "$DOWNLOAD_URL"
elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$TMP_FILE" "$DOWNLOAD_URL"
else
    echo "❌ Error: curl or wget is required"
    exit 1
fi

# Make executable
chmod +x "$TMP_FILE"

# Install (requires sudo if not writable)
echo "📦 Installing to $INSTALL_DIR/$BINARY_NAME..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
else
    echo "   (requires sudo)"
    sudo mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
fi

echo ""
echo "✅ GoDev installed successfully!"
echo ""
echo "🎯 Quick start:"
echo "   $ godev"
echo ""
echo "📚 Documentation: https://github.com/${REPO}"
echo "🐛 Issues: https://github.com/${REPO}/issues"
echo ""
