# zig-mvp Plan

## Goal
- POC: native macOS AppKit window with a single text input; on Enter, print input to stdout.
- Use Zig + zig-objc with the smallest long-term-friendly scaffold.

## Current gmenu observations (for eventual replacement)
- CLI wiring is in internal/cli/cli.go: reads stdin items, resolves config, and launches GUI/terminal mode.
- GUI setup is in core/gmenu.go: creates the Fyne app/window, search entry, items canvas, and match label; caches last input per menu ID.
- Key handling is in core/keyhandler.go: Up/Down/Tab navigate, Enter accepts, Escape cancels, numeric shortcuts 1-9 jump to items.
- Search entry behavior is in render/input.go: intercepts key events, supports Ctrl+L clear, and handles focus loss.
- Item rendering is in render/items.go: numbered list, selected highlight, alternating row colors, optional icon, and score metadata.
- Menu/search state lives in core/menu.go: search method, filtered list, selection index, dynamic updates via ItemsChan.

## POC Implementation steps
1. Create Zig project layout in zig-mvp (build.zig, build.zig.zon, src/main.zig).
2. Add zig-objc dependency and link AppKit + Foundation frameworks.
3. Build a minimal AppKit window (NSApplication, NSWindow).
4. Add an NSTextField with target/action; implement an Objective-C method on a Zig-defined class to print to stdout.
5. Add README with build/run instructions and notes.

## Validation
- zig build run shows a native window, accepts text, and prints input to stdout when Enter is pressed.

## Follow-on for gmenu replacement (out of scope for this POC)
- Menu list UI (items canvas + selection highlight + icons + scores).
- Keyboard nav and shortcuts mirroring current behavior.
- Search/filter pipeline, state caching, and config parity.
- CLI wiring for stdin + terminal mode parity.
