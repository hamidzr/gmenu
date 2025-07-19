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
git clone https://github.com/your-username/gmenu.git
cd gmenu
make build
```

The binary will be available at `bin/gmenu`.

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

Config files are located at:
- `~/.config/gmenu/config.yaml`
- `~/.gmenu/config.yaml`

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
make build

# Build for multiple platforms
make build-all

# Install dependencies
make get-deps
```

### Testing

```bash
# Run all tests
make test

# Run specific tests
go test ./core/...

# Lint code
make lint

# Format code
make fmt
```

## Architecture

- **CLI Layer** (`internal/cli/`): Command-line interface
- **Core** (`core/`): Application logic and menu management
- **Rendering** (`render/`): UI components and theming
- **Configuration** (`internal/config/`, `model/`): Config management
- **Storage** (`store/`): State persistence

## Inspiration

- [suckless dmenu](https://tools.suckless.org/dmenu/)
- [dmenu-mac](https://github.com/oNaiPs/dmenu-mac)
- [rofi](https://github.com/davatorium/rofi)

## License

[Add your license here]
