# 0001 — Language: Go

- Status: Accepted
- Date: 2026-06-19

## Context

readmarker is a CLI hit frequently by AI agents and meant to be distributed
widely as Public OSS. A CLI tool should ship as a binary so users can verify it
and run it without a runtime.

## Decision

Implement readmarker in Go.

## Alternatives considered

- Node/TS. Distribution leans on npx, which requires a Node runtime and pulls a
  package supply chain into every install. The shipped artifact is plain JS,
  hard to verify against the author's build.
- Python. stdlib covers a lot, but distribution still needs a Python runtime via
  pip or pipx, off the single-binary goal.
- Rust. Single binary and fast, but overkill for an I/O-bound tool of this size,
  with higher development cost. Go's startup and binary size are already enough.

## Consequences

- Single static binary, checksummed and signed, so it is tamper-evident.
- No runtime dependency, widest distribution via brew, direct download, or
  go install.
- Fast startup for a tool agents call often.
- Cross-compilation is a matter of changing GOOS, as long as cgo is avoided,
  which ties into ADR 0002.
