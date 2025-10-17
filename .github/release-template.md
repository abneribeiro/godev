## What's New

<!-- Describe the main features and improvements in this release -->

## Features

<!-- List new features -->

## Improvements

<!-- List improvements to existing features -->

## Bug Fixes

<!-- List bug fixes -->

## Breaking Changes

<!-- List any breaking changes, or remove this section if none -->

## Installation

### Quick Install (Linux/macOS)

```bash
# AMD64
wget https://github.com/abneribeiro/godev/releases/download/v0.2.0/godev-linux-amd64
chmod +x godev-linux-amd64
sudo mv godev-linux-amd64 /usr/local/bin/godev

# Apple Silicon (M1/M2)
wget https://github.com/abneribeiro/godev/releases/download/v0.2.0/godev-darwin-arm64
chmod +x godev-darwin-arm64
sudo mv godev-darwin-arm64 /usr/local/bin/godev
```

### Via Go Install

```bash
go install github.com/abneribeiro/godev@latest
```

## Assets

Download the appropriate binary for your platform:

- **Linux AMD64**: `godev-linux-amd64`
- **Linux ARM64**: `godev-linux-arm64`
- **macOS Intel**: `godev-darwin-amd64`
- **macOS Apple Silicon**: `godev-darwin-arm64`
- **Windows AMD64**: `godev-windows-amd64.exe`

Verify downloads with `checksums.txt`.

## Full Changelog

See [CHANGELOG.md](https://github.com/abneribeiro/godev/blob/main/CHANGELOG.md) for complete details.

---

**Note**: If you're upgrading from DevScope, GoDev will automatically migrate your saved requests from `~/.devscope` to `~/.godev` on first launch.
