---
name: use-readmarker
description: Use readmarker to track how far an AI agent has read in external sources such as Slack channels, GitHub issues or pull requests, email threads, Chatwork rooms, or issue trackers. Trigger when a task involves notification triage, inbox or thread follow-up, avoiding repeated reads across sessions, recording read cursors, comparing fetched items against the last read position, or advancing a local read ledger after reviewing source items.
---

# Use Readmarker

## Purpose

Use `readmarker` as a local read-cursor ledger. It stores only a
`source_key -> cursor` mapping. It does not fetch messages, store content, mark
anything read in the source service, or decide what is unread.

The caller owns source access and comparison logic:

1. Fetch current items from the source.
2. Read the stored cursor with `readmarker get`.
3. Keep only items newer than the cursor.
4. Review those items.
5. Advance the cursor after the review succeeds.

## Prerequisites

Check that the CLI exists before relying on it:

```bash
readmarker --version
```

If the command is missing, do not invent an install path. Tell the user to
install the CLI and, if useful, the official agent skill:

```bash
brew install --cask arcmanagement/readmarker/readmarker
npx skills add arcmanagement/readmarker --skill use-readmarker --agent codex --global
```

The skill installer owns agent-specific installation. Do not write directly to
`~/.codex/skills`, `~/.claude/skills`, or other agent skill directories unless
the user explicitly asks for manual installation.

## Source Keys

Choose a stable opaque `source_key` for each stream. readmarker never parses
the key, so keep enough namespace in the string for humans and future tools.

Good patterns:

```text
slack:<workspace>:<channel>
slack:<workspace>:<channel>:thread:<thread-ts>
github:<owner>/<repo>#<issue-or-pr-number>
github:<owner>/<repo>:notifications
email:<account>:inbox
email:<account>:thread:<thread-id>
chatwork:<room-id>
backlog:<space>:<project>:<issue-key>
```

Guidelines:

- Use one key per stream whose cursor is meaningful.
- Use thread-level keys when replies arrive outside the parent stream.
- Do not put message bodies, email addresses, secrets, access tokens, or personal
  contact details in the key.
- Avoid renaming keys casually; a rename starts a new cursor unless you migrate
  the old value with `get` and `set`.

## Cursor Values

Use a monotonically increasing non-negative base-10 integer. The value only has
meaning within one `source_key`.

Common choices:

- Slack message timestamp: convert `seconds.microseconds` to microseconds, for
  example `1712345678.123456 -> 1712345678123456`.
- GitHub issue or pull request comments: use the numeric comment id when you are
  only tracking comments.
- GitHub notifications or timeline items: use the source's stable numeric id if
  available; otherwise maintain the comparison outside readmarker and advance to
  the newest integer you control.
- Email: use IMAP UID for a folder, or internal date in milliseconds when the
  provider exposes a stable numeric timestamp.
- Issue trackers: use numeric comment ids, update ids, or timestamps converted
  to milliseconds or microseconds.

If the source only has string ids, keep a small adapter outside readmarker that
maps them to monotonic integers. Do not store JSON or structured state in
readmarker.

## Standard Workflow

Read the last cursor:

```bash
last=$(readmarker get "github:arcmanagement/readmarker#5")
```

Fetch source items with the relevant source tool or API, then filter items whose
numeric cursor is greater than `$last`.

After successfully reviewing all kept items, advance to the newest reviewed
cursor:

```bash
readmarker advance "github:arcmanagement/readmarker#5" 4756093926
```

`advance` is safe for repeated runs because it stores the max of the current and
new value. A stale session cannot move the cursor backward.

List known cursors when auditing state:

```bash
readmarker list
readmarker list --json
```

Use a task-local database only when you need isolation for tests, demos, or a
throwaway run:

```bash
readmarker --db /tmp/readmarker-demo.db get "slack:demo:alerts"
```

For durable agent workflows, prefer the default database or an explicit
`READMARKER_DB` chosen by the user.

## Recovery

Use `set` only for explicit correction. It can move the cursor backward or
forward, so do not use it in normal read flows.

```bash
readmarker set "github:arcmanagement/readmarker#5" 4756093926
```

Before using `set`, state why the correction is needed and what cursor is being
restored. If the source contains sensitive data, report only source keys and
cursor numbers, not message content.

## Reporting

When summarizing a readmarker-backed triage run, include:

- source keys checked
- previous cursor
- newest reviewed cursor
- whether `advance` changed the ledger
- any source that was skipped because cursor comparison was ambiguous

Do not claim source-side read state changed. readmarker only updates the local
ledger.
