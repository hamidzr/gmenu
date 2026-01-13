# zmenu

Native macOS AppKit proof-of-concept for the gmenu replacement (zmenu).

## What it does
- Opens a native window with a single text field.
- Press Enter to print the current text to stdout.

## Requirements
- macOS
- Zig 0.15.2+ (zig-objc requires a released Zig)
- Xcode Command Line Tools (for AppKit headers)

## Run
```bash
zig build run
```

If you want to see stdout, launch from a Terminal.
