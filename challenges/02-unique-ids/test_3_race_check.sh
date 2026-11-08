#!/usr/bin/env bash
# Builds with Go's race detector (-race) instrumented in, then runs the
# same partition test. The race detector watches every memory access at
# runtime and reports an error the instant two goroutines touch the same
# memory without synchronization -- even if the output still happens to
# come out correct on this particular run. This is the most direct way
# to confirm your counter really is safe, rather than inferring it from
# "no duplicates appeared this time."
#
# Race-detector builds are slower and use more memory, which is why this
# is a separate script instead of baked into every test run.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MAELSTROM_BIN="$REPO_ROOT/maelstrom/maelstrom"

if [ ! -x "$MAELSTROM_BIN" ]; then
    echo "Maelstrom isn't installed yet. Run: bash $REPO_ROOT/scripts/setup.sh" >&2
    exit 1
fi

cd "$SCRIPT_DIR"
echo "Building with -race..."
go build -race -o ./unique_ids_race .

"$MAELSTROM_BIN" test \
    -w unique-ids \
    --bin "$SCRIPT_DIR/unique_ids_race" \
    --time-limit 15 \
    --rate 500 \
    --node-count 3

echo
echo "If the race detector found anything, it prints a 'DATA RACE' report"
echo "to stderr (visible with --log-stderr added above) naming the exact"
echo "two goroutines and lines involved -- that's your bug, found directly,"
echo "not inferred from a symptom."
