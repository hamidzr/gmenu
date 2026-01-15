# IPC-only selection mode requirements

## Goal
Add an IPC-only mode that ignores stdin and prints the full selected JSON item to stdout on accept. This lets external launchers (combo-switcher) map the selection exactly using a stable `id`.

## Functional requirements
1. **IPC schema extension**
   - Extend `ipc.Item` with `id: []const u8` (string).
   - Include `id` in JSON parsing and serialization.
   - Preserve any extra fields where feasible.

2. **New CLI/config flag**
   - Add a flag (name TBD, e.g. `--ipc-only` or `--ipc-echo`).
   - When enabled:
     - Ignore stdin input entirely.
     - Menu starts with an empty list and waits for IPC `set`/`append`/`prepend` updates.

3. **Selection output**
   - On accept, output the **full JSON item** to stdout on a single line.
   - The output should match the original IPC item payload (including `id`, `label`, `icon`, and any extra fields).
   - On cancel, exit non-zero with no stdout output.

4. **Backward compatibility**
   - Default behavior (stdin input -> selection label output) remains unchanged when the flag is not set.

5. **Documentation**
   - Update `IPC_PROTOCOL.md` to document `id` and the IPC-only selection mode.
   - Update `README.md` with flag usage and output example.

## Suggested output example
```json
{"id":"window:123","label":"Safari — Docs","icon":"app"}
```

## Files to touch
- `src/ipc.zig`
- `src/cli.zig`
- `src/config.zig`
- `src/menu.zig` or `src/app.zig`
- `IPC_PROTOCOL.md`
- `README.md`
