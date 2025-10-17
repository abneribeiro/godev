# GoDev

> A lightweight, terminal-based HTTP API inspector and testing tool

[![Version](https://img.shields.io/badge/version-0.2.0-blue.svg)](https://github.com/abneribeiro/godev/releases)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build](https://img.shields.io/badge/build-passing-brightgreen.svg)]()

## Overview

**GoDev** is a modern, terminal-based alternative to Postman and Insomnia, designed for developers who live in the command line. Built with Go and the elegant Bubbletea framework, it provides a fast, keyboard-driven interface for testing and debugging HTTP APIs.

### Key Features

- **Full HTTP Support** - GET, POST, PUT, DELETE, PATCH methods
- **Request Builder** - Intuitive TUI for building API requests
- **Header Management** - Add, edit, and delete custom headers
- **Query Parameters** - Visual editor with full persistence
- **JSON Body Editor** - Built-in validation and syntax support
- **Response Viewer** - Formatted JSON with syntax highlighting
- **Request Persistence** - Save and reload frequently used requests
- **Visual Feedback** - Confirmation messages for all operations
- **Clipboard Integration** - Copy responses with a single keystroke
- **Offline-First** - No telemetry, all data stored locally

## Installation

### Download Binary (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/abneribeiro/godev/releases):

```bash
# Linux/macOS
wget https://github.com/abneribeiro/godev/releases/download/v0.2.0/godev-linux-amd64
chmod +x godev-linux-amd64
sudo mv godev-linux-amd64 /usr/local/bin/godev

# Or use the install script
curl -sSL https://raw.githubusercontent.com/abneribeiro/godev/main/install.sh | bash
```

### Via Go Install

```bash
go install github.com/abneribeiro/godev@latest
```

### Build from Source

```bash
git clone https://github.com/abneribeiro/godev.git
cd godev
go build -o godev
./godev
```

## Quick Start

1. Launch the application:
   ```bash
   godev
   ```

2. Enter your API URL (e.g., `https://api.github.com/users/octocat`)

3. Navigate with `Tab` and `↑↓` keys

4. Press `Enter` to send the request

5. View the formatted response

6. Press `s` to save the request for later use

## Usage

### Making Your First Request

```
1. URL: https://jsonplaceholder.typicode.com/posts/1
2. Method: GET (default)
3. Press Enter → View response
```

### POST Request with JSON Body

```
1. Method: POST (use ←/→ to change)
2. URL: https://httpbin.org/post
3. Tab to Headers → Press Enter
4. Add header: Content-Type: application/json
5. Esc → Tab to Body → Press Enter
6. Enter JSON: {"title": "Hello", "body": "World"}
7. Ctrl+S to save → Send request
```

### Adding Query Parameters

```
1. Tab to "Query Params" → Press Enter
2. Press 'n' to add new parameter
3. Key: page, Value: 1
4. Press Enter to save
5. URL automatically updates: https://api.example.com?page=1
```

## Keyboard Shortcuts

### Navigation
| Key | Action |
|-----|--------|
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `↑↓` / `j k` | Navigate lists |
| `←→` | Change HTTP method |

### Actions
| Key | Action |
|-----|--------|
| `Enter` | Send request / Select item |
| `Ctrl+L` | Load saved requests |
| `Ctrl+H` | Edit headers |
| `Ctrl+B` | Edit body |
| `Ctrl+Q` | Edit query parameters |
| `s` | Save current request |
| `c` | Copy response to clipboard |
| `?` | Show help |

### Editors
| Key | Action |
|-----|--------|
| `n` / `a` | Add new item |
| `e` | Edit selected item |
| `d` | Delete (with confirmation) |
| `Esc` | Exit editor |
| `Ctrl+S` | Save & validate (body editor) |

## Configuration

### Storage Location

All saved requests are stored in:
```
~/.godev/config.json
```

### Data Structure

```json
{
  "version": "0.2.0",
  "requests": [
    {
      "id": "uuid",
      "name": "GET https://api.example.com",
      "method": "GET",
      "url": "https://api.example.com",
      "headers": {"Authorization": "Bearer token"},
      "body": "{\"key\": \"value\"}",
      "query_params": {"page": "1", "limit": "10"},
      "created_at": "2025-10-17T...",
      "last_used": "2025-10-17T..."
    }
  ]
}
```

### Migration from DevScope

If you previously used DevScope, GoDev will automatically migrate your data from `~/.devscope` to `~/.godev` on first launch.

## Architecture

GoDev is built with:

- **[Go](https://go.dev/)** - Fast, compiled, garbage-collected language
- **[Bubbletea](https://github.com/charmbracelet/bubbletea)** - Elm-inspired TUI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Style definitions for TUIs
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - Common TUI components

```
godev/
├── main.go                    # Application entry point
├── internal/
│   ├── ui/
│   │   ├── model.go           # State machine & business logic
│   │   ├── editors.go         # Header/Body/Query editors
│   │   └── styles.go          # Visual design system
│   ├── http/
│   │   └── client.go          # HTTP client wrapper
│   └── storage/
│       └── requests.go        # Request persistence & migration
└── go.mod
```

## Roadmap

### v0.2.0 (Current)
- [x] HTTP method support (GET, POST, PUT, DELETE, PATCH)
- [x] Header editor with CRUD operations
- [x] Query parameters with persistence
- [x] JSON body editor with validation
- [x] Response viewer with syntax highlighting
- [x] Request persistence and loading
- [x] Visual feedback system
- [x] Delete confirmation dialogs
- [x] Automatic config migration

### v0.3.0
- [ ] Export requests to cURL commands
- [ ] Import cURL commands as requests
- [ ] Search and filter saved requests
- [ ] Custom color themes
- [ ] Request history tracking

### v0.4.0
- [ ] Environment variables support
- [ ] Request chaining (use response in next request)
- [ ] GraphQL support
- [ ] WebSocket inspector
- [ ] Response time metrics

### v1.0.0
- [ ] Performance benchmarking
- [ ] Plugin system
- [ ] Collections/workspaces
- [ ] Team sharing (export/import collections)

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/amazing-feature`
3. Commit your changes: `git commit -m "feat: add amazing feature"`
4. Push to the branch: `git push origin feat/amazing-feature`
5. Open a Pull Request

### Commit Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `refactor:` - Code refactoring
- `docs:` - Documentation changes
- `test:` - Test additions/modifications
- `chore:` - Maintenance tasks

## Troubleshooting

### Colors not displaying correctly

Ensure your terminal supports true color:
```bash
echo $COLORTERM  # Should return "truecolor" or "24bit"
```

### Configuration not saving

Check directory permissions:
```bash
mkdir -p ~/.godev
chmod 755 ~/.godev
```

### Build errors

Verify Go version:
```bash
go version  # Should be 1.21 or higher
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

**Abner Ribeiro**
- GitHub: [@abneribeiro](https://github.com/abneribeiro)

## Acknowledgments

- [Charm](https://charm.sh/) - For the amazing TUI toolkit
- The Go community - For excellent tooling and support

---

**Built with Go** • [Report Bug](https://github.com/abneribeiro/godev/issues) • [Request Feature](https://github.com/abneribeiro/godev/issues)
