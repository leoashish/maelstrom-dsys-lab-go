#!/usr/bin/env bash
# Stage C: same as B, but the network now splits apart and heals at
# random (--nemesis partition).
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
    --rate 10 \
    --nemesis partition

echo
echo "If messages come back as :lost even though stage B passed, a"
echo "fire-and-forget send is probably getting dropped mid-partition --"
echo "you likely need to retry or periodically resend anything unconfirmed."
