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
    { "label": "Item label", "icon": "app" }
  ]
}
```

- `v` is the protocol version (currently `1`).
- `cmd` determines how items are applied.
- `items` is required for `set`, `append`, and `prepend`.
- `label` is required. `icon` is optional (app/file/folder/info).

## Behavior

- `set`: replaces all items.
- `append`: adds items to the end.
- `prepend`: adds items to the beginning.

zmenu re-filters the list after each update using the current query text.

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
{"v":1,"cmd":"append","items":[{"label":"hi"}]}
```
