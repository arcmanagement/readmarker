# Implementation Plan — readmarker

Design is locked. See `README.md` for the model and the competitive gap. Stack
decided: Go with bbolt. This document is the build plan.

## Stack decided: Go + bbolt

A single static binary is the deciding factor. A CLI tool should ship as a
binary.

- Tamper-evident. Ship with checksums and signatures so users can verify the
  binary matches the author's build.
- No runtime dependency. Node and Python aren't required, so the distributable
  range is widest: `brew`, direct download, or `go install`.
- Fast startup for a CLI that agents hit frequently.
- Small supply-chain surface. Dependencies are compiled in, no install scripts
  run, and nothing is resolved at runtime.

Storage is `go.etcd.io/bbolt`, a pure-Go embedded key-value store.

- Pure Go and no cgo, so cross-compilation stays a matter of changing `GOOS`.
- Tiny dependency footprint, which matters for a tool meant to be widely
  trusted.
- The data is just a `source_key` to an integer, so a key-value store fits and
  SQL is unnecessary.
- Transactions give atomic updates and crash safety.

Rejected: `modernc.org/sqlite`, because SQL is unnecessary here and its
dependency tree is heavy. Its only real edge was letting a human poke the file
with sqlite3, and we decided that isn't needed. The cgo `mattn/go-sqlite3`,
because cgo breaks single-binary cross-compilation. Node/TS and Python, because
both require a runtime and carry an interpreted-package supply chain.

## Phases

- Phase 0, scaffold: `go.mod`, `go.etcd.io/bbolt`, CI for `go vet` / `go test` /
  `go build`, plus GoReleaser for checksummed and signed release binaries.
- Phase 1, core: a bbolt store with one bucket mapping `source_key` to a cursor
  integer. `get`, `advance` with atomic max inside a read-write transaction,
  `list`, and `set` for recovery. DB path via a `--db` flag or env var,
  defaulting under XDG or `Application Support`.
- Phase 2, CLI: arg parsing with the stdlib `flag`, keeping the dependency
  surface minimal. Stable exit codes, and a `--json` output mode for `list`.
- Phase 3, tests with `go test`: every subcommand, `advance` never rewinds,
  `set` overrides, concurrent `advance` stays atomic, and an unknown
  `source_key` reads as not-yet-read.
- Phase 4, release: finalize README and usage examples, set up the GoReleaser
  pipeline with checksums and signature, flip the repo to public, then apply
  branch-protection Rulesets. Rulesets are blocked while private on the current
  plan, so until then it stays PR-only with no direct push to `main`.

## Out of scope for v1

- unread tracking. readmarker stores only the cursor. The caller compares it
  against the source's current latest.
- message-body storage, which lives in the agent-memory store
- status and task flags
- thread-level granularity, since the channel-level cursor comes first
- HTTP and daemon modes

## Recovery

There is no automatic rollback. `advance` only moves forward. If a cursor is
advanced by mistake, fix it with `set`, which force-writes a position. bbolt
files can't be edited with a generic external tool, so `set` is the recovery
path.

## Future

- thread-level cursors, since Slack replies arrive out of band after the
  channel's last ts
- optional HTTP wrapper for remote and cloud agents
- per-source adapter helpers that turn a raw ts or id into the cursor integer
  and fetch the current latest

## Build handoff

Implementation runs in a separate repo session or subagent, since brain stops at
the plan. Phase 1 onward is a clean, well-scoped task now that storage and the
CLI surface are fixed.
