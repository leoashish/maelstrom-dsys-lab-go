# Challenge 4: Grow-Only Counter (CRDT) (Go)

**Difficulty:** builds directly on Challenge 3
**Concepts:** CRDTs, state-based replication, merge functions, convergence

## The task

```json
{"type": "add", "msg_id": 4, "delta": 3}   -> {"type": "add_ok", "in_reply_to": 4}
{"type": "read", "msg_id": 5}              -> {"type": "read_ok", "in_reply_to": 5, "value": 11}
```

Every `add` increases the counter by `delta`. A `read` on *any* node
should eventually return the sum of every delta added anywhere in the
cluster -- Maelstrom checks this with a final read on every node after
the test, confirming they all agree.

## The idea: one counter per node, summed

Don't maintain one shared integer -- that needs coordination on every
`add`. Instead:

- Give each node its own entry in a map: `counts = {"n1": 7, "n2": 3}`.
- `add` only ever touches **this node's own entry**. Always available.
- The effective value is `sum(counts)`.
- Periodically gossip your whole `counts` map to the other nodes, and
  **merge** what comes back: for each node_id, keep whichever value is
  larger.

"Keep the larger value" is what makes this correct: a node's own counter
never decreases, so between two values for the same node_id, the larger
one is always at least as recent -- merging by max converges to the same
final state no matter how redundantly or out-of-order nodes gossip.

## This is Challenge 3 again, wearing a different hat

Local state + periodic gossip + a merge function is exactly the shape of
Broadcast. The only thing that changed is *what* you're merging: a set
(merge = union) there, a per-node counter (merge = max) here -- and the
gossip here doesn't need topology-based neighbors or batching
acknowledgement tracking, since there's always something worth sending
(your whole current map) and nothing to lose by resending it.

## Same concurrency rule as Challenge 3

`add`, `read`, and incoming `replicate` messages all run on their own
goroutines, concurrently with your background gossip loop. Guard
`counts` with a `sync.Mutex`, same as before.

One new wrinkle: you need `n.NodeIDs()` to know who to gossip to, but
that's only populated once the `init` message arrives -- which hasn't
necessarily happened yet when `main()` starts registering handlers.
Register a handler for `"init"` itself (`n.Handle("init", ...)`) and
kick off your background goroutine from inside it -- see the comment in
`maelstrom/node.go`'s `handleInitMessage` for how that's wired up.

## What to edit

`main.go` -- the state, the three handlers, and the gossip goroutine.

## Running it

```sh
./test_1_quick.sh       # 3 nodes, no faults
./test_2_partition.sh   # the real test: network partitions
```

## Next

That's the core lab. See the top-level `README.md`'s "Going further"
section for what's next -- Kafka-style logs, a linearizable KV store,
transactions, and Raft all build on exactly this same workflow.
