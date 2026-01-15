# zmenu IPC protocol (v1)

zmenu listens on a local Unix domain socket and accepts framed JSON messages.
The API is intended for updating the in-memory item list while the GUI is
running. Unknown fields are ignored for forward compatibility.

## Socket path

The socket lives in the OS temp directory:

- `$TMPDIR/zmenu.<menu_id>.sock` when `--menu-id` (or config `menu_id`) is set
- `$TMPDIR/zmenu.sock` when no menu id is provided

`$TMPDIR` falls back to `$TMP`, `$TEMP`, then `/tmp`.

## Framing

Each message is sent as:

```
<length>\n<json-payload>
```

Where `<length>` is the decimal byte count of the JSON payload.

## JSON schema

```json
{
  "v": 1,
  "cmd": "set" | "append" | "prepend",
  "items": [
    { "id": "item-id", "label": "Item label", "icon": "app" }
  ]
}
```

- `v` is the protocol version (currently `1`).
- `cmd` determines how items are applied.
- `items` is required for `set`, `append`, and `prepend`.
- `id` is a stable identifier string (recommended for IPC-only mode).
- `label` is required. `icon` is optional (app/file/folder/info).
- Extra fields on item objects are preserved for IPC-only output.

## Behavior

- `set`: replaces all items.
- `append`: adds items to the end.
- `prepend`: adds items to the beginning.

zmenu re-filters the list after each update using the current query text.

## IPC-only selection mode

When `--ipc-only` (or config `ipc_only: true`) is set, zmenu ignores stdin and
starts with an empty list. On accept, it prints the full JSON item to stdout on a
single line, preserving any extra item fields. Cancels exit non-zero without
stdout output.

## Size limits

Messages larger than 1MB are ignored.

## Examples

Using `zmenuctl`:

```bash
printf "alpha\nbravo\n" | zmenuctl --menu-id demo set --stdin
zmenuctl --menu-id demo append "charlie"
```

Or raw protocol:

```text
46
{"v":1,"cmd":"append","items":[{"id":"hi","label":"hi"}]}
```
