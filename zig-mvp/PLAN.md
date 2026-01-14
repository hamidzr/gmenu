# zmenu Plan

## Goal
Parity with the Go gmenu feature set captured in `GMENU_V1_PLAN.md`.

Status: Milestones 1-5 are complete (GUI, search, interaction, persistence, config/CLI, and nice-to-haves).
Documentation and a basic visual snapshot helper are also in place.

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
- `src/config.zig` defaults for window/text/search/theme options.
- `src/pid.zig` pid file guard for single-instance behavior.
- `src/cache.zig` cache load/save for last query + selection + timestamp.
- `src/cli.zig` CLI + env + config file merging.
- `src/terminal.zig` terminal-mode prompt flow.
Dynamic updates are supported via `--follow-stdin` (polls stdin and appends items), with set/prepend/append
helpers available on the menu model.

## Validation
- `zig build run` with stdin opens a window, shows the list, and filters as you type.
- Pressing Enter prints the selected item to stdout and exits 0.
- Pressing Esc exits with a non-zero code.
- Running without stdin exits with an error.
- `zig build test` runs search tests.
- `just visual` captures a snapshot (requires Accessibility + Screen Recording permissions).

## Follow-on ideas (if needed)
- Release packaging + notarization for macOS.
- Richer theming (fonts, selection colors, row separators).
- Expand dynamic update protocol beyond follow-stdin.
