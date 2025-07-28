# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gmenu is a fuzzy menu selector written in Go that provides a GUI (via Fyne framework) and terminal interface for selecting items from lists. It's inspired by dmenu, rofi, and similar tools. The application supports reading items from stdin and outputs the selected item.

## Key Features

- GUI mode using Fyne framework for desktop interfaces
- Terminal mode for command-line usage
- Fuzzy search capabilities using sahilm/fuzzy library
- Configurable via YAML files, environment variables, and CLI flags
- Menu state management with caching and persistence
- Single instance enforcement per menu ID

## Commands

### Build and Development
```bash
# Build the main binary
just build
# or: go build -o bin/gmenu -v ./cmd/main.go

# Build for multiple platforms (Darwin amd64/arm64)
just build-all

# Install dependencies and tools
just setup

# Run tests
just test

# Lint code
just lint

# Format code and clean trailing spaces
just fmt

# Run all checks (format, lint, test)
just check

# Development mode (example usage)
just dev

# Clean build artifacts
just clean
```

### Testing
- Uses `gotestsum` for test execution
- Tests are in core/ directory (e.g., search_test.go, ui_test.go, terminal_test.go)
- Run specific tests: `go test ./core/...`
- Run all checks with: `just check`

## Architecture

### Core Components

1. **CLI Layer** (`internal/cli/`)
   - Cobra-based command line interface in `cli.go`
   - Handles flag parsing and stdin reading

2. **Core Application** (`core/`)
   - `gmenu.go`: Main GMenu struct and application lifecycle
   - `gmenuitems.go`: Menu item management and operations
   - `gmenurun.go`: Application execution and runtime logic
   - `grender.go`: Core rendering coordination
   - `menu.go`: Menu state management and selection logic
   - `search.go`: Search functionality with fuzzy/exact/regex methods
   - `keyhandler.go`: Keyboard input handling
   - `terminal.go`: Terminal mode implementation
   - `util.go`: Utility functions

3. **Rendering** (`render/`)
   - `items.go`: Item list rendering
   - `input.go`: Search input rendering  
   - `layout.go`: Window layout management
   - `theme.go`: Theme and styling

4. **Configuration** (`internal/config/`, `model/`)
   - Hierarchical config system: CLI flags > env vars > config files
   - Menu ID-based namespacing for different use cases
   - Config files in ~/.config/gmenu/ or ~/.gmenu/

5. **Storage** (`store/`)
   - `store.go`: Storage interface and main operations
   - `file-store.go`: File-based storage implementation
   - `cache.go`: Cache management for selection history
   - `config.go`: Storage configuration
   - `utils.go`: Storage utility functions

### Entry Points

- `cmd/main.go`: Main GUI application entry point
- `cmd/terminal/main.go`: Terminal mode example/testing

### Configuration Management

gmenu uses a sophisticated config system with three layers:
1. CLI flags (highest priority)
2. Environment variables (GMENU_ prefix)
3. YAML config files (lowest priority)

Config files support menu ID namespacing for different use cases.

## Development Notes

### Dependencies
- Uses Go modules with go.mod (Go 1.23+)
- Key dependencies: 
  - Fyne v2.5.2 (GUI framework)
  - Cobra v1.8.1 (CLI framework)
  - Viper v1.20.1 (configuration management)
  - sahilm/fuzzy v0.1.1 (fuzzy search)
  - logrus v1.9.3 (logging)
- Tests use testify/stretchr v1.10.0
- Build tools: gotestsum, golangci-lint

### Code Organization
- `model/`: Data structures and types (config.go, error.go, menu_item.go)
- `pkg/`: Reusable packages (config loading, utilities)
- `internal/`: Internal packages (cli, config, logger)
- `constant/`: Application constants
- `samples/`: Sample input files for testing
- `scripts/`: Build and deployment scripts

### Current Development Focus
Based on README TODO section, active areas include:
- Fuzzy search improvements
- Multiple selection support  
- Terminal mode enhancements
- dmenu compatibility options

### Testing Strategy
- Comprehensive test suite across multiple packages
- Core functionality tests: search_test.go, ui_test.go, terminal_test.go
- Integration tests: integration_test.go, keyboard_test.go
- Configuration tests: config_test.go, model_test.go
- CLI tests: cli_test.go
- Storage tests: store_test.go
- Rendering tests: render_test.go
- Use `just test` or `gotestsum -- ./...`
- Use `just check` for full validation (format, lint, test)

## Dev Memories

- On focus loss, the menu should behave exactly as a user-cancelled event that comes from pressing escape 