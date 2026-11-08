# Distributed Systems Lab -- Go track (built on Maelstrom)

A hands-on path through [Maelstrom](https://github.com/jepsen-io/maelstrom)
-- Kyle Kingsbury/Jepsen's workbench for learning distributed systems by
building toy implementations of them -- in Go. This is a sibling to the
Python version of this lab: same challenges, same Maelstrom harness,
same "write it yourself, get real feedback" workflow, but using Go's
official helper library and leaning into what's genuinely different
about writing this kind of system in Go: **real concurrency**, not
cooperative single-threaded scheduling.

## Start here: Chapter 0

**[`00-go-fundamentals/`](00-go-fundamentals/)** -- a single consolidated
refresher chapter (language basics, structs & interfaces, error
handling, collections & generics, JSON, and concurrency -- both the
primitives and the named patterns) plus a few small test-first warm-up
exercises. If Go is fresh in your head, skim it in a couple of minutes
with `go run .`. If it's been a while, read every comment. Either way,
it's worth doing before Challenge 1 -- the concurrency section in
particular sets up exactly what the later challenges need.

## Then: the four challenges

| # | Challenge | What it teaches |
|---|---|---|
| 1 | [Echo](challenges/01-echo/) | the request/response loop itself -- no distribution yet |
| 2 | [Unique ID Generation](challenges/02-unique-ids/) | coordination-free design under partitions, **and** a real concurrency trap (a racy counter can actually produce duplicate ids here) |
| 3 | [Broadcast](challenges/03-broadcast/) | gossip protocols, topology, partition tolerance, efficiency -- with genuinely concurrent handlers and background goroutines sharing state |
| 4 | [Grow-Only Counter](challenges/04-grow-counter/) | CRDTs -- state-based replication with a merge function, same concurrency discipline as Broadcast |

Each challenge folder has a stub `main.go` with TODOs, a README with the
task and the traps to watch for, and one or more `test*.sh` scripts that
build your code and run it against the real Maelstrom harness.
`reference_solutions/` has verified-working answers for when you're
stuck (gated behind a README that nudges you toward debugging first).

## What's different from the Python version

If you've also got the Python track: the biggest conceptual difference
isn't syntax, it's **concurrency**. Python's `asyncio` runs one
coroutine at a time, so a plain `set` or `dict` is safe shared state
without a lock. **Go's `maelstrom.Node` dispatches every incoming
message on its own goroutine** -- real parallelism, on real OS threads.
That means every challenge past Echo needs a `sync.Mutex` (or
`sync/atomic`) around its shared state, and the unique-ids challenge in
particular is run at a high enough rate to *actually produce duplicate
ids* if you get this wrong -- not a theoretical concern. Chapter 0's
concurrency sections, and the "concurrency note" at the top of each
challenge's `main.go`, exist because of this.

## How this is built

Every challenge is its own Go module (its own `go.mod`). Right inside it
sits a `maelstrom/` package -- Go's official helper library, vendored
straight from
[the Maelstrom repo's Go demo](https://github.com/jepsen-io/maelstrom/tree/main/demo/go),
so each challenge builds standalone with zero external module fetches.
It handles the JSON-RPC boilerplate: reading STDIN, writing STDOUT, the
`init`/`init_ok` handshake, message IDs, matching replies to requests.
You register a handler per message type with `n.Handle(type, fn)`, reply
with `n.Reply(msg, body)`, and call another node with `n.SyncRPC(ctx,
dest, body)` or fire-and-forget with `n.Send(dest, body)`. You shouldn't
need to edit anything under `maelstrom/` -- the interesting decisions
live in each challenge's `main.go`.

## Setup

You'll need:
- **Go 1.21+** -- check with `go version` (this lab's code targets 1.22;
  see Chapter 0's note on loop-variable semantics if you're on 1.21)
- **A JDK, 11+** -- Maelstrom itself is a Clojure program on the JVM
- **Graphviz** and **Gnuplot** (optional, but you'll want them -- for the
  anomaly diagrams and latency/throughput plots under `store/.../`)

Then run:

```sh
bash scripts/setup.sh
```

This checks for the above, installs what it can via `apt` or `brew`
(Go itself is not installed by this script -- see
[go.dev/doc/install](https://go.dev/doc/install) if you need it), and
downloads+extracts the Maelstrom release tarball into `./maelstrom/` at
the repo root (not checked into this project -- it's ~70MB, fetched
fresh). It finishes with a build-and-run smoke test of the echo
reference solution.

## Working through it

```sh
go run ./00-go-fundamentals      # refresh, ~1 second
cd challenges/01-echo
cat README.md
# edit main.go
./test.sh
```

Work through the challenges in order. On failure, Maelstrom prints a
results blob ending in `Analysis invalid! (ﾉಥ益ಥ)ﾉ ┻━┻` -- look for
`:valid? false` and the nearby `:errors` / workload stats (`:lost`,
`:duplicated-count`). Add `--log-stderr` to any `maelstrom test` command
to see your own `log.Printf(...)` calls interleaved with Maelstrom's log
(the vendored library already logs every send/receive this way).

If something only fails under load, or intermittently, try building with
`go build -race -o ./yourbinary .` and running *that* through the test
-- Go's race detector often names the exact bug directly. Every
challenge from 2 onward is worth checking this way at least once.

Every run writes results, logs and plots under `store/<timestamp>/`
inside that challenge's own folder. Browse them with
`./maelstrom/maelstrom serve`, then open <http://localhost:8080>.

## Going further

Once the four core challenges pass, `./maelstrom/doc/` (after `setup.sh`
has downloaded it) covers a lot more ground with the same workflow:
`lin-kv`/`seq-kv` services, a Kafka-style append-only log, a
linearizable KV store and transactions, and -- the big one -- building
real Raft consensus from scratch (`doc/06-raft/`), tested against
`--nemesis kill` and `--nemesis pause`, not just partitions. The
`maelstrom` package and `test.sh` pattern used here carries over
directly.

## Troubleshooting

- **"Maelstrom isn't installed yet"** -- run `bash scripts/setup.sh`.
- **`undefined: maelstrom.NewNode` or similar** -- make sure you're
  running `go build`/`go run` from inside the challenge's own folder
  (each is a separate module); the vendored `maelstrom/` package is
  local to each one.
- **A handler returns a `NotSupported` RPCError** -- that's the TODO you
  haven't filled in yet; expected until you implement it.
- **`fatal error: concurrent map read and map write`, or a flaky test
  that sometimes shows duplicates/lost messages** -- a synchronization
  bug almost always; see "What's different from the Python version"
  above, and try a `-race` build.
