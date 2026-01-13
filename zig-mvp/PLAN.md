# zmenu Plan

## Goal
Milestone 1 MVP from `GMENU_V1_PLAN.md`:
- Read stdin items into memory and show them in a native AppKit list UI.
- Provide a search field that filters items with a tokenized fuzzy match (case-insensitive).
- Preserve-order option and result limit behavior (limit=10).
- Enter prints the selected (default top) item to stdout and exits 0.
- Up/Down/Tab move the selection; Esc cancels with a non-zero exit.

## Current gmenu observations (for eventual replacement)
- CLI wiring is in internal/cli/cli.go: reads stdin items, resolves config, and launches GUI/terminal mode.
- GUI setup is in core/gmenu.go: creates the Fyne app/window, search entry, items canvas, and match label; caches last input per menu ID.
- Key handling is in core/keyhandler.go: Up/Down/Tab navigate, Enter accepts, Escape cancels, numeric shortcuts 1-9 jump to items.
- Search entry behavior is in render/input.go: intercepts key events, supports Ctrl+L clear, and handles focus loss.
- Item rendering is in render/items.go: numbered list, selected highlight, alternating row colors, optional icon, and score metadata.
- Menu/search state lives in core/menu.go: search method, filtered list, selection index, dynamic updates via ItemsChan.

## Current zmenu structure
- `src/main.zig` entry point that delegates to the AppKit runner.
- `src/app.zig` AppKit wiring, callbacks, window/layout, and event handling.
- `src/menu.zig` menu model, filtered indices, and selection state.
- `src/search.zig` search pipeline + tests.
- `src/config.zig` defaults for window/text/search options.
- `src/pid.zig` pid file guard for single-instance behavior.
- `src/cache.zig` cache load/save for last query + selection.
- `src/cli.zig` CLI + env + config file merging.
- `src/terminal.zig` terminal-mode prompt flow.

Planned follow-on modules: config/env parsing, cache helpers, CLI flags, and theming.

## M1 Implementation steps
1. Read stdin items (one per line) into memory; exit with a non-zero error if stdin is empty.
2. Build AppKit UI: window + search NSTextField + list view (NSTableView in a scroll view).
3. Wire live filtering on text change using a tokenized fuzzy match.
4. Enforce preserve-order behavior and a hard limit of 10 results.
5. Enter prints the top filtered item and exits 0; Esc cancels with a non-zero exit code.

## Validation
- `zig build run` with stdin opens a window, shows the list, and filters as you type.
- Pressing Enter prints the top filtered item to stdout and exits 0.
- Pressing Esc exits with a non-zero code.
- Running without stdin exits with an error.
- `zig test src/search.zig` passes.

## Follow-on for gmenu replacement (out of scope for M1)
- Menu list selection + output to stdout for explicit selection.
- Keyboard nav and shortcuts mirroring current behavior.
- Search/filter pipeline parity, state caching, and config handling.
- CLI wiring for terminal mode parity.
