#!/bin/bash

# Script para build local de todos os binários
# Útil para testar antes de criar a release

set -e

VERSION=${1:-"dev"}
OUTPUT_DIR="dist"

echo "🔨 Compilando GoDev $VERSION para múltiplas plataformas..."
echo ""

# Criar diretório de output
mkdir -p "$OUTPUT_DIR"

# Flags de build (reduz tamanho do binário)
LDFLAGS="-s -w"

# Linux AMD64
echo "📦 Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$OUTPUT_DIR/godev-linux-amd64"

# Linux ARM64
echo "📦 Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$OUTPUT_DIR/godev-linux-arm64"

# macOS AMD64 (Intel)
echo "📦 macOS AMD64 (Intel)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$OUTPUT_DIR/godev-darwin-amd64"

# macOS ARM64 (Apple Silicon)
echo "📦 macOS ARM64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$OUTPUT_DIR/godev-darwin-arm64"

# Windows AMD64
echo "📦 Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$OUTPUT_DIR/godev-windows-amd64.exe"

echo ""
echo "✅ Build completo!"
echo ""
echo "📋 Binários gerados em $OUTPUT_DIR/:"
ls -lh "$OUTPUT_DIR/"

# Gerar checksums
echo ""
echo "🔐 Gerando checksums SHA256..."
cd "$OUTPUT_DIR"
sha256sum godev-* > checksums.txt
cat checksums.txt
cd ..

echo ""
echo "✅ Tudo pronto! Arquivos em ./$OUTPUT_DIR/"
