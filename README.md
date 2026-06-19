# readmarker

A tiny, source-agnostic **read-cursor ledger for AI agents**.

`readmarker` tracks *how far an agent has read* across any number of
communication sources — Slack, email, GitHub, issue trackers — in one local
SQLite ledger, exposed through a single CLI. No server, no port, no daemon.

Built in **Go** and shipped as a single static binary — no runtime required,
and release binaries are checksummed/signed so you can verify them.

> Status: design locked, implementation not started. See
> [`IMPLEMENTATION_PLAN.md`](./IMPLEMENTATION_PLAN.md).

## Why

AI agents re-read the same threads every session, lose track of things they
meant to follow up on, and can't tell "what's new since last time" — because
**bots have no read state**. Slack's `last_read` belongs to *human* accounts;
bot tokens never get one. Most other tools are the same or worse.

Existing tools don't fill this gap:

- **Service-side "mark as read"** (MCP email servers, etc.) writes read state
  *into each service* — the opposite of what a cross-tool agent needs, and
  impossible when the bot has no read concept.
- **read-it-later apps** (Omnivore, Wallabag) store the *content* and are heavy,
  server-backed applications.
- **read-state libraries** (e.g. `ledermann/unread`) live inside one app's DB —
  no cross-source, no CLI.
- **per-service CLIs** (Slack CLIs, etc.) track one service only.

The closest prior art is hand-rolled `triage-index.jsonl` files inside specific
triage workflows — proof the need is real, but nobody ships it as a standalone,
cross-source tool.

## What it is / isn't

readmarker holds **only the cursor**. The other two layers already exist:

| Layer | Who owns it |
|---|---|
| Content, context, what the agent did | an agent-memory store (already in session history) |
| What's *currently* in the source | the source itself (Slack API, etc.) |
| **How far the agent has read** | **readmarker** ← the gap nobody fills |

**Non-goals:** no message-body storage, no touching service-side read flags,
no HTTP / port / daemon.

## Model

- **`source_key`** — opaque string identifying one stream:
  `slack:<ws>:<channel>`, `chatwork:<room>`, `github:<repo>#<n>`.
  readmarker never parses it; new sources cost zero code.
- **cursor value** — normalized to **epoch ms**, so cross-source comparison is
  just numeric `>`. Converting a raw ts / id → epoch is the *caller's* job
  (keeps readmarker source-agnostic).
- **`status`** — optional: `pending` (read but not yet acted on), etc. Lets the
  ledger answer "read but un-handled", which a plain unread badge can't.

## CLI (sketch)

```
readmarker get     <source_key>                 # last read position
readmarker advance <source_key> <epoch_ms>      # move forward (monotonic max)
readmarker check   <source_key> --latest <ms>   # is there unread? (latest > cursor)
readmarker list    [--unread]                   # all sources / only ones with unread
readmarker mark    <source_key> --status pending
```

`advance` only ever moves forward (`max(current, new)`), so re-reading never
rewinds the cursor. Unread detection = the caller passes the source's current
latest; readmarker compares. readmarker itself never fetches.

## Storage

Single local SQLite file is the source of truth.

```sql
CREATE TABLE cursors (
  source_key   TEXT PRIMARY KEY,
  last_pos     INTEGER NOT NULL,   -- epoch ms
  last_read_at INTEGER NOT NULL,
  status       TEXT,               -- NULL | 'pending' | ...
  note         TEXT
);
```

A remote HTTP layer can wrap this later without changing the source of truth.

## Roadmap

- thread-level granularity (Slack thread replies arrive out of band, after the
  channel's last ts)
- optional HTTP wrapper for remote / cloud agents
- per-source adapter helpers (raw ts/id → epoch ms, fetch-latest)

## Governance note

Branch-protection Rulesets are blocked while the repo is private on the current
plan; they will be applied when the repo flips to public (Phase 4). Until then:
no direct pushes to `main`, all changes via PR.
