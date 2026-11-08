# Reference solutions -- spoilers

Everything in this folder is a complete, working solution to the
corresponding challenge, each verified against the real Maelstrom
harness (and, where concurrency is the whole point, against Go's race
detector too).

**Use these to get unstuck, not as a starting point.** If you're stuck:

1. Re-read the README in the challenge's own folder -- the "traps to
   watch for" and concurrency sections describe the most common failure
   modes on purpose.
2. Add `--log-stderr` to the `maelstrom test` command in the relevant
   `test*.sh` script, and sprinkle `log.Printf(...)` calls in your code
   (the vendored `maelstrom` package already logs every message it
   sends/receives this way). Watching what your node actually thinks is
   happening is usually more illuminating than reading someone else's
   code.
3. If something is failing intermittently or only under load, try
   `go build -race -o ./yourbinary .` and run that through the test
   instead -- the race detector often finds the exact bug directly,
   rather than you inferring it from a flaky symptom.
4. Only then, look here -- and try reading just enough to get unstuck
   (e.g. one function), rather than copying the whole file.

## What's here

| Folder | Challenge | Verified result |
|---|---|---|
| `01-echo/` | 01-echo | passes `test.sh` |
| `02-unique-ids/` | 02-unique-ids | 0 duplicates under `--nemesis partition --availability total`; confirmed race-free with `go build -race` |
| `03-broadcast/` | 03-broadcast | passes all 4 stages, including the 25-node efficiency run; confirmed race-free with `go build -race` |
| `04-grow-counter/` | 04-grow-counter | all nodes converge to the same final read under `--nemesis partition` |

Each is its own Go module (its own `go.mod`, with the `maelstrom` package
vendored locally), buildable on its own:

```sh
cd 02-unique-ids
go build -o ./unique_ids .
../../maelstrom/maelstrom test -w unique-ids --bin ./unique_ids \
    --time-limit 20 --rate 1000 --node-count 3 --availability total --nemesis partition
```

(adjust the path to the `maelstrom` binary if you're running from
somewhere else -- see the top-level README).
