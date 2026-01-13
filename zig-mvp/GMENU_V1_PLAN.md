# gmenu Zig v1 plan (feature + options inventory)

Purpose: capture the current Go gmenu feature set and configuration surface, then
propose a prioritized split and milestones for a Zig/AppKit v1 rewrite.

Sources (local):
- README.md, CONFIG.md, gmenu.yaml.example
- model/config.go, internal/config/model.go, internal/config/config.go
- pkg/config/load.go, internal/config/keys.go
- core/gmenu.go, core/gmenurun.go, core/gmenuitems.go, core/menu.go, core/search.go
- core/keyhandler.go, core/terminal.go, render/items.go, render/input.go, render/theme.go
- store/*

Important mismatches/gaps are called out in a dedicated section below so Zig v1
can pick a single source of truth.

## Feature inventory (current Go behavior)

### Input and modes
- Reads items from stdin (one per line). If no items are provided in GUI mode, it
  exits with an error.
- GUI mode (default) using Fyne: search input + list, outputs selected item to stdout.
- Terminal mode (`--terminal`): raw terminal input, live filtering, outputs the
  first match on Enter (simple contains match; not using the GUI search methods).

### Search and filtering
- Search methods map: "direct", "fuzzy", "fuzzy1", "fuzzy3", "default".
- "fuzzy" and "default" are tokenized by spaces and use a brute-force fuzzy
  search requiring 2 consecutive chars before fuzzy matching.
- "fuzzy1" uses sahilm/fuzzy scoring, sorted by score; "fuzzy3" is another
  brute-force variant.
- Preserve-order option keeps original item order for fuzzy matches.
- Result limit is capped at 10 items (hard-coded in core/menu.go).

### Selection and navigation
- Up/Down/Tab moves selection; wraps at ends.
- Enter selects; Escape cancels.
- Numeric shortcuts 1-9 select items when enabled.
- Mouse click selects an item.
- Focus loss acts like user cancel after a short grace period.
- Auto-accept: if enabled and exactly one match is present, selection is
  automatically completed.
- Custom selection: when enabled, Enter accepts the raw query even if there are
  no matches; otherwise it errors.

### UI rendering (GUI)
- Window title and prompt label are configurable.
- Input field supports Ctrl+L to clear and optional select-all on show.
- List renders alternating row stripes, selection highlight, optional icon hint,
  optional numeric hint, and an optional score column.
- Match counter label ("[matched/total]") updates on each search.
- Custom theme overrides sizes and selection colors.

### Config, CLI, and environment
- Priority order: CLI flags > env vars (GMENU_*) > config file.
- Menu ID namespaces config and cache paths.
- Config file name is `config.yaml`.
- `--init-config` writes a default config for the menu ID.
- Config keys accept snake_case and camelCase; invalid keys error.

### Persistence and single-instance
- Cache is stored under `~/.cache/gmenu/<menu_id>/cache.yaml`.
- Config is stored under `~/.config/gmenu/<menu_id>/config.yaml`.
- Cache stores last input, last entry, and timestamps.
- PID file in OS temp dir prevents multiple instances per menu ID.

### Embedding / dynamic items
- API supports SetItems / AppendItems / PrependItems with a channel for live
  updates (not exposed via CLI).

## Options inventory (current Go behavior)

Defaults from model.DefaultConfig(). CLI flags from internal/config/model.go.

| Config key | CLI flag | Env var | Default | Notes |
| --- | --- | --- | --- | --- |
| title | --title, -t | GMENU_TITLE | gmenu | Window title |
| prompt | --prompt, -p | GMENU_PROMPT | Search | Prompt text |
| menu_id | --menu-id, -m | GMENU_MENU_ID | "" | Namespaces config/cache/pid |
| search_method | --search-method, -s | GMENU_SEARCH_METHOD | fuzzy | See search methods above |
| preserve_order | --preserve-order, -o | GMENU_PRESERVE_ORDER | false | Keep original order for matches |
| initial_query | --initial-query, -q | GMENU_INITIAL_QUERY | "" | Pre-filled search text |
| auto_accept | --auto-accept | GMENU_AUTO_ACCEPT | false | Auto-select single match |
| terminal_mode | --terminal | GMENU_TERMINAL_MODE | false | Terminal-only mode |
| no_numeric_selection | --no-numeric-selection | GMENU_NO_NUMERIC_SELECTION | false | Disable numeric shortcuts + hints |
| min_width | --min-width | GMENU_MIN_WIDTH | 600 | Min window width |
| min_height | --min-height | GMENU_MIN_HEIGHT | 300 | Min window height |
| max_width | --max-width | GMENU_MAX_WIDTH | 1920 | Max window width |
| max_height | --max-height | GMENU_MAX_HEIGHT | 1080 | Max window height |
| accept_custom_selection | (no flag) | GMENU_ACCEPT_CUSTOM_SELECTION | true | Accept raw query if no match |
| init-config | --init-config | (none) | false | Writes config.yaml and exits |

## Mismatches and decisions to make for Zig v1

These appear inconsistent across docs vs code. Zig v1 should choose a single
source of truth and document it.

- Search method names: docs mention "exact" and "regex", but code implements
  "direct", "fuzzy", "fuzzy1", "fuzzy3", "default" and has no regex search.
- Config file name/paths: docs mention gmenu.yaml and current directory, but
  code loads `config.yaml` from ~/.config/gmenu and ~/.gmenu (menu-id scoped),
  not the current directory.
- Defaults: docs/examples show min_width/min_height 800/400 and max 0, while
  code defaults are 600/300 and 1920/1080.
- gmenu.yaml.example omits initial_query and uses a different filename.

## Zig v1 split (prioritized components)

P0 (must-have for v1 parity)
- cli: args parsing, stdin loading, stdout selection, exit codes.
- config: env + config.yaml + flags merge, menu-id namespaces, init-config.
- model: Config, MenuItem, ExitCode/ExitError.
- search: direct + fuzzy (tokenized) + preserve-order + result limit.
- gui: AppKit window, input field, list view, match counter.
- input: key handling (Up/Down/Tab, Enter, Esc), numeric shortcuts, Ctrl+L.

P1 (strongly preferred)
- persistence: cache (last input/entry) per menu ID.
- single-instance: pid file in temp dir per menu ID.
- auto-accept + accept custom selection logic.
- focus-loss behavior matching cancel semantics.

P2 (optional/secondary)
- terminal mode implementation.
- dynamic items API for embedding (SetItems/Append/Prepend).
- icon rendering + score metadata.
- theme customization (colors/sizes).

## Milestones (prioritized)

Milestone 0: Minimum UI + input loop
- Zig project layout, AppKit dependency wiring.
- Read stdin items and show them in a basic list UI.
- Basic search algorithm (simple contains match is enough for v0).
- Exit codes and Esc to cancel (basic key handling).

Milestone 1: Search + core model parity
- MenuItem model and search pipeline (direct + fuzzy tokenized).
- Preserve-order and result limit behavior (limit=10).
- Enter to select; output selection to stdout and exit cleanly.

Milestone 2: Interaction parity
- Numeric shortcuts 1-9 and numeric hints.
- Auto-accept logic and custom-selection acceptance.
- Focus loss cancellation behavior.
- Ctrl+L clear, select-all on focus.

Milestone 3: Persistence + single-instance
- Cache last input/entry; apply last input when menu-id set and no initial_query.
- PID file guard per menu ID.

Milestone 4: Config + CLI parity
- CLI arg parsing, env var mapping, config.yaml loading and validation.
- Menu-id path resolution and init-config writer.

Milestone 5: Nice-to-have parity
- Terminal mode, icon + score rendering, theme tuning, dynamic item updates.

## Suggested Zig v1 scope (if a cut is needed)

- Required: Milestones 0-3.
- Strongly preferred: Milestone 4.
- Optional: Milestone 5.
