---
description: Execute one bounded subject and report its observable result without implicit publication or workspace changes
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
  <subject> [--reference <reference>...] [--revision <revision>]
  [--kind <selection>] [--perspective <selection>] [--criteria <criteria>...]
  [--method <selection>] [--capability <selection>...] [--provider <selection>]
  [--result <destination-reference>] [--dry-run]
---

# /xonovex-workflow:execute — Execute

## Arguments

- `subject` (required): Inline instructions or one opaque provider-native reference.
- `--reference` (repeatable): Supporting inputs or opaque references.
- `--revision` (optional): Exact native revision of a referenced subject.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, optional): Explicit completion and verification criteria.
- `--result` (optional): Explicit provider-native result destination.
- `--dry-run` (optional): Preview actions and side effects without applying them.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Execute** operation with these arguments. The skill is the source of truth for the
procedure, output, and error handling; do not restate them.
