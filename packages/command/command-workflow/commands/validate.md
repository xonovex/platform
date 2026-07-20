---
description: Validate an exact subject against explicit criteria and return evidence without changing the subject
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  <subject> --criteria <criteria>... [--reference <reference>...]
  [--revision <revision>] [--kind <selection>] [--perspective <selection>]
  [--method <selection>] [--capability <selection>...]
  [--provider <selection>] [--result <destination-reference>]
---

# /xonovex-workflow:validate — Validate

## Arguments

- `subject` (required): Inline content or one opaque provider-native reference.
- `--criteria` (repeatable, required): Inline criterion or opaque criteria reference.
- `--reference` (repeatable): Supporting evidence or opaque references.
- `--revision` (optional): Exact native revision; required when provider context does
  not otherwise pin a referenced subject.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--result` (optional): Explicit provider-native evidence destination.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Validate** operation with these arguments. The skill is the source of truth for the
procedure, output, and error handling; do not restate them.
