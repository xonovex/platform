---
description: Produce a traceable revision from explicit feedback without overwriting the source
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  <subject> --feedback <feedback>... [--reference <reference>...]
  [--revision <revision>] [--kind <selection>] [--perspective <selection>]
  [--criteria <criteria>...] [--method <selection>]
  [--capability <selection>...] [--provider <selection>]
  [--result <destination-reference>] [--dry-run]
---

# /xonovex-workflow:revise — Revise

## Arguments

- `subject` (required): Inline content or one opaque provider-native reference.
- `--feedback` (repeatable, required): Inline feedback or opaque feedback reference.
- `--reference` (repeatable): Supporting inline content or opaque references.
- `--revision` (optional): Exact native source revision; required when provider
  context does not otherwise pin it.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, optional): Constraints the revision must retain.
- `--result` (optional): Explicit provider-native destination reference.
- `--dry-run` (optional): Preview persistence or source-adjacent edits.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Revise** operation with these arguments. The skill is the source of truth for the
procedure, output, and error handling; do not restate them.
