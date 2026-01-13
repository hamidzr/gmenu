# zmenu

Native macOS AppKit MVP for the gmenu replacement (zmenu).

## What it does
- Reads menu items from stdin (one per line). If stdin is empty, it exits with a non-zero code.
- Opens a native window with a search field and a list of items.
- Typing filters the list with a tokenized, case-insensitive fuzzy match (results capped at 10).
- Enter prints the selected (default top) item to stdout and exits 0; Esc cancels with a non-zero exit code.
- Up/Down/Tab move the selection within the filtered list.
- Double-clicking a row accepts that item.

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
