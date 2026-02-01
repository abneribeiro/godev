# GoDev

> A lightweight, terminal-based HTTP API inspector and testing tool

[![Version](https://img.shields.io/badge/version-0.4.0-blue.svg)](https://github.com/abneribeiro/godev/releases)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build](https://img.shields.io/badge/build-passing-brightgreen.svg)]()


<div align="center">

   https://github.com/user-attachments/assets/a81c4c2c-29ae-4fa9-ac02-55698c4f94c3

</div>

## Overview

**GoDev** is a modern, terminal-based alternative to Postman and Insomnia, designed for developers who live in the command line. Built with Go and the elegant Bubbletea framework, it provides a fast, keyboard-driven interface for testing HTTP APIs and managing PostgreSQL databases.

### Key Features

#### HTTP API Testing
- **Full HTTP Support** - GET, POST, PUT, DELETE, PATCH methods
- **Request Builder** - Intuitive TUI for building API requests
- **Header Management** - Add, edit, and delete custom headers
- **Query Parameters** - Visual editor with full persistence
- **JSON Body Editor** - Built-in validation and syntax support
- **Response Viewer** - Formatted JSON with syntax highlighting
- **Request Persistence** - Save and reload frequently used requests
- **Request History** - Track last 100 executions with full details
- **Search & Filter** - Find saved requests instantly
- **cURL Export** - Copy requests as cURL commands

#### PostgreSQL Database
- **Database Connections** - Connect to PostgreSQL databases
- **SQL Query Editor** - Execute queries with syntax highlighting
- **Result Viewer** - Formatted table display with scroll
- **Query Management** - Save and organize frequently used queries
- **Query History** - Track last 100 executions
- **Connection Persistence** - Save database configurations

#### Environment Variables
- **Multiple Environments** - Create dev, staging, prod configurations
- **Variable Management** - Define reusable variables (API_URL, API_KEY, etc)
- **Template Syntax** - Use {{VARIABLE}} in URLs, headers, and body
- **Active Environment** - Switch between environments instantly
- **Visual Indicator** - See active environment in request builder

#### General
- **Visual Feedback** - Confirmation messages for all operations
- **Clipboard Integration** - Copy responses and commands
- **Offline-First** - No telemetry, all data stored locally
- **Keyboard-Driven** - Fast navigation with F-keys and shortcuts
- **Home Screen** - Choose between API or Database mode at startup

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

### Using Database Features

#### Connecting to PostgreSQL

```
1. Press Ctrl+D to enter Database mode
2. Press 'c' to open connection form
3. Fill in:
   - Host: localhost
   - Port: 5432
   - Database: your_database
   - User: postgres
   - Password: your_password
4. Press Enter to connect
```

#### Executing SQL Queries

```
1. From Database menu, press 'q' for Query Editor
2. Write your SQL:
   SELECT * FROM users LIMIT 10;
3. Press F5 to execute
4. View results in formatted table
5. Press Ctrl+S to save query for later
```

#### Managing Saved Queries

```
1. From Database menu, press 'l' for Saved Queries
2. Navigate with ↑↓
3. Press Enter to load query in editor
4. Press 'd' to delete
```

### Using Environment Variables

#### Creating Environments

```
1. Press F6 from API mode to open Environment Variables
2. Press 'n' to create new environment
3. Enter name: dev, staging, or prod
4. Press Ctrl+S to save
```

#### Adding Variables

```
1. Select environment and press Enter
2. Press 'n' to add variable
3. Key: API_URL
4. Value: https://api.dev.example.com
5. Press Tab, then Enter to save
```

#### Using Variables in Requests

```
1. Set environment as active (press 's' on selected environment)
2. In request builder, use template syntax:
   - URL: {{API_URL}}/users
   - Headers: Authorization: Bearer {{API_TOKEN}}
   - Body: {"endpoint": "{{API_URL}}"}
3. Variables are replaced automatically when sending request
4. Active environment shown in title: [ENV: dev]
```

## Keyboard Shortcuts

> [!NOTE]
> **Updated in v0.4.1**: F-key shortcuts have been replaced with Ctrl-based alternatives for better terminal compatibility.

### Global Navigation
| Key | Action |
|-----|--------|
| `Ctrl+H` / `?` | Show help |
| `Ctrl+Q` | Quit application |
| `Ctrl+C` | Cancel/Quit |
| `Esc` | Back/Cancel |
| `Tab` | Next field |
| `↑↓` | Navigate lists |

### API Mode - Main Actions
| Key | Action |
|-----|--------|
| `Ctrl+Enter` | Send request |
| `Ctrl+L` | Load saved requests |
| `Ctrl+R` | Request history |
| `Ctrl+D` | Database mode |
| `Ctrl+E` | Environment variables |

### API Mode - Editing
| Key | Action |
|-----|--------|
| `h` | Edit headers |
| `b` | Edit body |
| `q` | Edit query parameters |
| `s` | Save current request |
| `x` | Copy request as cURL |
| `c` | Copy response |
| `/` | Search (in lists) |
| `←/→` | Change HTTP method |

### Database Mode
| Key | Action |
|-----|--------|
| `c` | Connect to database |
| `q` | Query editor |
| `l` | Saved queries |
| `d` | Disconnect |
| `Ctrl+Enter` | Execute query |
| `Ctrl+S` | Save query |

### Environment Variables
| Key | Action |
|-----|--------|
| `n` | New environment/variable |
| `Enter` | Edit environment |
| `s` | Set as active |
| `d` | Delete |
| `e` | Edit variable |
| `Esc` | Back |


## Configuration

### Storage Location

All data is stored locally in `~/.godev/`:
```
~/.godev/
├── config.json         # HTTP requests and history
├── database.json       # Database queries and connections
├── environments.json   # Environment variables
└── exports/            # Exported query results
```

### Data Structure

**config.json** (HTTP):
```json
{
  "version": "0.4.0",
  "requests": [
    {
      "id": "uuid",
      "name": "GET https://api.example.com",
      "method": "GET",
      "url": "https://api.example.com",
      "headers": {"Authorization": "Bearer token"},
      "body": "{\"key\": \"value\"}",
      "query_params": {"page": "1", "limit": "10"},
      "created_at": "2025-10-18T...",
      "last_used": "2025-10-18T..."
    }
  ],
  "history": [
    {
      "id": "uuid",
      "timestamp": "2025-10-18T...",
      "method": "GET",
      "url": "https://api.example.com",
      "status_code": 200,
      "response_time_ms": 145
    }
  ]
}
```

**database.json** (PostgreSQL):
```json
{
  "version": "0.4.0",
  "saved_queries": [
    {
      "id": "uuid",
      "name": "Get Users",
      "query": "SELECT * FROM users LIMIT 10",
      "created_at": "2025-10-18T...",
      "last_used": "2025-10-18T..."
    }
  ],
  "query_history": [
    {
      "id": "uuid",
      "timestamp": "2025-10-18T...",
      "query": "SELECT COUNT(*) FROM users",
      "rows_affected": 1,
      "execution_time_ms": 23
    }
  ],
  "saved_connections": [
    {
      "host": "localhost",
      "port": 5432,
      "database": "mydb",
      "user": "postgres",
      "sslmode": "disable"
    }
  ]
}
```

**environments.json** (Environment Variables):
```json
{
  "version": "0.4.0",
  "environments": [
    {
      "name": "dev",
      "variables": [
        {"key": "API_URL", "value": "https://api.dev.example.com"},
        {"key": "API_TOKEN", "value": "dev_token_123"}
      ]
    },
    {
      "name": "prod",
      "variables": [
        {"key": "API_URL", "value": "https://api.example.com"},
        {"key": "API_TOKEN", "value": "prod_token_456"}
      ]
    }
  ],
  "active_environment": "dev"
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

### v0.4.0 (Current)
- [x] Home screen with mode selection (API/Database)
- [x] Standardized keyboard shortcuts (F-keys)
- [x] Environment variables support
- [x] Multiple environments (dev/staging/prod)
- [x] Variable template syntax {{VAR}}
- [x] Active environment indicator
- [x] F5 key for query execution
- [ ] Import cURL commands as requests
- [ ] Custom color themes
- [ ] Request collections/folders

### v0.5.0 (Planned)
- [ ] Request chaining (use response in next request)
- [ ] Multiple database connections UI
- [ ] Transaction support (BEGIN/COMMIT/ROLLBACK)
- [ ] Query templates
- [ ] SQL autocomplete
- [ ] GraphQL support

### v0.6.0 (Planned)
- [ ] WebSocket inspector
- [ ] Response time metrics and charts
- [ ] Export/Import collections
- [ ] Team sharing features
- [ ] Custom themes and color schemes

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
