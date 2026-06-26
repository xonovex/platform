---
description: >-
  Walk open decisions one by one in plain text — files involved, explanation,
  pros, cons, recommendation — capturing the user's call on each before moving
  to the next
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Task
  - Write
  - Skill
argument-hint: "[input-file] [--save-to <file>]"
---

# /xonovex-workflow:plan-clarify — Clarify Open Decisions One by One

## Arguments

- `input-file` (optional): Path to a research file or plan document. Auto-detects: research findings in the current conversation first, then the most recent `plans/*.md`
- `--save-to <file>` (optional): Save the consolidated agreed-direction summary to a file

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-clarify** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
