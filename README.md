# gmenu

A fast, fuzzy menu selector for desktop and terminal environments, inspired by dmenu and rofi.

## Features

- **GUI Mode**: Desktop interface using Fyne framework
- **Terminal Mode**: Command-line interface for headless environments
- **Fuzzy Search**: Intelligent matching with sahilm/fuzzy library
- **Configurable**: YAML config files, environment variables, and CLI flags
- **Menu State**: Caching and persistence of selections
- **Single Instance**: Per-menu ID instance enforcement

## Installation

### From Source

```bash
git clone https://github.com/hamidzr/gmenu.git
cd gmenu
just build
```

The binary will be available at `bin/gmenu` and installed into your Go bin dir
(`$GOBIN` or `$GOPATH/bin`) for PATH usage.

## Usage

### Basic Usage

```bash
# GUI mode (default)
echo -e "option1\noption2\noption3" | gmenu

# Terminal mode
echo -e "option1\noption2\noption3" | gmenu --terminal
```

### Configuration

gmenu uses a hierarchical configuration system:

1. CLI flags (highest priority)
2. Environment variables (`GMENU_` prefix)
3. YAML config files (lowest priority)

Config files are located at (first match wins):
- `~/.config/gmenu/<menu-id>/config.yaml` or `~/.config/gmenu/config.yaml`
- `~/.gmenu/<menu-id>/config.yaml` or `~/.gmenu/config.yaml`
- `$XDG_CONFIG_HOME/gmenu/<menu-id>/config.yaml` or `$XDG_CONFIG_HOME/gmenu/config.yaml` (macOS: `~/Library/Application Support/gmenu/...`)

See `CONFIG.md` for the full search order and menu ID details.

### Menu IDs

Use menu IDs to maintain separate state for different use cases:

```bash
echo -e "file1\nfile2\nfile3" | gmenu --menu-id files
echo -e "action1\naction2\naction3" | gmenu --menu-id actions
```

## Development

### Building

```bash
# Build main binary
just build

# Build for multiple platforms
just build-all

# Install dependencies and tools
just setup

# Clean build artifacts
just clean
```

### Testing

```bash
# Run all tests
just test

# Run specific tests
go test ./core/...

# Lint code
just lint

# Format code
just fmt

# Run all checks (format, lint, test)
just check
```

## Architecture

- **CLI Layer** (`internal/cli/`): Command-line interface
- **Core** (`core/`): Application logic and menu management
- **Rendering** (`render/`): UI components and theming
- **Configuration** (`internal/config/`, `model/`): Config management
- **Storage** (`store/`): State persistence

## Inspiration and Related Projects

- [suckless dmenu](https://tools.suckless.org/dmenu/)
- [dmenu-mac](https://github.com/oNaiPs/dmenu-mac)
- [rofi](https://github.com/davatorium/rofi)
- [choose](https://github.com/chipsenkbeil/choose)
