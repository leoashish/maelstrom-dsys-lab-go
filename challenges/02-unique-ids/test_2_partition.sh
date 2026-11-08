#!/usr/bin/env bash
# The real test: high rate, network partitions, and --availability total
# means every single request must succeed, even while nodes can't reach
# each other. This is what actually proves your scheme needs no
# coordination -- and, in Go, that your counter is safe under concurrent
# access (see the concurrency note at the top of main.go).
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
    --time-limit 20 \
    --rate 1000 \
    --node-count 3 \
    --availability total \
    --nemesis partition

echo
echo "Look for :workload :duplicated-count 0 above -- any duplicates means"
echo "either your id scheme can collide across nodes, or your counter isn't"
echo "safe under concurrent access. See test_3_race_check.sh for a direct"
echo "way to check the latter."
