# zmenu Plan

## Goal
Milestone 0 MVP from `GMENU_V1_PLAN.md`:
- Read stdin items into memory and show them in a native AppKit list UI.
- Provide a search field that filters items with a simple case-insensitive substring match.
- Enter prints the top filtered item to stdout and exits 0; Esc cancels with a non-zero exit.

## Current gmenu observations (for eventual replacement)
- CLI wiring is in internal/cli/cli.go: reads stdin items, resolves config, and launches GUI/terminal mode.
- GUI setup is in core/gmenu.go: creates the Fyne app/window, search entry, items canvas, and match label; caches last input per menu ID.
- Key handling is in core/keyhandler.go: Up/Down/Tab navigate, Enter accepts, Escape cancels, numeric shortcuts 1-9 jump to items.
- Search entry behavior is in render/input.go: intercepts key events, supports Ctrl+L clear, and handles focus loss.
- Item rendering is in render/items.go: numbered list, selected highlight, alternating row colors, optional icon, and score metadata.
- Menu/search state lives in core/menu.go: search method, filtered list, selection index, dynamic updates via ItemsChan.

## M0 Implementation steps
1. Read stdin items (one per line) into memory; exit with a non-zero error if stdin is empty.
2. Build AppKit UI: window + search NSTextField + list view (NSTableView in a scroll view).
3. Wire live filtering on text change using a case-insensitive substring match.
4. Enter prints the top filtered item and exits 0; Esc cancels with a non-zero exit code.

## Validation
- `zig build run` with stdin opens a window, shows the list, and filters as you type.
- Pressing Enter prints the top filtered item to stdout and exits 0.
- Pressing Esc exits with a non-zero code.
- Running without stdin exits with an error.

## Follow-on for gmenu replacement (out of scope for M0)
- Menu list selection + output to stdout for explicit selection.
- Keyboard nav and shortcuts mirroring current behavior.
- Search/filter pipeline parity, state caching, and config handling.
- CLI wiring for terminal mode parity.
