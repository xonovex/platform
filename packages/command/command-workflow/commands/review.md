---
description: Review an exact subject against explicit criteria without changing it
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  <subject> [--reference <reference>...] [--revision <revision>]
  [--kind <selection>] [--perspective <selection>] [--criteria <criteria>...]
  [--method <selection>] [--capability <selection>...] [--provider <selection>]
  [--result <destination-reference>]
---

# /xonovex-workflow:review — Review

## Arguments

- `subject` (required): Inline content or one opaque provider-native reference.
- `--reference` (repeatable): Supporting evidence or opaque references.
- `--revision` (optional): Exact native revision; required when provider context does
  not otherwise pin a referenced subject.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, conditionally required): Explicit review criteria. A
  selected method may supply them only when that contract is unambiguous.
- `--result` (optional): Explicit provider-native destination reference.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Review** operation with these arguments. The skill is the source of truth for the
procedure, output, and error handling; do not restate them.
