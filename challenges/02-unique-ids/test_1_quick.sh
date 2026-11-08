#!/usr/bin/env bash
# A fast, gentle check: 3 nodes, no faults. Get this passing first.
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
go build -o ./unique_ids .

"$MAELSTROM_BIN" test \
    -w unique-ids \
    --bin "$SCRIPT_DIR/unique_ids" \
    --time-limit 15 \
    --rate 100 \
    --node-count 3
