# zmenu

Native macOS AppKit MVP for the gmenu replacement (zmenu).

## What it does
- Reads menu items from stdin (one per line). If stdin is empty, it exits with a non-zero code.
- Opens a native window with a search field and a list of items.
- Typing filters the list with a tokenized, case-insensitive fuzzy match (results capped at 10).
- Enter prints the selected (default top) item to stdout and exits 0; Esc cancels with a non-zero exit code.
- Up/Down/Tab move the selection within the filtered list.
- Double-clicking a row accepts that item.
- Keys 1-9 accept the corresponding item when numeric selection is enabled.
- Ctrl+L clears the query.
- A pid file in the temp dir prevents multiple instances per menu id.
- When `menu_id` is set, the last query + selection are stored under `~/.cache/gmenu/<menu_id>/cache.yaml`.

## Requirements
- macOS
- Zig 0.15.2+ (zig-objc requires a released Zig)
- Xcode Command Line Tools (for AppKit headers)

## Run
Provide stdin, then launch the app:

```bash
printf "alpha\nbravo\ncharlie\n" | zig build run
```

If you run without stdin, zmenu exits with an error.

### Visual regression
`just visual` captures a snapshot of the UI using `scripts/visual_test.sh`. The first run writes
`samples/visual_baseline.png`; set `UPDATE_SNAPSHOT=1` to refresh the baseline. The script relies
on macOS Accessibility + Screen Recording permissions to capture the window.

### Config + CLI
Supported flags: `--menu-id/-m`, `--initial-query/-q`, `--search-method/-s`, `--preserve-order/-o`, `--auto-accept`, `--terminal`, `--follow-stdin`, `--ipc-only`, `--no-numeric-selection`, `--show-icons`, `--show-score`, `--title/-t`, `--prompt/-p`, `--min-width`, `--min-height`, `--max-width`, `--max-height`, `--row-height`, `--field-height`, `--padding`, `--numeric-column-width`, `--icon-column-width`, `--score-column-width`, `--alternate-rows`, `--background-color`, `--list-background-color`, `--field-background-color`, `--init-config`.
Supported env: `GMENU_MENU_ID`, `GMENU_INITIAL_QUERY`, `GMENU_SEARCH_METHOD`, `GMENU_PRESERVE_ORDER`, `GMENU_AUTO_ACCEPT`, `GMENU_TERMINAL_MODE`, `GMENU_FOLLOW_STDIN`, `GMENU_IPC_ONLY`, `GMENU_NO_NUMERIC_SELECTION`, `GMENU_SHOW_ICONS`, `GMENU_SHOW_SCORE`, `GMENU_ACCEPT_CUSTOM_SELECTION`, `GMENU_TITLE`, `GMENU_PROMPT`, `GMENU_MIN_WIDTH`, `GMENU_MIN_HEIGHT`, `GMENU_MAX_WIDTH`, `GMENU_MAX_HEIGHT`, `GMENU_ROW_HEIGHT`, `GMENU_FIELD_HEIGHT`, `GMENU_PADDING`, `GMENU_NUMERIC_COLUMN_WIDTH`, `GMENU_ICON_COLUMN_WIDTH`, `GMENU_SCORE_COLUMN_WIDTH`, `GMENU_ALTERNATE_ROWS`, `GMENU_BACKGROUND_COLOR`, `GMENU_LIST_BACKGROUND_COLOR`, `GMENU_FIELD_BACKGROUND_COLOR`.
Config file: searches `config.yaml` in `~/.config/gmenu[/<menu_id>]`, then `~/.gmenu[/<menu_id>]`, then `$XDG_CONFIG_HOME/gmenu` (macOS: `~/Library/Application Support/gmenu`) when present. Only `--menu-id` selects namespaced config paths. Use `--init-config` to write a default config to `~/.config/gmenu`.

Terminal mode is a simple prompt/response flow (non-live) that reads the query from `/dev/tty` and returns the top match. `--follow-stdin` keeps the GUI running and appends new stdin lines as they arrive. `--ipc-only` ignores stdin entirely and waits for IPC updates.
When `--show-icons` is enabled, input lines can prefix `[app]`, `[file]`, `[folder]`, or `[info]` to show icons.
`--row-height` and `--alternate-rows` adjust table density and zebra striping.
Theme colors accept hex strings like `#RRGGBB` or `#RRGGBBAA` (empty/`none`/`default` keeps system defaults). Size tuning is available via `field_height`, `padding`, and the column width settings.

### IPC + dynamic items
zmenu listens on a local Unix socket for dynamic item updates. Use `zmenuctl` to send
`set`, `append`, or `prepend` commands to a running instance:

```bash
printf "alpha\nbravo\n" | zmenuctl --menu-id demo set --stdin
zmenuctl --menu-id demo append "charlie"
```

Protocol details live in `IPC_PROTOCOL.md`.

IPC-only selection mode ignores stdin and prints the full JSON item on accept:

```bash
zmenu --menu-id demo --ipc-only
```

Example stdout:

```json
{"id":"window:123","label":"Safari — Docs","icon":"app"}
```

### Compatibility notes
- Search methods supported: `direct`, `fuzzy`, `fuzzy1`, `fuzzy3`, `default` (`default` matches `fuzzy`). Regex or `exact` modes are not implemented.
- The config filename is `config.yaml` in the standard gmenu config locations; `gmenu.yaml` is not read.
- Default window bounds are `600x300` with max `1920x1080` (override via config/env/flags).

### Migration from Go gmenu
- Copy your existing `config.yaml` into the same gmenu config locations; ensure `search_method` is one of the supported values above.
- If you previously used `exact` or `regex`, switch to `direct` or `fuzzy` since zmenu does not implement regex search.
- Numeric selection is opt-in by setting `no_numeric_selection: false` in config or env.
- Terminal mode is intentionally minimal and chooses the first match on Enter.
