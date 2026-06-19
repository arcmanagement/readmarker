# readmarker

A tiny, source-agnostic read-cursor ledger for AI agents.

`readmarker` tracks how far an agent has read across any number of communication
sources such as Slack, email, GitHub, and issue trackers, in one local key-value
store exposed through a single CLI. No server, no port, no daemon.

Built in Go and shipped as a single static binary. No runtime required, and
release binaries are checksummed and signed so you can verify them.

> Status: design locked, implementation not started. See
> `IMPLEMENTATION_PLAN.md`.

## Why

AI agents re-read the same threads every session, lose track of things they
meant to follow up on, and can't tell what's new since last time. The reason is
simple: bots have no read state. Slack's `last_read` belongs to human accounts,
and bot tokens never get one. Most other tools are the same or worse.

Existing tools don't fill this gap.

- Service-side "mark as read", such as MCP email servers, writes read state into
  each service. That's the opposite of what a cross-tool agent needs, and
  impossible when the bot has no read concept.
- Read-it-later apps such as Omnivore and Wallabag store the content and are
  heavy, server-backed applications.
- Read-state libraries such as `ledermann/unread` live inside one app's DB, with
  no cross-source support and no CLI.
- Per-service CLIs track one service only.

The closest prior art is hand-rolled `triage-index.jsonl` files inside specific
triage workflows. That proves the need is real, but nobody ships it as a
standalone, cross-source tool.

## What it is / isn't

readmarker holds only the cursor, meaning how far the agent has read. The other
two layers already exist.

| Layer | Who owns it |
|---|---|
| Content, context, what the agent did | an agent-memory store, already in session history |
| What's currently in the source | the source itself, such as the Slack API |
| How far the agent has read | readmarker — the gap nobody fills |

Non-goals: no unread tracking, no message-body storage, no touching
service-side read flags, no HTTP, no port, no daemon.

## Model

- `source_key` is an opaque string identifying one stream, such as
  `slack:<ws>:<channel>`, `chatwork:<room>`, or `github:<repo>#<n>`. readmarker
  never parses it, so new sources cost zero code.
- The cursor is a monotonically increasing integer, compared only within the
  same `source_key`. There is no cross-source comparison, so the caller picks
  whatever integer carries enough precision for that source. Slack can use the
  ts as microseconds, GitHub can use a comment id.

## CLI sketch

```
readmarker get     <source_key>           # print last read position
readmarker advance <source_key> <pos>     # move forward, atomic max
readmarker list    [--json]               # all source_key and positions
readmarker set     <source_key> <pos>     # force-write a position, recovery only
```

`advance` only ever moves forward, taking the max of current and new, so
re-reading never rewinds the cursor. `set` ignores that rule and is meant only
for fixing a mistaken position.

Unread detection lives outside readmarker. The agent opens a source, fetches the
current messages, calls `get`, keeps the messages newer than the cursor, then
calls `advance` after reading. readmarker supplies the position; the comparison
happens in the caller, typically a thin per-source adapter.

## Storage

bbolt, a pure-Go embedded key-value store, holds the source of truth in one
local file. A single bucket maps each `source_key` to its cursor position.
Transactions give atomic updates and crash safety, and there is no cgo, so the
binary stays single-file and cross-compiles cleanly.

The data is just a key to an integer, so a key-value store fits and SQL is
unnecessary. To inspect the ledger, use `readmarker list --json` rather than an
external DB tool.

## Roadmap

- thread-level granularity. Slack thread replies arrive out of band, after the
  channel's last ts.
- optional HTTP wrapper for remote and cloud agents
- per-source adapter helpers that turn a raw ts or id into the cursor integer
  and fetch the current latest

## Governance note

Branch-protection Rulesets are blocked while the repo is private on the current
plan. They will be applied when the repo flips to public in Phase 4. Until then,
there are no direct pushes to `main`; all changes go through a PR.
