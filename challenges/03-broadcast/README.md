# Challenge 3: Broadcast (Go)

**Difficulty:** the core of the lab
**Concepts:** gossip protocols, network topology, partition tolerance, message efficiency, real concurrent state

## The task

Three message types:

```json
{"type": "topology", "msg_id": 1, "topology": {"n1": ["n2", "n3"], "n2": ["n1"], "n3": ["n1"]}}
→ {"type": "topology_ok", "in_reply_to": 1}

{"type": "broadcast", "msg_id": 2, "message": 1000}
→ {"type": "broadcast_ok", "in_reply_to": 2}

{"type": "read", "msg_id": 3}
→ {"type": "read_ok", "in_reply_to": 3, "messages": [1000, 1001, ...]}
```

Clients send `broadcast` requests to *any* node, and expect that value to
eventually show up in a `read` from *every* node. `topology` arrives once,
right after init, telling you (optionally) which nodes are your
"neighbors."

## Why this is the centerpiece of the lab

This is the first challenge where nodes genuinely have to communicate to
do their job. It's also the first one where Go's concurrency model stops
being a footnote and becomes the main event -- see below before you
write any code.

## Build it in stages, not all at once

### Stage A -- `./test_a_single_node.sh` (1 node)
No peers exist yet. Just remember every message you're told about and
return the full set on `read`. Confirms your handlers and the harness
are wired up correctly.

### Stage B -- `./test_b_multi_node.sh` (5 nodes, no faults)
Now a message broadcast to `n3` needs to be visible when `n5` is read.
Watch for two classic bugs:

1. **Infinite loops.** If every node forwards every message it receives
   straight back to whoever sent it, messages bounce forever. Only
   forward a message the *first* time you see it.
2. **Who do you tell?** Every node in `n.NodeIDs()` (simple), or just
   the neighbors from `topology` (sparser, needed for stage D).

### Stage C -- `./test_c_fault_tolerant.sh` (5 nodes, `--nemesis partition`)
The network now splits and heals at random. A message broadcast mid-split
must still reach everyone once it heals -- track what you *haven't* had
acknowledged, and periodically retry it.

### Stage D -- `./test_d_efficient.sh` (25 nodes, higher rate, tree topology)
Check `:net :all :msgs-per-op`. Batch everything pending for a neighbor
into one message per round instead of one message per value.

## The concurrency story -- read this first

In the Python version of this lab, a single `asyncio` event loop handles
one thing at a time, so a plain `set` works as shared state without a
lock. **Go is different, and this challenge is where it bites.**

Look at `maelstrom/node.go`'s `Run()` method: every incoming message that
isn't a reply to your own RPC gets dispatched with `go func() { ... }()`
-- its own goroutine. In this challenge, that means:

- A `broadcast` handler and a `read` handler can run **at the same time**,
  on different OS threads, both touching your message set.
- Once you add background gossip (stage C+), you'll have one long-lived
  goroutine per neighbor, *also* reading that same set, concurrently with
  everything else.

A bare Go `map` is **not** safe for concurrent reads and writes -- the
runtime will often just panic with `fatal error: concurrent map read and
map write` under enough load (which, conveniently, is exactly what stage
D's higher rate is likely to surface if you haven't protected it yet).
Guard your shared state with a `sync.Mutex`: lock before you read or
write the set (or your per-neighbor tracking maps), unlock right after.
See `../../00-go-fundamentals/` for worked examples of exactly this
pattern, plus the fan-out-style "one goroutine per peer" structure stage
C/D's gossip loops want.

Once you have something passing stage B or C, it's worth a quick
`go build -race -o ./broadcast_race . ` and running one test through that
binary -- the race detector will tell you directly if you missed a lock,
rather than you finding out from a flaky test or a crash under load.

## What to edit

`main.go` -- the three handlers, your shared state, and (stage C+) a
background gossip goroutine per neighbor.

## Running it

```sh
./test_a_single_node.sh
./test_b_multi_node.sh
./test_c_fault_tolerant.sh
./test_d_efficient.sh
```

Run them in order.

## Next

Move on to `../04-grow-counter/`. Same shape -- shared state, background
gossip, a merge function -- for a counter instead of a set.
