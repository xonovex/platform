---
description: Simplify project instruction files (AGENTS.md) by reducing verbosity
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - TodoWrite
  - Skill
argument-hint: "[instruction-file] [--dry-run] [--target-reduction <percent>]"
---

# /xonovex-utility:instructions-simplify — Simplify project instruction files

## Arguments

- `instruction-file` (required): Path to AGENTS.md file
- `--dry-run` (optional): Preview without modifying
- `--target-reduction <percent>` (optional): Override default 45% (range: 30-60)

## Delegation

Load the `instruction-guide` skill (plugin `xonovex-skill-instruction`) and perform its
**simplify** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
