#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
INPUT="$ROOT/samples/visual_input.txt"
OUT_DIR="$ROOT/zig-out/visual"
ACTUAL="$OUT_DIR/actual.png"
BASELINE="$ROOT/samples/visual_baseline.png"

mkdir -p "$OUT_DIR"

if [ ! -f "$INPUT" ]; then
  echo "missing sample input: $INPUT" >&2
  exit 1
fi

zig build

( cat "$INPUT" | zig build run -- --title "zmenu-visual" --initial-query "al" --min-width 600 --min-height 300 ) &
PID=$!

sleep 0.8

WINDOW_ID=$(osascript -e 'tell application "System Events" to tell process "zmenu" to get the id of front window' 2>/dev/null || true)
if [ -z "$WINDOW_ID" ]; then
  echo "unable to locate zmenu window (grant Accessibility + Screen Recording permissions)" >&2
  kill "$PID" 2>/dev/null || true
  wait "$PID" 2>/dev/null || true
  exit 1
fi

screencapture -l "$WINDOW_ID" "$ACTUAL"

osascript -e 'tell application "System Events" to keystroke (ASCII character 27)' 2>/dev/null || true
kill "$PID" 2>/dev/null || true
wait "$PID" 2>/dev/null || true

if [ "${UPDATE_SNAPSHOT:-}" = "1" ] || [ ! -f "$BASELINE" ]; then
  cp "$ACTUAL" "$BASELINE"
  echo "updated baseline: $BASELINE"
  exit 0
fi

if cmp -s "$ACTUAL" "$BASELINE"; then
  echo "visual snapshot matches"
  exit 0
fi

echo "visual snapshot mismatch: $ACTUAL" >&2
exit 1
