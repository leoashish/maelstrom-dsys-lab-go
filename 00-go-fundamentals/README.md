# Chapter 0: Go Fundamentals Refresher

One consolidated chapter -- not split into separate sessions -- covering
language fundamentals, error handling, collections and generics, JSON
(the thing every Maelstrom message flows through), and concurrency: both
the raw primitives and the named patterns built from them. Read it,
run it, and move on to `../challenges/01-echo/`.

## Running it

```sh
go run .                # everything, in order, ~1 second
go run . concurrency     # just one section
go test ./...            # the test-first warm-up exercises (see below) -- these FAIL until you fill them in
go test -race ./...      # the real check for the SafeCounter exercise
```

Valid section names for the single-section form: `basics`, `structs`,
`errors`, `collections`, `json`, `concurrency`, `patterns`.

## What's in each section

| File | Section | Covers |
|---|---|---|
| `01_basics.go` | 1. Basics | variables & constants, type inference, `if`/`for`/`switch`, multi-return and named-return functions, variadics, closures, `defer`/`panic`/`recover` |
| `02_structs_interfaces.go` | 2. Structs & interfaces | structs, value vs. pointer receivers, embedding (composition, not inheritance), implicit interface satisfaction, `any`, type switches |
| `03_errors.go` | 3. Errors | the `error` interface, `fmt.Errorf` with `%w`, `errors.Is`/`errors.As`, custom error types, sentinel errors |
| `04_collections_generics.go` | 4. Collections & generics | arrays vs. slices, the shared-backing-array gotcha, maps and the comma-ok idiom, generic functions (`Map`/`Filter`/`Reduce`) and a generic `Stack[T]` type |
| `05_json_and_appdev.go` | 5. JSON & app dev | struct tags, `Marshal`/`Unmarshal`, why `map[string]any` turns every number into `float64`, `json.RawMessage` (exactly how `maelstrom.Message.Body` works), module/package layout, table-driven tests |
| `06_concurrency_basics.go` | 6. Concurrency basics | goroutines, unbuffered/buffered channels, closing + ranging, `select` (including non-blocking and timeout forms), `sync.WaitGroup`/`Mutex`/`Once`, a live data race next to its `sync/atomic` fix, `context` cancellation |
| `07_concurrency_patterns.go` | 7. Concurrency patterns | worker pools, fan-out/fan-in, pipelines, rate limiting with `time.Ticker`, the "done channel" cancellation pattern |
| `exercises.go` + `exercises_test.go` | warm-up | four small test-first exercises applying the above yourself |

## Why this order, and why these topics

Sections 1-4 are general-purpose Go. Section 5 exists specifically
because every challenge ahead of you talks JSON over STDIN/STDOUT --
`json.RawMessage` in particular is worth understanding before you open
`maelstrom/node.go` and wonder why `Message.Body` isn't already parsed.

Sections 6 and 7 get the most real estate on purpose. The Python version
of this lab (if you've seen it) gets away with un-locked shared state,
because `asyncio` only ever runs one coroutine at a time. **Go's
`maelstrom.Node` dispatches every incoming message on its own goroutine**
-- real concurrency, not cooperative scheduling -- which means the
broadcast and grow-counter challenges will only be correct if you
actually apply what's in section 6. That's not a hypothetical: the
unique-ids challenge's own test is run at a high enough rate to
*actually trigger* an unguarded counter's race condition, not just
theorize about it.

## The exercises

`exercises_test.go` is written already (four small, already-failing
tests). `exercises.go` has the TODOs. This mirrors the workflow you'll
use for the rest of the lab -- tests (here, Maelstrom itself) define
"done," you write code until they pass -- at a much smaller scale, and
with every technique you need already fully worked out somewhere in
sections 1-7:

- `SumInts` -- a plain loop (section 1).
- `Dedup[T comparable]` -- a generic function using a map as a set
  (sections 4's `Filter` + the comma-ok idiom).
- `SafeCounter` -- a `sync.Mutex`-protected counter (section 6's
  `safeCounter` example, promoted into a reusable type). Only
  `go test -race ./...` tells you for certain whether you actually fixed
  it -- the plain `go test` can pass by luck on a racy implementation.
- `ParallelSum` -- split work across goroutines, combine over a channel
  (section 7's worker pool / fan-out-fan-in).

No answer key is included on purpose -- if you're stuck, the pattern you
need is sitting fully worked in the section each TODO comment points at.

## Next

`../challenges/01-echo/` -- the lab itself.
