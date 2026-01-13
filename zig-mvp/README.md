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

### Config + CLI
Supported flags: `--menu-id/-m`, `--initial-query/-q`, `--search-method/-s`, `--preserve-order/-o`, `--auto-accept`, `--no-numeric-selection`, `--title/-t`, `--prompt/-p`.
Supported env: `GMENU_MENU_ID`, `GMENU_INITIAL_QUERY`, `GMENU_SEARCH_METHOD`, `GMENU_PRESERVE_ORDER`, `GMENU_AUTO_ACCEPT`, `GMENU_NO_NUMERIC_SELECTION`, `GMENU_ACCEPT_CUSTOM_SELECTION`, `GMENU_TITLE`, `GMENU_PROMPT`.
Config file: searches `config.yaml` in `$XDG_CONFIG_HOME/gmenu[/<menu_id>]` (macOS: `~/Library/Application Support/gmenu`) and `~/.gmenu[/<menu_id>]`.
