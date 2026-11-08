# Challenge 2: Unique ID Generation (Go)

**Difficulty:** easy correctness story, real concurrency trap
**Concepts:** availability vs. coordination, `sync/atomic` / `sync.Mutex`

## The task

Respond to `generate` requests with a globally unique `id`:

```json
{"type": "generate", "msg_id": 7}
→
{"type": "generate_ok", "in_reply_to": 7, "id": "anything, as long as it's unique"}
```

Maelstrom gathers every `id` returned by every node over the whole run and
fails the test if it finds a duplicate.

## Why this isn't trivial

The obvious instinct is "just use a UUID" or "ask a central server for
the next number." Both miss the point of the exercise. This test runs
with `--availability total --nemesis partition`: the network randomly
splits into groups that can't talk to each other, and **every** client
request must still succeed through that. A scheme that needs to ask
another node "has anyone used this id yet?" will stall the moment it's
cut off.

The instructive solution -- and the one worth building, since you'll
reuse the pattern in later challenges -- is to make collisions
*structurally impossible*: combine `n.ID()` (unique per node) with a
local counter (unique per call on that node). Zero network calls needed.

## The Go-specific trap

In the Python version of this lab, a single process handles one message
at a time (cooperative `asyncio`), so a bare `counter += 1` is safe
without a lock. **That is not true here.** Look at `maelstrom/node.go`'s
`Run()` method: every incoming message that isn't a reply gets
`go func() { ... }()` -- its own goroutine, scheduled onto any available
OS thread. At `--rate 1000`, two `generate` requests genuinely can be
handled at the same instant, on different threads.

A plain `counter++` is not one atomic step -- it's read, add one, write
back. If two goroutines both read `counter` as `5` before either writes
`6`, you've just handed out the same id to two different clients, and
the test will catch it as `:duplicated-count` greater than `0`. This
isn't a style nitpick; it's a real, triggerable bug.

Two ways to fix it, both covered in `../../00-go-fundamentals/`:
- **`sync/atomic`** -- `atomic.Int64` gives you a counter whose
  `Add(1)` is guaranteed indivisible, no matter how many goroutines call
  it at once. Usually the simplest fix for "just a counter."
  `go run ./00-go-fundamentals` for the Synchronization & Concurrency section.
- **`sync.Mutex`** -- lock around the read-modify-write by hand. More
  general (works for more than a single counter), slightly more
  ceremony.

## What to edit

`main.go` -- the state declaration and the `generate` handler.

## Running it

```sh
./test_1_quick.sh        # 3 nodes, no faults -- get this green first
./test_2_partition.sh    # the real test: partitions + total availability
./test_3_race_check.sh   # builds with Go's race detector -- confirms
                          # there's no unsynchronized access at all, not
                          # just that none happened to bite this run
```

Watch `:workload :duplicated-count` -- it must be `0`. Watch
`test_3_race_check.sh`'s output for a `DATA RACE` report -- there
shouldn't be one.

## Next

Move on to `../03-broadcast/` -- this is where nodes start actually
talking to each other, and where the concurrency story gets more
interesting still (several goroutines sharing one message set, plus
background gossip loops of their own).
