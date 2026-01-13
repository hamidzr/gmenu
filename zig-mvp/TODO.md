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

## Milestone 2: ⏳ TODO - Interaction Parity
**Current limitation**: Enter always outputs first filtered item, no selection state tracked

- [ ] Add selection state tracking to AppState (current_index field)
- [ ] Up/Down/Tab navigation through list
  - [ ] Add keyDown: method to search field class
  - [ ] Implement moveUp: handler (decrement selection, wrap at 0)
  - [ ] Implement moveDown: handler (increment selection, wrap at end)
  - [ ] Implement insertTab: handler (move down like arrow)
- [ ] Visual selection highlight in table view
  - [ ] Implement tableView:shouldSelectRow: delegate
  - [ ] Track selected row in AppState
  - [ ] Update onSubmit to use selected row instead of filtered[0]
- [ ] Mouse click to select item
  - [ ] Add tableViewSelectionDidChange: delegate method
  - [ ] Update selection state on click
  - [ ] Double-click or Enter on selected row to submit
- [ ] Numeric shortcuts (1-9) to jump to items
  - [ ] Intercept number keys in keyDown:
  - [ ] Map key to filtered item index
  - [ ] Output item immediately on number press
- [ ] Numeric hints display (optional)
  - [ ] Prepend "1. ", "2. ", etc. to table view labels
- [ ] Ctrl+L to clear search field
  - [ ] Add Cmd+L or Ctrl+L handler in keyDown:
  - [ ] Clear text field and reset filter
- [ ] Auto-accept when single match
  - [ ] Check filtered.items.len == 1 after filter
  - [ ] Exit immediately with that item
- [ ] Custom selection (accept raw query when no match)
  - [ ] When filtered.items.len == 0, output raw query text
  - [ ] Make configurable (currently always exits with code 1)
- [ ] Focus loss behavior (grace period + cancel)
  - [ ] Add windowDidResignKey: delegate
  - [ ] Timer for grace period before exit
- [ ] Select-all on focus (optional)
  - [ ] Add becomeFirstResponder override
  - [ ] Call selectText: on text field

## Milestone 3: 📋 TODO - Persistence & Single-Instance
- [ ] Menu ID system for namespacing
- [ ] Cache last input per menu ID (~/.cache/gmenu/<menu_id>/cache.yaml)
- [ ] Cache last entry and timestamps
- [ ] Restore last input on startup (when no initial_query)
- [ ] PID file in OS temp dir per menu ID
- [ ] Single-instance enforcement (prevent multiple instances)

## Milestone 4: ⚙️ TODO - Config & CLI Parity
- [ ] CLI argument parsing
  - [ ] --title, -t (window title)
  - [ ] --prompt, -p (prompt text)
  - [ ] --menu-id, -m (namespace config/cache/pid)
  - [ ] --search-method, -s (direct/fuzzy/fuzzy1/fuzzy3)
  - [ ] --preserve-order, -o
  - [ ] --initial-query, -q (pre-filled search)
  - [ ] --auto-accept
  - [ ] --no-numeric-selection
  - [ ] --min-width, --min-height
  - [ ] --max-width, --max-height
  - [ ] --init-config (write default config and exit)
- [ ] Environment variable support (GMENU_* prefix)
- [ ] YAML config file loading
  - [ ] ~/.config/gmenu/<menu-id>/config.yaml
  - [ ] ~/.gmenu/<menu-id>/config.yaml
  - [ ] XDG_CONFIG_HOME support
- [ ] Config priority: CLI flags > env vars > config file
- [ ] Config validation (snake_case and camelCase keys)
- [ ] Multiple search method implementations
  - [ ] fuzzy1 (sahilm/fuzzy equivalent scoring)
  - [ ] fuzzy3 (brute-force variant)
- [ ] Configurable window size constraints

## Milestone 5: 🎨 TODO - Nice-to-Have Features
- [ ] Terminal mode (--terminal flag)
- [ ] Match counter label ("[matched/total]")
- [ ] Alternating row colors (zebra striping)
- [ ] Icon hints display
- [ ] Score metadata column (for debugging/tuning)
- [ ] Theme customization (colors/sizes)
- [ ] Dynamic items API (for embedding)
  - [ ] SetItems
  - [ ] AppendItems
  - [ ] PrependItems
  - [ ] ItemsChan for live updates

## Documentation & Testing
- [ ] Add `zig build test` step to build.zig
- [ ] Document search method differences (docs mention "exact"/"regex", code has "direct"/"fuzzy"/"fuzzy1"/"fuzzy3")
- [ ] Reconcile config file name inconsistency (docs: gmenu.yaml vs code: config.yaml)
- [ ] Update default window size docs (docs: 800/400 vs code: 600/300)
- [ ] Document migration path from Go gmenu to Zig zmenu
- [ ] Add visual regression tests for GUI changes
