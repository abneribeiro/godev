# Changelog

All notable changes to GoDev will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2025-10-19

### Added

- **PostgreSQL Database Support**: Full database integration
  - Connect to PostgreSQL databases
  - SQL query editor with syntax highlighting
  - Execute SELECT, INSERT, UPDATE, DELETE queries
  - Schema viewer (tables, columns, types)
  - Save favorite queries
  - Query execution history (last 100)
  - Export results to CSV/JSON
  - Multiple saved connections
  - Database mode accessible via Ctrl+D or F5

- **Environment Variables**: Complete environment management
  - Multiple environments (dev, staging, prod, etc.)
  - CRUD operations for variables
  - Template syntax `{{VARIABLE}}` in URLs, headers, and body
  - Automatic substitution on request send
  - Active environment indicator in title
  - Persistent storage in `~/.godev/environments.json`

- **Home Screen**: Initial mode selection interface
  - Choose between API Testing and Database modes
  - Clean navigation with visual indicators
  - F-key shortcuts displayed

- **F-Key Shortcuts**: Standardized quick navigation
  - F1 / ?: Help screen
  - F2: Send request / Execute query
  - F3: Load saved requests
  - F4: Request history
  - F5: Database mode / Execute SQL
  - F6: Environment variables

- **Request History**: Complete execution tracking
  - Stores last 100 request executions
  - Full request details (method, URL, headers, body, query params)
  - Response data (status, body, execution time)
  - Error tracking
  - Load request from history
  - Delete individual executions
  - Clear all history (with confirmation)

- **Search and Filter**: Real-time search for saved requests
  - Filter by name, method, or URL
  - Case-insensitive search
  - Visual feedback for results
  - Preserved functionality (load, delete) on filtered lists

- **cURL Export**: Export requests as cURL commands
  - Multi-line formatted output
  - Automatic clipboard copy
  - All HTTP methods supported
  - Headers and body included

- **Table Component**: Data display for database results
  - Dynamic column sizing
  - Scrollable rows
  - Formatted headers
  - Visual separators

### Changed

- **UI Organization**: Complete state machine refactor
  - Added 12 new states (Database modes, Environments, History, etc.)
  - Standardized F-key navigation across all modes
  - Consistent visual design between API and Database modes

- **Enhanced Metrics**: Request and query execution metrics
  - Response time in milliseconds
  - Body size formatting (KB, MB)
  - Rows affected for database operations

- **Improved Storage**: Multiple configuration files
  - `~/.godev/config.json` - HTTP requests
  - `~/.godev/database.json` - Database queries and connections
  - `~/.godev/environments.json` - Environment variables

### Technical

- **Dependencies**: Added PostgreSQL driver (`github.com/lib/pq`)
- **Architecture**: Modular design with separate database and storage layers
- **Code Size**: ~3000 lines added across all components
- **State Machine**: Expanded from 8 to 20+ states

## [0.2.0] - 2025-10-17

### Added

- **Query Parameters Editor**: Full CRUD operations for URL query parameters
  - Add, edit, and delete parameters with dedicated UI
  - Parameters automatically appended to request URL
  - Persistent storage with saved requests
  - Visual preview of final URL with query params

- **JSON Body Editor**: Multi-line text editor with validation
  - Real-time JSON syntax validation
  - Visual error feedback with red borders
  - 10,000 character limit
  - Ctrl+S to save and validate

- **Response Headers Viewer**: Toggle between response body and headers
  - Press 'h' to switch views
  - Formatted key-value display
  - Scrollable for long header lists

- **Visual Feedback System**: User confirmation for operations
  - "Copied to clipboard!" message (3-second auto-dismiss)
  - "Request saved successfully!" message (3-second auto-dismiss)
  - [SAVED] indicator in title when request is saved
  - Delete confirmation dialog (press 'd' then 'y')

- **Clipboard Integration**: Copy response body with 'c' key

- **URL Validation**: Comprehensive validation before sending requests
  - Protocol validation (http/https required)
  - Host validation
  - Clear error messages for invalid URLs

- **Quick Send**: Press Enter in URL field to immediately send request

- **Configuration Migration**: Automatic migration from DevScope
  - Migrates `~/.devscope` to `~/.godev` on first launch
  - Preserves all saved requests

### Changed

- **Project Renamed**: DevScope → GoDev
  - New module path: `github.com/abneribeiro/godev`
  - Configuration directory: `~/.godev/`
  - Binary name: `godev`

- **Enhanced Request Preview**: Body preview increased from 30 to 80 characters

- **Improved Storage**: Query parameters now persisted with saved requests

### Fixed

- Error message capitalization for Go conventions
- Unused struct fields removed
- Storage initialization error handling improved

### Technical

- **Dependencies**: Updated to latest Bubbletea, Lipgloss, and Bubbles versions
- **State Machine**: Added StateQueryEditor state
- **Model Fields**: 13 new fields for enhanced functionality
- **Code Size**: ~440 lines added across model.go and editors.go

## [0.1.2] - 2025-10-16

### Fixed

- **Clipboard Operations**: Fixed Ctrl+V paste functionality
  - Ctrl+V now works for pasting URLs
  - Ctrl+A selects all text
  - Ctrl+C copies selected text
  - Ctrl+X cuts selected text

### Changed

- **URL Input**: Increased character limit from 500 to 2000 characters
  - Supports very long URLs with many parameters

## [0.1.1] - 2025-10-16

### Fixed

- **URL Input Field**: Fixed input not accepting keystrokes
  - Keys now properly passed to text input component
  - Navigation keys work as expected

- **Responsive Width**: Input field now adjusts to window size
  - Dynamic width calculation based on terminal dimensions

- **UI Layout**: Improved visual styling and alignment
  - Rounded borders on focused input
  - Better visual feedback for active field

## [0.1.0] - 2025-10-16

### Added

- Initial release of DevScope (renamed to GoDev in v0.2.0)
- HTTP request builder with GET, POST, PUT, DELETE, PATCH support
- JSON response viewer with syntax highlighting
- Request persistence to local storage
- Header editor with add/edit/delete capabilities
- Saved request list with quick loading
- TUI navigation with keyboard shortcuts

---

## Version History

- **0.2.0**: Full-featured release with query params, validation, and visual feedback
- **0.1.2**: Clipboard fixes
- **0.1.1**: Input field fixes
- **0.1.0**: Initial release

[0.2.0]: https://github.com/abneribeiro/godev/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/abneribeiro/godev/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/abneribeiro/godev/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/abneribeiro/godev/releases/tag/v0.1.0
