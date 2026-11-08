#!/usr/bin/env bash
# Stage D: 25 nodes, higher request rate, a sparser tree topology. This
# checks Maelstrom's "messages per operation" stat, printed under :net.
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
    --node-count 25 \
    --time-limit 20 \
    --rate 100 \
    --topology tree4

echo
echo "Check ':net :all :msgs-per-op' in the output above. One network"
echo "message per broadcast per neighbor adds up fast at this scale --"
echo "batching unacknowledged messages into one send per neighbor per"
echo "round will bring this down a lot."
