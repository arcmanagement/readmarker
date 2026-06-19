# Implementation Plan — readmarker

Design is locked (see [`README.md`](./README.md) for the model and the
competitive gap). Stack **decided: Go**. This document is the build plan.

## Stack: Go (decided)

A single static binary is the deciding factor — a CLI tool should ship as a
binary:

- **tamper-evident** — ship with checksums / signatures so users can verify the
  binary matches the author's build. A mutable `node_modules` (plain, runtime-
  loaded JS) can't be verified the same way and is easy to alter unnoticed.
- **no runtime dependency** — no Node / Python required, so the distributable
  range is widest (`brew`, direct download, `go install`).
- **fast startup** (~5ms) for a CLI that agents hit frequently.
- **small supply-chain surface** — dependencies compiled in, no install scripts,
  no runtime module resolution.
- SQLite via **`modernc.org/sqlite`** (pure Go, no cgo) → trivial
  cross-compilation (mac / linux / win via `GOOS`).

Rejected: **Node/TS** & **Python** (runtime required + interpreted-package
supply chain); **Rust** (overkill for an I/O-bound tool, higher dev cost — Go's
startup and binary size are already enough for this scope).

## Phases

- **Phase 0 — scaffold**: `go.mod`, `modernc.org/sqlite`, CI (`go vet` /
  `go test` / `go build`), GoReleaser for checksummed + signed release binaries.
- **Phase 1 — core**: SQLite schema + cursor store. `get` / `advance` (monotonic
  max) / `check` / `list` / `mark`. Source of truth = one local SQLite file;
  path via `--db` flag / env var, default under XDG / `Application Support`.
- **Phase 2 — CLI**: arg parsing (`cobra` or stdlib `flag`), **stable exit
  codes** (`check` exits non-zero when unread exists, so it composes in shell
  pipelines), `--json` output mode.
- **Phase 3 — tests** (`go test`): every subcommand; `advance` never rewinds;
  epoch cross-source comparison; `status` transitions; unknown / empty
  `source_key`.
- **Phase 4 — release**: finalize README + usage examples, GoReleaser pipeline
  (checksums + signature), flip the repo to **public**, then apply
  branch-protection Rulesets (blocked while private on the current plan — until
  then, PR-only, no direct push to `main`).

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

Implementation runs in a separate repo session / subagent (brain stops at the
plan). Phase 1 onward is a clean, well-scoped task now that the stack is fixed.
