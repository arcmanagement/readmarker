# 0004 — No status field

- Status: Accepted
- Date: 2026-06-19

## Context

We considered a status field, for example pending meaning read but not yet acted
on, so the ledger could answer "did I handle this?". That is appealing but it is
a different axis from "how far have I read".

## Decision

No status field. readmarker holds only the cursor.

## Alternatives considered

- A status or task flag per source_key. Rejected as task-management territory,
  separate from the read position. It widens the surface and pulls readmarker
  toward being a todo tool.

## Consequences

- Minimal surface: the value stored is just the cursor integer.
- Tracking whether something was handled is left to the caller or a separate
  tool.
- If a real need appears later, status can be added without breaking the cursor
  model, but it must clear the same bar: tied to the core value, not just nice
  to have.
