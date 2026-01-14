# zmenu (Zig MVP) Tasks

## TODO

## Milestone 1: ✅ COMPLETE - Search + Core Model Parity
- [x] Read stdin items into memory
- [x] Native AppKit window with search field and list view
- [x] Tokenized fuzzy search (case-insensitive)
- [x] Direct/contains search method
- [x] Preserve-order option and result limit (10)
- [x] Enter prints top item to stdout and exits 0
- [x] Esc cancels with exit code 1
- [x] Exit with error if stdin is empty

## Milestone 2: ✅ COMPLETE - Interaction Parity
- [x] Add selection state tracking to AppState (current_index field)
- [x] Up/Down/Tab navigation through list
  - [x] Add keyDown: method to search field class
  - [x] Implement moveUp: handler (decrement selection, wrap at 0)
  - [x] Implement moveDown: handler (increment selection, wrap at end)
  - [x] Implement insertTab: handler (move down like arrow)
- [x] Visual selection highlight in table view
  - [x] Implement tableView:shouldSelectRow: delegate
  - [x] Track selected row in AppState
  - [x] Update onSubmit to use selected row instead of filtered[0]
- [x] Mouse click to select item
  - [x] Add tableViewSelectionDidChange: delegate method
  - [x] Update selection state on click
  - [x] Double-click or Enter on selected row to submit
- [x] Numeric shortcuts (1-9) to jump to items
  - [x] Intercept number keys in keyDown:
  - [x] Map key to filtered item index
  - [x] Output item immediately on number press
- [x] Numeric hints display (optional)
  - [x] Render numeric index column instead of prefixing labels
- [x] Ctrl+L to clear search field
  - [x] Add Cmd+L or Ctrl+L handler in keyDown:
  - [x] Clear text field and reset filter
- [x] Auto-accept when single match
  - [x] Check filtered.items.len == 1 after filter
  - [x] Exit immediately with that item
- [x] Custom selection (accept raw query when no match)
  - [x] When filtered.items.len == 0, output raw query text
  - [x] Make configurable (currently always exits with code 1)
- [x] Focus loss behavior (grace period + cancel)
  - [x] Add windowDidResignKey: delegate
  - [x] Timer for grace period before exit
- [x] Select-all on focus (optional)
  - [x] Add becomeFirstResponder override
  - [x] Call selectText: on text field

## Milestone 3: ✅ COMPLETE - Persistence & Single-Instance
- [x] Menu ID system for namespacing
- [x] Cache last input per menu ID (~/.cache/gmenu/<menu_id>/cache.yaml)
- [x] Cache last entry and timestamps
- [x] Restore last input on startup (when no initial_query)
- [x] PID file in OS temp dir per menu ID
- [x] Single-instance enforcement (prevent multiple instances)

## Milestone 4: ✅ COMPLETE - Config & CLI Parity
- [x] CLI argument parsing
  - [x] --title, -t (window title)
  - [x] --prompt, -p (prompt text)
  - [x] --menu-id, -m (namespace config/cache/pid)
  - [x] --search-method, -s (direct/fuzzy/fuzzy1/fuzzy3)
  - [x] --preserve-order, -o
  - [x] --initial-query, -q (pre-filled search)
  - [x] --auto-accept
  - [x] --no-numeric-selection
  - [x] --min-width, --min-height
  - [x] --max-width, --max-height
  - [x] --init-config (write default config and exit)
- [x] Environment variable support (GMENU_* prefix)
- [x] YAML config file loading
  - [x] ~/.config/gmenu/<menu-id>/config.yaml
  - [x] ~/.gmenu/<menu-id>/config.yaml
  - [x] XDG_CONFIG_HOME support
- [x] Config priority: CLI flags > env vars > config file
- [x] Config validation (snake_case and camelCase keys)
- [x] Multiple search method implementations
  - [x] fuzzy1 (sahilm/fuzzy equivalent scoring)
  - [x] fuzzy3 (brute-force variant)
- [x] Configurable window size constraints

## Milestone 5: 🎨 TODO - Nice-to-Have Features
- [x] Terminal mode (--terminal flag)
- [x] Match counter label ("[matched/total]")
- [x] Alternating row colors (zebra striping)
- [x] Icon hints display
- [x] Score metadata column (for debugging/tuning)
- [x] Theme customization (colors/sizes)
- [ ] Dynamic items API (for embedding)
  - [ ] SetItems
  - [ ] AppendItems
  - [ ] PrependItems
  - [ ] ItemsChan for live updates

## Documentation & Testing
- [x] Add `zig build test` step to build.zig
- [x] Document search method differences (docs mention "exact"/"regex", code has "direct"/"fuzzy"/"fuzzy1"/"fuzzy3")
- [x] Reconcile config file name inconsistency (docs: gmenu.yaml vs code: config.yaml)
- [x] Update default window size docs (docs: 800/400 vs code: 600/300)
- [x] Document migration path from Go gmenu to Zig zmenu
- [ ] Add visual regression tests for GUI changes
