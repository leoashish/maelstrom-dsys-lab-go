#!/usr/bin/env bash
# The real test: network partitions. Your merge-by-max logic (not just
# your gossip loop) is what gets exercised here.
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
go build -o ./counter .

"$MAELSTROM_BIN" test \
    -w g-counter \
    --bin "$SCRIPT_DIR/counter" \
    --time-limit 20 \
    --rate 10 \
    --node-count 3 \
    --nemesis partition

echo
echo "All final reads should still agree once the partitions heal. If they"
echo "don't, double check your merge keeps the MAX per node_id, not the sum."
