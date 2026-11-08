#!/usr/bin/env bash
# A gentle first check: 3 nodes, no faults. Get this passing first.
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
    --node-count 3

echo
echo "Check ':workload :final-reads' above -- every node should report the"
echo "same number, and it should fall inside ':acceptable'."
