#!/usr/bin/env bash
# Builds the Go binary, then runs the echo challenge against the real
# Maelstrom harness.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MAELSTROM_BIN="$REPO_ROOT/maelstrom/maelstrom"

if [ ! -x "$MAELSTROM_BIN" ]; then
    echo "Maelstrom isn't installed yet. Run: bash $REPO_ROOT/scripts/setup.sh" >&2
    exit 1
fi

cd "$SCRIPT_DIR"
echo "Building..."
go build -o ./echo .

"$MAELSTROM_BIN" test \
    -w echo \
    --bin "$SCRIPT_DIR/echo" \
    --node-count 1 \
    --time-limit 10

echo
echo "Results, logs and plots were written under: $SCRIPT_DIR/store/"
echo "Browse them with: $MAELSTROM_BIN serve   (then open http://localhost:8080)"
