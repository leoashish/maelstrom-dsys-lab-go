#!/usr/bin/env bash
# Installs the prerequisites for this lab and downloads the Maelstrom
# binary. Safe to re-run -- it skips anything already present.
#
# What this installs:
#   - A JDK (Maelstrom is a Clojure program that runs on the JVM)
#   - Graphviz (for anomaly diagrams) and Gnuplot (for latency/throughput
#     plots) -- optional but recommended
#   - The Maelstrom release tarball itself, into ./maelstrom/ at the repo
#     root (next to this scripts/ folder)
#
# This script does NOT install Go -- most systems already have it; see
# the README if you need to (or https://go.dev/doc/install). Go 1.21+ is
# required (this lab's code uses Go 1.22, but 1.21 should also work for
# everything except the per-iteration loop variable note in Chapter 0).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MAELSTROM_DIR="$REPO_ROOT/maelstrom"
MAELSTROM_VERSION="v0.2.4"
MAELSTROM_URL="https://github.com/jepsen-io/maelstrom/releases/download/${MAELSTROM_VERSION}/maelstrom.tar.bz2"

echo "== Checking prerequisites =="

PKGS_NEEDED=()

if command -v java >/dev/null 2>&1; then
    echo "[ok] java found: $(java -version 2>&1 | head -1)"
else
    echo "[missing] java (JDK 11+)"
    PKGS_NEEDED+=("openjdk-17-jdk-headless")
fi

if command -v go >/dev/null 2>&1; then
    echo "[ok] go found: $(go version)"
else
    echo "[missing] go"
    echo "    -> install Go 1.21+ yourself (not handled by this script):"
    echo "       https://go.dev/doc/install"
    echo "       (or on Debian/Ubuntu: apt install golang-go)"
fi

if command -v dot >/dev/null 2>&1; then
    echo "[ok] graphviz (dot) found"
else
    echo "[missing] graphviz"
    PKGS_NEEDED+=("graphviz")
fi

if command -v gnuplot >/dev/null 2>&1; then
    echo "[ok] gnuplot found"
else
    echo "[missing] gnuplot"
    PKGS_NEEDED+=("gnuplot-nox")
fi

if [ "${#PKGS_NEEDED[@]}" -gt 0 ]; then
    echo
    echo "== Installing missing packages: ${PKGS_NEEDED[*]} =="
    if command -v apt-get >/dev/null 2>&1; then
        if [ "$(id -u)" -eq 0 ]; then
            apt-get update -qq
            apt-get install -y "${PKGS_NEEDED[@]}"
        else
            sudo apt-get update -qq
            sudo apt-get install -y "${PKGS_NEEDED[@]}"
        fi
    elif command -v brew >/dev/null 2>&1; then
        BREW_PKGS=()
        for p in "${PKGS_NEEDED[@]}"; do
            case "$p" in
                openjdk-17-jdk-headless) BREW_PKGS+=("openjdk@17") ;;
                gnuplot-nox) BREW_PKGS+=("gnuplot") ;;
                *) BREW_PKGS+=("$p") ;;
            esac
        done
        brew install "${BREW_PKGS[@]}"
    else
        echo "Don't recognize your package manager. Please install manually: ${PKGS_NEEDED[*]}"
        echo "(On Debian/Ubuntu: apt install openjdk-17-jdk-headless graphviz gnuplot)"
        echo "(On macOS: brew install openjdk@17 graphviz gnuplot)"
    fi
else
    echo "[ok] all system packages present"
fi

echo
echo "== Installing Maelstrom $MAELSTROM_VERSION =="
if [ -x "$MAELSTROM_DIR/maelstrom" ]; then
    echo "[ok] already present at $MAELSTROM_DIR"
else
    mkdir -p "$MAELSTROM_DIR"
    TARBALL="$REPO_ROOT/maelstrom.tar.bz2"
    echo "Downloading $MAELSTROM_URL ..."
    curl -fL -o "$TARBALL" "$MAELSTROM_URL"
    echo "Extracting..."
    tar xjf "$TARBALL" -C "$REPO_ROOT"
    rm -f "$TARBALL"
    chmod +x "$MAELSTROM_DIR/maelstrom"
    echo "[ok] installed to $MAELSTROM_DIR"
fi

if command -v go >/dev/null 2>&1; then
    echo
    echo "== Sanity check =="
    echo "Building and running the reference echo solution for 5 seconds..."
    (cd "$REPO_ROOT/reference_solutions/01-echo" && go build -o ./echo .)
    "$MAELSTROM_DIR/maelstrom" test -w echo \
        --bin "$REPO_ROOT/reference_solutions/01-echo/echo" \
        --time-limit 5 --node-count 1 | tail -5
fi

echo
echo "All set. Next: go run ./00-go-fundamentals (if Go is fresh in your"
echo "head, skim it), then open challenges/01-echo/README.md."
echo "Browse past results any time with: $MAELSTROM_DIR/maelstrom serve  (then open http://localhost:8080)"
