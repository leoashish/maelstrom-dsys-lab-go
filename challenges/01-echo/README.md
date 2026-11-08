# Challenge 1: Echo (Go)

**Difficulty:** warm-up
**Concepts:** the Maelstrom request/response loop, nothing distributed yet

## The task

A client sends your node a message like:

```json
{"src": "c1", "dest": "n1", "body": {"type": "echo", "msg_id": 1, "echo": "Please echo 35"}}
```

Your node should reply:

```json
{"src": "n1", "dest": "c1", "body": {"type": "echo_ok", "in_reply_to": 1, "echo": "Please echo 35"}}
```

That's the whole job: whatever's in `echo`, send it straight back in an
`echo_ok` response.

## What to edit

Open `main.go` and fill in the `echo` handler registered with
`n.Handle("echo", ...)`. Everything else -- reading STDIN, writing
STDOUT, the init handshake, filling in `in_reply_to` -- is handled by
the vendored `maelstrom` package sitting in `./maelstrom/` (straight
from the official Maelstrom repo's Go demo). Skim `maelstrom/node.go`
once if you're curious how the plumbing works, but you shouldn't need
to touch it.

A `maelstrom.Message` carries its body as `json.RawMessage` (raw,
unparsed JSON bytes) -- `Unmarshal` it into a `map[string]any` (or a
struct, if you'd rather), edit the `"type"` field, and hand it to
`n.Reply(msg, body)`.

## Running it

```sh
./test.sh
```

This builds your code with `go build`, then asks Maelstrom to spin up
one copy of it, fire some echo requests, and check every response
matches. If it's right, you'll see:

```
Everything looks good! ヽ(‘ー`)ノ
```

If something's wrong, Maelstrom prints a JSON blob explaining which
operations failed (look for `:valid? false` and the `:errors` nearby).
You can add `--log-stderr` to the `maelstrom test` line in `test.sh` to
see Go's own `log.Printf` output (the vendored library already logs
every message it sends/receives) interleaved with Maelstrom's own log --
handy for every later challenge too.

## Next

Once this passes, move on to `../02-unique-ids/`.
