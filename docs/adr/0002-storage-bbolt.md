# 0002 — Storage: bbolt

- Status: Accepted
- Date: 2026-06-19

## Context

readmarker stores a mapping from source_key to a cursor integer. There are no
relational queries; reads are get-one and list-all, writes are an atomic
forward-update. The tool must stay a single binary with a small dependency
footprint.

## Decision

Use go.etcd.io/bbolt, a pure-Go embedded key-value store, with one bucket
mapping source_key to its cursor position.

## Alternatives considered

- modernc.org/sqlite. Pure Go and no cgo, but SQL is unnecessary here and its
  dependency tree is heavy. Its only real edge was letting a human inspect the
  file with sqlite3 or DB Browser, and we decided direct human editing is not a
  needed use case. Inspection is covered by list --json.
- mattn/go-sqlite3. The C SQLite via cgo, fastest and most battle-tested, but
  cgo breaks single-binary cross-compilation, which contradicts ADR 0001.

## Consequences

- Pure Go, no cgo, so cross-compilation stays trivial.
- Tiny dependency footprint, which matters for a widely trusted tool.
- Transactions give atomic updates and crash safety.
- The DB file is not editable with a generic external tool, so recovery is done
  through the set command rather than poking the file directly.
