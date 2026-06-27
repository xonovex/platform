---
description: >-
  Distill a fat slash command into a thin delegator that loads its guideline
  skill at run time — move the procedure into a skill reference and keep only
  the argument contract and a delegation block
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - TodoWrite
  - AskUserQuestion
  - Skill
argument-hint: "[command-file] [--skill <plugin>] [--operation <name>] [--dry-run]"
---

# /xonovex-utility:slashcommand-distill — Distill a Command into a Skill Delegator

## Arguments

- `command-file` (required): Path to the slash command file to distill.
- `--skill <plugin>` (optional): Guideline-skill plugin that should own the procedure (auto-detected from the command's domain if omitted).
- `--operation <name>` (optional): Operation/reference name within the skill (defaults to the command's verb).
- `--dry-run` (optional): Preview the thin command, the skill reference, and the manifest changes without writing.

## Delegation

Load the `command-guide` skill (plugin `xonovex-skill-command`) and perform its
**distill** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
