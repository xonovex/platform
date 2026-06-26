---
description: Sync AGENTS.md files with current directory structure and state
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - Bash
  - TodoWrite
  - Skill
argument-hint: "[agents-file | --all] [--dry-run] [--update-workflows]"
---

# /xonovex-utility:instructions-sync — Sync AGENTS.md with Current State

## Arguments

- `agents-file` (optional): Path to specific AGENTS.md file to update
- `--all` (optional): Update all AGENTS.md files in repository
- `--dry-run` (optional): Preview without modifying
- `--update-workflows` (optional): Refresh commands from package.json/config files

## Delegation

Load the `instruction-guide` skill (plugin `xonovex-skill-instruction`) and perform its
**sync** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
