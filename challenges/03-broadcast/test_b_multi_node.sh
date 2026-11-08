#!/usr/bin/env bash
# Stage B: five nodes, no faults. Storing locally is no longer enough --
# nodes must actually tell each other about new messages.
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
    --node-count 5 \
    --time-limit 20 \
    --rate 10

echo
echo "If you see :lost-count greater than 0, some nodes never heard about"
echo "a message -- you likely need to propagate it to peers, not just store it."
