# 0003 — readmarker does not track unread

- Status: Accepted
- Date: 2026-06-19

## Context

The whole point is that an agent can mechanically tell which messages it hasn't
read. The tempting move is to make readmarker compute or store the unread set.
But deciding what is unread needs each source's current latest position, which
only the source knows, and readmarker does not fetch.

## Decision

readmarker stores only the cursor, meaning how far the agent has read. It does
not store the source's latest, an unread list, or any unread state. Unread
detection is the caller's job: open the source, fetch the current messages, get
the cursor, keep what is newer, then advance after reading.

## Alternatives considered

- Store a last-seen latest alongside the cursor, so list --unread could be
  self-contained. Rejected because it adds state, couples readmarker to a notion
  of current, and risks a stored unread view drifting from reality.
- A check --latest helper that compares a passed-in latest to the cursor.
  Dropped to keep the surface minimal; the caller can compare just as easily.

## Consequences

- readmarker stays a pure cursor store.
- "Do not manage unread" and "mechanically detect unread" coexist: the caller
  compares fresh source state against the cursor, so there is never a stale
  unread list.
- The CLI does not have check or list --unread.
