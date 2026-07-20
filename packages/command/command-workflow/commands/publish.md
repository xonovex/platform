---
description: Publish an exact subject to an explicit provider destination and return its native locator
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
  <subject> --result <destination-reference> [--reference <reference>...]
  [--revision <revision>] [--kind <selection>] [--perspective <selection>]
  [--criteria <criteria>...] [--method <selection>]
  [--capability <selection>...] [--provider <selection>]
  [--confirm] [--dry-run]
---

# /xonovex-workflow:publish — Publish

## Arguments

- `subject` (required): Inline content or one opaque provider-native reference.
- `--result` (required): Explicit provider-native destination reference.
- `--reference` (repeatable): Supporting inputs or opaque references.
- `--revision` (optional): Exact native source revision; required when provider
  context does not otherwise pin it.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, optional): Preconditions the publication must satisfy.
- `--confirm` (optional): Explicitly authorize the described publication effect.
- `--dry-run` (optional): Resolve and preview the publication without applying it.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Publish** operation with these arguments. The skill is the source of truth for the
procedure, output, and error handling; do not restate them.
