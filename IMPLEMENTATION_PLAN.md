# Implementation Plan — readmarker

Design is locked. See `README.md` for the model and the competitive gap. Stack
decided: Go. This document is the build plan.

## Stack decided: Go

A single static binary is the deciding factor. A CLI tool should ship as a
binary.

- Tamper-evident. Ship with checksums and signatures so users can verify the
  binary matches the author's build. A mutable `node_modules`, which is plain
  runtime-loaded JS, can't be verified the same way and is easy to alter
  unnoticed.
- No runtime dependency. Node and Python aren't required, so the distributable
  range is widest: `brew`, direct download, or `go install`.
- Fast startup, around 5ms, for a CLI that agents hit frequently.
- Small supply-chain surface. Dependencies are compiled in, no install scripts
  run, and nothing is resolved at runtime.
- SQLite via `modernc.org/sqlite`, which is pure Go and needs no cgo, so
  cross-compilation to mac, linux, and windows is trivial via `GOOS`.

Rejected: Node/TS and Python, because both require a runtime and carry an
interpreted-package supply chain. Rust too, because it's overkill for an
I/O-bound tool and costs more dev time, while Go's startup and binary size are
already enough for this scope.

## Phases

- Phase 0, scaffold: `go.mod`, `modernc.org/sqlite`, CI for `go vet` / `go test`
  / `go build`, plus GoReleaser for checksummed and signed release binaries.
- Phase 1, core: SQLite schema and cursor store. `get`, `advance` with monotonic
  max, `check`, `list`, `mark`. Source of truth is one local SQLite file, with
  path via `--db` flag or env var, defaulting under XDG or `Application Support`.
- Phase 2, CLI: arg parsing with `cobra` or stdlib `flag`, stable exit codes so
  `check` exits non-zero when unread exists and composes in shell pipelines, and
  a `--json` output mode.
- Phase 3, tests with `go test`: every subcommand, `advance` never rewinds,
  epoch cross-source comparison, `status` transitions, and unknown or empty
  `source_key`.
- Phase 4, release: finalize README and usage examples, set up the GoReleaser
  pipeline with checksums and signature, flip the repo to public, then apply
  branch-protection Rulesets. Rulesets are blocked while the repo is private on
  the current plan, so until then it stays PR-only with no direct push to
  `main`.

## Out of scope for v1

- message-body storage, which lives in the agent-memory store
- fetching the latest from any source, which the caller passes in instead
- thread-level granularity, since the channel-level cursor comes first
- HTTP and daemon modes

## Future

- thread-level cursors, since Slack replies arrive out of band after the
  channel's last ts
- optional HTTP wrapper for remote and cloud agents
- per-source adapter helpers that convert a raw ts or id to epoch ms and fetch
  the latest

## Build handoff

Implementation runs in a separate repo session or subagent, since brain stops at
the plan. Phase 1 onward is a clean, well-scoped task now that the stack is
fixed.
