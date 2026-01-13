# Configuration Management

gmenu uses a hierarchical configuration system with the following priority order (highest to lowest):

1. **CLI flags** (highest priority)
2. **Environment variables**
3. **Configuration file** (lowest priority)

## Configuration File

gmenu looks for configuration files in the following locations (in order):

**When a menu ID is provided (e.g., `--menu-id my-menu`):**
1. `~/.config/gmenu/my-menu/config.yaml` (namespaced by menu ID)
2. `~/.gmenu/my-menu/config.yaml` (namespaced by menu ID)
3. `$XDG_CONFIG_HOME/gmenu/my-menu/config.yaml` (or OS user config dir)
4. `~/.config/gmenu/config.yaml` (default)
5. `~/.gmenu/config.yaml` (default)
6. `$XDG_CONFIG_HOME/gmenu/config.yaml` (or OS user config dir)

**When no menu ID is provided:**
1. `~/.config/gmenu/config.yaml`
2. `~/.gmenu/config.yaml`
3. `$XDG_CONFIG_HOME/gmenu/config.yaml` (or OS user config dir)

This namespacing allows you to have different configurations for different use cases. For example, you might have one config for git branch selection and another for file selection.

## Generating Config Files

You can automatically generate config files using the `--init-config` flag:

```bash
# Generate default config file
gmenu --init-config

# Generate namespaced config file for a specific menu ID
gmenu --init-config --menu-id my-menu
```

The generated config files include helpful comments and all available options with their default values.

See `gmenu.yaml.example` for a complete configuration file example.

## Environment Variables

All configuration options can be set via environment variables using the `GMENU_` prefix and converting kebab-case to SNAKE_CASE:

```bash
export GMENU_TITLE="My Custom Menu"
export GMENU_PROMPT="Choose an option"
export GMENU_MENU_ID="main-menu"
export GMENU_SEARCH_METHOD="direct"
export GMENU_PRESERVE_ORDER=true
export GMENU_INITIAL_QUERY=""
export GMENU_AUTO_ACCEPT=false
export GMENU_TERMINAL_MODE=false
export GMENU_NO_NUMERIC_SELECTION=false
export GMENU_MIN_WIDTH=800
export GMENU_MIN_HEIGHT=400
export GMENU_MAX_WIDTH=1200
export GMENU_MAX_HEIGHT=800
export GMENU_ACCEPT_CUSTOM_SELECTION=true
```

## CLI Flags

All options can be overridden via CLI flags:

```bash
gmenu --title "Override Title" --prompt "Override Prompt" --search-method direct
```

Use `gmenu --help` to see all available flags.

## Configuration Options

| Option | CLI Flag | Environment Variable | Config File Key | Default | Description |
|--------|----------|---------------------|-----------------|---------|-------------|
| Title | `--title`, `-t` | `GMENU_TITLE` | `title` | `gmenu` | Title of the menu window |
| Prompt | `--prompt`, `-p` | `GMENU_PROMPT` | `prompt` | `Search` | Prompt text in the search bar |
| Menu ID | `--menu-id`, `-m` | `GMENU_MENU_ID` | `menu_id` | `""` | Unique identifier for menu state |
| Search Method | `--search-method`, `-s` | `GMENU_SEARCH_METHOD` | `search_method` | `fuzzy` | Search algorithm (direct, fuzzy, fuzzy1, fuzzy3, default) |
| Preserve Order | `--preserve-order`, `-o` | `GMENU_PRESERVE_ORDER` | `preserve_order` | `false` | Keep original item order |
| Initial Query | `--initial-query`, `-q` | `GMENU_INITIAL_QUERY` | `initial_query` | `""` | Pre-filled search query |
| Auto Accept | `--auto-accept` | `GMENU_AUTO_ACCEPT` | `auto_accept` | `false` | Auto-select if only one match |
| Terminal Mode | `--terminal` | `GMENU_TERMINAL_MODE` | `terminal_mode` | `false` | Run in terminal-only mode |
| No Numeric Selection | `--no-numeric-selection` | `GMENU_NO_NUMERIC_SELECTION` | `no_numeric_selection` | `false` | Disable numeric shortcuts |
| Min Width | `--min-width` | `GMENU_MIN_WIDTH` | `min_width` | `600` | Minimum window width |
| Min Height | `--min-height` | `GMENU_MIN_HEIGHT` | `min_height` | `300` | Minimum window height |
| Max Width | `--max-width` | `GMENU_MAX_WIDTH` | `max_width` | `1920` | Maximum window width |
| Max Height | `--max-height` | `GMENU_MAX_HEIGHT` | `max_height` | `1080` | Maximum window height |
| Accept Custom Selection | (none) | `GMENU_ACCEPT_CUSTOM_SELECTION` | `accept_custom_selection` | `true` | Accept raw query when no match is selected |

Search method notes:
- `direct`: case-insensitive contains match (smart-case when query has uppercase).
- `fuzzy`: space-split tokens with a brute-force fuzzy matcher (min 2 consecutive chars).
- `fuzzy1`: sahilm/fuzzy scoring.
- `fuzzy3`: brute-force fuzzy variant.
- `default`: same behavior as `fuzzy`.

## Examples

### Using Config File
```yaml
# ~/.config/gmenu/config.yaml
title: "Project Selector"
prompt: "Select project:"
search_method: "fuzzy"
min_width: 800
min_height: 500
```

### Using Environment Variables
```bash
export GMENU_TITLE="Git Branch Selector"
export GMENU_PROMPT="Switch to branch:"
git branch | gmenu
```

### Using CLI Flags (Override Everything)
```bash
echo -e "option1\noption2\noption3" | gmenu --title "Quick Select" --prompt "Pick one:"
```

### Menu ID-Based Configuration Namespacing
You can create different configurations for different use cases using menu IDs:

```bash
# Create a config for git operations
mkdir -p ~/.config/gmenu/git
cat > ~/.config/gmenu/git/config.yaml << EOF
title: "Git Branch Selector"
prompt: "Switch to branch:"
search_method: "fuzzy"
auto_accept: true
EOF

# Create a config for file operations  
mkdir -p ~/.config/gmenu/files
cat > ~/.config/gmenu/files/config.yaml << EOF
title: "File Selector"
prompt: "Open file:"
search_method: "direct"
min_width: 800
EOF

# Use the configs
git branch | gmenu --menu-id git
find . -name "*.go" | gmenu --menu-id files
```

### Mixed Configuration
You can combine all three methods. For example:
- Set defaults in `~/.config/gmenu/config.yaml`
- Create specialized configs in `~/.config/gmenu/{menu-id}/config.yaml`
- Override some settings with environment variables for specific use cases
- Use CLI flags for one-off customizations

The configuration system will automatically merge all sources with the correct priority. 
