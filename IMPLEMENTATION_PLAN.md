# Implementation Plan — readmarker

Design is locked (see [`README.md`](./README.md) for the model and the
competitive gap). This document is the build plan, opened for review **before
any code lands**.

## Proposed stack

Recommendation: **Node.js + TypeScript + `better-sqlite3`**.

| Option | Pros | Cons |
|---|---|---|
| **Node/TS + better-sqlite3** (recommended) | matches arcm's Vitest standard; `npx readmarker` for zero-install use; synchronous SQLite API is ideal for a short-lived CLI; strong OSS ecosystem | requires a Node runtime |
| Go | single static binary, no runtime | not the arcm stack; SQLite needs cgo / modernc |
| Python | stdlib `sqlite3`, no deps | distribution via pipx; off arcm's TS standard |
| bash + `sqlite3` CLI | tiniest footprint | weak tests / maintainability |

**Decision pending owner approval.** Everything below assumes the recommended
stack; swapability is high since the surface is tiny.

## Phases

- **Phase 0 — scaffold** (after stack approval): `package.json`, `tsconfig.json`,
  `vitest.config.ts`, `biome.json`, CI (`lint` / `typecheck` / `test` / `build`).
  This is the deferred bootstrap Step 4/5.
- **Phase 1 — core**: SQLite schema + cursor store. `get` / `advance` (monotonic
  max) / `check` / `list` / `mark`. Source of truth = one local SQLite file;
  path via `--db` flag / env var, default under XDG / `Application Support`.
- **Phase 2 — CLI**: arg parsing, **stable exit codes** (`check` exits non-zero
  when unread exists, so it composes in shell pipelines), `--json` output mode.
- **Phase 3 — tests** (Vitest): every subcommand; `advance` never rewinds;
  epoch cross-source comparison; `status` transitions; unknown / empty
  `source_key`.
- **Phase 4 — docs & publish**: finalize README, add usage examples (Slack /
  GitHub adapters shown as *examples*, not built in), flip the repo to
  **public**, then apply branch-protection Rulesets (blocked while private on
  the current plan — until then, PR-only, no direct push to `main`).

## Out of scope (v1)

- message-body storage (lives in the agent-memory store)
- fetching "latest" from any source — the caller passes it in
- thread-level granularity (channel-level cursor first)
- HTTP / daemon

## Future

- thread-level cursors (Slack replies arrive out of band)
- optional HTTP wrapper for remote / cloud agents
- per-source adapter helpers (raw ts/id → epoch ms, fetch-latest)

## Build handoff

Implementation runs in a separate repo session / subagent (brain stops at plan
+ scaffold review). Phase 1 onward is a clean, well-scoped task once the stack
is approved.
