---
description: Create a new result from an inline or provider-referenced subject without changing the source
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
  [--result <destination-reference>] [--dry-run]
---

# /xonovex-workflow:create — Create

## Arguments

- `subject` (required): Inline source content or one opaque provider-native reference.
- `--reference` (repeatable): Supporting inline content or opaque references.
- `--revision` (optional): Exact native revision of a referenced subject.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, optional): Explicit properties the new result must satisfy.
- `--result` (optional): Explicit provider-native destination reference.
- `--dry-run` (optional): Preview any persistence or other side effect.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Create** operation with these arguments. The skill is the source of truth for the
procedure, output, and error handling; do not restate them.
