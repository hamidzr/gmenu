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
make build
# or: go build -o bin/gmenu -v ./cmd/main.go

# Build for multiple platforms (Darwin amd64/arm64)
make build-all

# Install dependencies
make get-deps

# Run tests
make test

# Lint code
make lint

# Format code  
make fmt

# Development mode (example usage)
make dev
```

### Testing
- Uses `gotestsum` for test execution
- Tests are in core/ directory (e.g., search_test.go, ui_test.go)
- Run specific tests: `go test ./core/...`

## Architecture

### Core Components

1. **CLI Layer** (`internal/cli/`)
   - Cobra-based command line interface in `cli.go`
   - Handles flag parsing and stdin reading

2. **Core Application** (`core/`)
   - `gmenu.go`: Main GMenu struct and application lifecycle
   - `menu.go`: Menu state management and selection logic
   - `search.go`: Search functionality with fuzzy/exact/regex methods
   - `keyhandler.go`: Keyboard input handling
   - `terminal.go`: Terminal mode implementation

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
   - File-based storage for menu state persistence
   - Cache management for selection history

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
- Uses Go modules with go.mod
- Key dependencies: Fyne (GUI), Cobra (CLI), Viper (config), sahilm/fuzzy (search)
- Tests use testify/stretchr

### Code Organization
- `model/`: Data structures and types
- `pkg/`: Reusable packages
- `internal/`: Internal packages not meant for external use
- `constant/`: Application constants

### Current Development Focus
Based on README TODO section, active areas include:
- Fuzzy search improvements
- Multiple selection support  
- Terminal mode enhancements
- dmenu compatibility options

### Testing Strategy
- Unit tests in core/ package
- Use `make test` or `gotestsum -- ./...`
- Focus on search functionality and UI components