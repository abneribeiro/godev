# GoDev v0.2.0 - Full-Featured Release

We're excited to announce GoDev v0.2.0, a complete rebranding and major feature release! This version introduces query parameter management, JSON validation, visual feedback systems, and automatic configuration migration.

## What's New

This release marks the transition from **DevScope** to **GoDev**, with a complete set of new features that make API testing more efficient and user-friendly.

## Features

### Query Parameters Editor
- **Full CRUD operations**: Add, edit, and delete URL query parameters with a dedicated visual editor
- **Automatic URL construction**: Parameters are automatically appended to your request URL
- **Persistent storage**: Query params are now saved with your requests
- **Real-time preview**: See the final URL with all parameters before sending

### JSON Body Editor with Validation
- **Multi-line editor**: Comfortable editing area for JSON payloads (10,000 char limit)
- **Real-time validation**: Instant feedback when your JSON has syntax errors
- **Visual error indicators**: Red borders and clear error messages guide you to fix issues
- **Save shortcut**: Ctrl+S to save and validate in one action

### Visual Feedback System
- **Operation confirmations**: Clear visual feedback for copy and save operations
- **Auto-dismiss messages**: 3-second confirmations that don't interrupt your workflow
- **Saved indicator**: [SAVED] badge shows when a request is persisted
- **Delete confirmation**: Two-step process prevents accidental deletions

### Response Headers Viewer
- **Quick toggle**: Press 'h' to switch between response body and headers
- **Formatted display**: Clean key-value presentation
- **Full scrolling**: Navigate through long header lists easily

### Configuration Migration
- **Automatic migration**: Seamlessly migrates your data from `~/.devscope` to `~/.godev`
- **Zero manual work**: Just launch the new version, and your requests are ready

## Improvements

- **Enhanced body preview**: Now shows 80 characters instead of 30 for better visibility
- **URL validation**: Comprehensive checks with helpful error messages before sending requests
- **Quick send**: Press Enter directly in the URL field to send requests faster
- **Clipboard integration**: Copy response bodies with a single keystroke ('c')

## Bug Fixes

- Fixed error message capitalization to follow Go conventions
- Removed unused struct fields that triggered linter warnings
- Improved storage initialization error handling for better reliability

## Breaking Changes

### Project Renamed: DevScope → GoDev

- **Module path**: `github.com/abneribeiro/devscope` → `github.com/abneribeiro/godev`
- **Binary name**: `devscope` → `godev`
- **Config directory**: `~/.devscope/` → `~/.godev/` (automatic migration included)

If you have the old version installed via `go install`, run:
```bash
go install github.com/abneribeiro/godev@latest
```

## Installation

### Quick Install (Linux/macOS)

#### Linux AMD64
```bash
wget https://github.com/abneribeiro/godev/releases/download/v0.2.0/godev-linux-amd64
chmod +x godev-linux-amd64
sudo mv godev-linux-amd64 /usr/local/bin/godev
```

#### macOS Apple Silicon (M1/M2/M3)
```bash
wget https://github.com/abneribeiro/godev/releases/download/v0.2.0/godev-darwin-arm64
chmod +x godev-darwin-arm64
sudo mv godev-darwin-arm64 /usr/local/bin/godev
```

#### macOS Intel
```bash
wget https://github.com/abneribeiro/godev/releases/download/v0.2.0/godev-darwin-amd64
chmod +x godev-darwin-amd64
sudo mv godev-darwin-amd64 /usr/local/bin/godev
```

### Via Go Install

```bash
go install github.com/abneribeiro/godev@latest
```

### Windows

Download `godev-windows-amd64.exe` from the assets below and add it to your PATH.

## Assets

Download the appropriate binary for your platform:

- **godev-linux-amd64** - Linux 64-bit
- **godev-linux-arm64** - Linux ARM64 (Raspberry Pi, etc.)
- **godev-darwin-amd64** - macOS Intel
- **godev-darwin-arm64** - macOS Apple Silicon (M1/M2/M3)
- **godev-windows-amd64.exe** - Windows 64-bit
- **checksums.txt** - SHA256 checksums for verification

### Verify Download

```bash
sha256sum -c checksums.txt
```

## Full Changelog

See [CHANGELOG.md](https://github.com/abneribeiro/godev/blob/main/CHANGELOG.md) for complete details.

## What's Next?

Check out our [roadmap](https://github.com/abneribeiro/godev#roadmap) for upcoming features:

- **v0.3.0**: cURL import/export, request search, custom themes
- **v0.4.0**: Environment variables, GraphQL support, WebSocket inspector
- **v1.0.0**: Performance benchmarking, plugin system, collections

## Acknowledgments

Thank you to everyone who provided feedback on DevScope! Your input helped shape GoDev into a better tool.

Special thanks to:
- [Charm](https://charm.sh/) for the amazing Bubbletea framework
- The Go community for excellent tooling and support

---

**Upgrade Note**: If you're upgrading from DevScope, GoDev will automatically migrate your saved requests from `~/.devscope` to `~/.godev` on first launch. No manual action required!

**Need Help?** [Open an issue](https://github.com/abneribeiro/godev/issues) or check out the [README](https://github.com/abneribeiro/godev/blob/main/README.md) for detailed usage instructions.
