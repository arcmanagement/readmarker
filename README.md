# readmarker

A tiny, source-agnostic read-cursor ledger for AI agents.

`readmarker` tracks how far an agent has read across any number of communication
sources such as Slack, email, GitHub, and issue trackers, in one local SQLite
ledger exposed through a single CLI. No server, no port, no daemon.

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

readmarker holds only the cursor. The other two layers already exist.

| Layer | Who owns it |
|---|---|
| Content, context, what the agent did | an agent-memory store, already in session history |
| What's currently in the source | the source itself, such as the Slack API |
| How far the agent has read | readmarker — the gap nobody fills |

Non-goals: no message-body storage, no touching service-side read flags, no
HTTP, no port, no daemon.

## Model

- `source_key` is an opaque string identifying one stream, such as
  `slack:<ws>:<channel>`, `chatwork:<room>`, or `github:<repo>#<n>`. readmarker
  never parses it, so new sources cost zero code.
- The cursor value is normalized to epoch ms, so cross-source comparison is just
  a numeric `>`. Converting a raw ts or id to epoch is the caller's job, which
  keeps readmarker source-agnostic.
- `status` is optional. `pending` means read but not yet acted on. This lets the
  ledger answer "read but un-handled", which a plain unread badge can't.

## CLI sketch

```
readmarker get     <source_key>                 # last read position
readmarker advance <source_key> <epoch_ms>      # move forward, monotonic max
readmarker check   <source_key> --latest <ms>   # is there unread? latest > cursor
readmarker list    [--unread]                   # all sources, or only ones with unread
readmarker mark    <source_key> --status pending
```

`advance` only ever moves forward, taking the max of current and new, so
re-reading never rewinds the cursor. For unread detection the caller passes the
source's current latest and readmarker compares. readmarker itself never
fetches.

## Storage

A single local SQLite file is the source of truth.

```sql
CREATE TABLE cursors (
  source_key   TEXT PRIMARY KEY,
  last_pos     INTEGER NOT NULL,   -- epoch ms
  last_read_at INTEGER NOT NULL,
  status       TEXT,               -- NULL, 'pending', ...
  note         TEXT
);
```

A remote HTTP layer can wrap this later without changing the source of truth.

## Roadmap

- thread-level granularity. Slack thread replies arrive out of band, after the
  channel's last ts.
- optional HTTP wrapper for remote and cloud agents
- per-source adapter helpers that convert a raw ts or id to epoch ms and fetch
  the latest

## Governance note

Branch-protection Rulesets are blocked while the repo is private on the current
plan. They will be applied when the repo flips to public in Phase 4. Until then,
there are no direct pushes to `main`; all changes go through a PR.
