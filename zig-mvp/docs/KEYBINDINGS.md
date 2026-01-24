# Keybinding Notes (Developer)

## Cmd+A select-all
AppKit does not always route Cmd+A through `keyDown:` for the search field. In this app it can surface as `noop:` via `control:textView:doCommandBySelector:`. To ensure select-all works, the handler is implemented in `performKeyEquivalent:` on `ZigSearchField`.
