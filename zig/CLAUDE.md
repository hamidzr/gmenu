# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

zmenu is a native macOS AppKit-based fuzzy menu selector written in Zig. It's a rewrite of the original Go-based gmenu, providing a GUI interface using AppKit and terminal mode for selecting items from lists. The application reads items from stdin or IPC and outputs the selected item.

## Requirements

- macOS
- Zig 0.15.2+ (zig-objc requires a released Zig version)
- Xcode Command Line Tools (for AppKit headers)

## Commands

### Build and Development
```bash
# Build both zmenu and zmenuctl binaries
just build
# or: zig build

# Run the app (requires stdin input)
just run
# or: zig build run

# Development mode with sample data
just dev

# Run tests
just test
# or: zig build test

# Clean build artifacts
just clean

# Visual regression testing
just visual
```

### Testing
- Tests are primarily in `src/search.zig` (search algorithm tests)
- Visual regression tests capture UI snapshots via `scripts/visual_test.sh`
- Set `UPDATE_SNAPSHOT=1` to refresh the visual baseline
- Requires macOS Accessibility + Screen Recording permissions for visual tests

## Architecture

### Entry Points & Control Flow

1. **main.zig**: Application entry point that parses CLI config and dispatches to either terminal or GUI mode
2. **app.zig**: GUI mode orchestration using AppKit (NSApplication lifecycle, window management, IPC handling)
3. **terminal.zig**: Terminal mode implementation (minimal, selects first match on Enter)
4. **zmenuctl.zig**: IPC client for sending commands to running zmenu instances

### Core Components

1. **Configuration System** (`cli.zig`, `config.zig`)
   - Three-layer config: CLI flags > env vars (GMENU_* prefix) > YAML config files
   - Config locations: `~/.config/gmenu/config.yaml` or `~/.gmenu/config.yaml`
   - Menu ID-based namespacing for different use cases
   - Use `--init-config` to generate default config file

2. **Menu & Items** (`menu.zig`)
   - `MenuItem`: Core data structure with label, index, icon, and optional IPC payload
   - `Model`: Menu state holding items, filtered indices, and selection
   - `readStdinLines`: Reads and parses items from stdin (max 16MB)
   - `parseItem`: Parses item lines, extracts icon hints like `[app] Label`
   - Icon types: none, app, file, folder, info

3. **Search** (`search.zig`)
   - Methods: `direct`, `fuzzy`, `fuzzy1`, `fuzzy3`, `default` (maps to fuzzy)
   - Tokenized, case-insensitive fuzzy matching
   - Scoring system with bonuses for first char, separators, camelCase, adjacent matches
   - Results capped at configurable limit (default 10)
   - `filterIndices`: Core search function that populates filtered index list

4. **IPC System** (`ipc.zig`, IPC_PROTOCOL.md)
   - Unix domain socket at `$TMPDIR/zmenu.<menu_id>.sock`
   - Protocol: `<length>\n<json-payload>` framing
   - Commands: `set`, `append`, `prepend` for dynamic item updates
   - IPC-only mode (`--ipc-only`) ignores stdin, outputs full JSON on selection
   - Size limit: 1MB per message

5. **State Management** (`cache.zig`, `pid.zig`)
   - `cache.zig`: Persists last query + selection to `~/.cache/gmenu/<menu_id>/cache.yaml`
   - `pid.zig`: Single instance enforcement via pid file in temp directory

6. **AppKit Integration** (`app.zig`)
   - Direct Objective-C interop via zig-objc vendor library
   - Custom NSApplication delegate, NSWindow, NSTextField, NSTableView setup
   - `UpdateQueue`: Thread-safe queue for stdin/IPC updates to GUI thread
   - Key handlers for Up/Down/Tab/Enter/Esc/Ctrl+L/1-9 (numeric selection)
   - Double-click support for row selection

### Dependencies

- `zig-objc` (vendored in `vendor/zig-objc`): Objective-C runtime bindings for AppKit interop
- AppKit and Foundation frameworks (macOS system frameworks)

## Key Features

- **Search**: Fuzzy matching with configurable methods and scoring
- **GUI Mode**: Native AppKit window with search field and table view
- **Terminal Mode**: Minimal CLI interface (selects first match)
- **IPC**: Dynamic item updates via Unix socket protocol
- **Configuration**: Hierarchical config (flags > env > files) with menu ID namespacing
- **Theming**: Customizable colors via hex strings (#RRGGBB or #RRGGBBAA)
- **Single Instance**: PID file prevents multiple instances per menu ID
- **Caching**: Persists query and selection state per menu ID
- **Icon Support**: Optional icon column with app/file/folder/info types
- **Numeric Selection**: Keys 1-9 select corresponding items (opt-in)

## Development Notes

### Configuration Management

Config priority (highest to lowest):
1. CLI flags (`--menu-id`, `--search-method`, etc.)
2. Environment variables (`GMENU_MENU_ID`, `GMENU_SEARCH_METHOD`, etc.)
3. YAML config files (`~/.config/gmenu/config.yaml`)

Config files support both snake_case and camelCase keys for compatibility.

### Search Methods

- `direct`: Substring match only
- `fuzzy`: Full fuzzy matching with scoring (default)
- `fuzzy1`: Fuzzy with single-char optimization
- `fuzzy3`: Fuzzy with 3-char minimum
- `default`: Alias for fuzzy

Regex and exact modes from Go gmenu are not implemented.

### IPC Protocol

See `IPC_PROTOCOL.md` for full protocol details. Key points:
- Framed JSON messages: `<length>\n<json-payload>`
- Commands affect running instance's item list
- Items can include extra fields preserved in IPC-only mode output
- `zmenuctl` provides CLI wrapper for IPC commands

### Zig-Specific Patterns

- Uses arena allocator in main for simplified memory management
- AppKit objects are managed via Objective-C runtime (manual retain/release)
- Thread synchronization via std.Thread.Mutex for UpdateQueue
- Error handling via Zig error unions and try/catch

### Migration from Go gmenu

- Config filename is `config.yaml` (not `gmenu.yaml`)
- Numeric selection is opt-out via `no_numeric_selection: false` (opposite of Go version)
- Default window size is 600x300 (max 1920x1080)
- Regex search method not supported (use direct or fuzzy instead)
