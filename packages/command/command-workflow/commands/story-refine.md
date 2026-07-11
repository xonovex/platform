---
description: "Draft: refine user stories against INVEST, split too-big ones vertically (SPIDR / splitting flowchart), and complete their acceptance criteria"
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Bash
  - Skill
argument-hint: "[stories-or-feature] [--out <dir>]"
---

# /xonovex-workflow:story-refine — Refine user stories into ready items

## Arguments

- `[stories-or-feature]` (optional): Path/URL to stories, or a feature name (auto-detects `stories/<feature>/` or reads the prompt)
- `[--out <dir>]` (optional): Output directory for refined stories (default `stories/<feature>/`)

## Delegation

Load the `user-stories-guide` skill (plugin `xonovex-skill-user-stories`) — and
`bdd-guide` for Given-When-Then acceptance-criteria format — and follow them to refine
each story: check it against INVEST, split any that fail "Small" vertically (SPIDR or the
splitting flowchart, never horizontally), identify the walking skeleton, and write
acceptance criteria covering happy path, boundaries, errors, and UI states. The skills
are the source of truth for the method — do not restate them.

Write refined stories to the output directory. If a work-item-tracker host skill is
installed (any `xonovex-skill-<host>` exposing work-item sync), additionally create/update
the tracker items; otherwise write to disk only and note which host skill enables sync.
