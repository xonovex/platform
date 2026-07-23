---
description: Distill a fat slash command into a thin delegator with one owner and explicit supporting requirements
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - TodoWrite
  - AskUserQuestion
  - Skill
argument-hint: >-
  [command-file] [--owner-skill <skill-name>] [--owner-plugin <plugin-name>]
  [--operation <name>] [--requires <semantic-capability>...] [--dry-run]
---

# /xonovex-utility:slashcommand-distill — Distill a Command

## Arguments

- `command-file` (required): Exact slash command to distill.
- `--owner-skill`, `--owner-plugin`, `--operation` (optional): Exact guide,
  distribution plugin, and operation that own the procedure.
- `--requires` (repeatable, optional): Soft supporting semantic requirement.
- `--dry-run` (optional): Preview command, skill, and manifest changes.

## Delegation

Load the `command-guide` skill (plugin `xonovex-skill-command`) and perform its
**distill** operation with these arguments. The skill is the source of truth for the
procedure, output, safety, and ownership rules.
