#!/usr/bin/env bash
# Stage A: one node, no peers. Confirms your plumbing works before any
# actual distribution is involved.
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
go build -o ./broadcast .

"$MAELSTROM_BIN" test \
    -w broadcast \
    --bin "$SCRIPT_DIR/broadcast" \
    --node-count 1 \
    --time-limit 10 \
    --rate 10
