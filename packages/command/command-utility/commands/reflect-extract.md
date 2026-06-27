---
description: Analyze the session for development mistakes and lessons learned
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Skill
argument-hint: "[category] [--out-dir <dir>]"
---

# /xonovex-utility:reflect-extract — Extract Development Lessons

## Arguments

- `category` (optional): Focus on a specific mistake category (e.g., `tool-usage`, `dependencies`, `validation`).
- `--out-dir` (optional): The directory to save insight files in. Defaults to `reflections/`.

## Delegation

Load the `reflect-guide` skill (plugin `xonovex-skill-reflect`) and perform its
**extract** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
