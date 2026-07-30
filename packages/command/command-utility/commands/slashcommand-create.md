---
description: Create a thin slash command that delegates its reusable procedure to one owner skill
allowed-tools:
  - Read
  - Write
  - Edit
  - Glob
  - TodoWrite
  - AskUserQuestion
  - Skill
argument-hint: >-
  [description] [--name <name>] [--owner-skill <skill-name>]
  [--owner-plugin <plugin-name>] [--operation <name>]
  [--interactive] [--dry-run] [--force]
---

# /xonovex-utility:slashcommand-create - Create Slash Command

## Arguments

- `description` (required): What the reusable command should accomplish.
- `--name` (optional): Kebab-case command name.
- `--owner-skill`, `--owner-plugin`, `--operation` (optional): Exact guide,
  distribution plugin, and operation that own the procedure.
- `--interactive` (optional): Ask about arguments and delegation choices.
- `--dry-run` (optional): Preview command, skill, and manifest artifacts.
- `--force` (optional): Replace the exact existing command only after preview.

## Delegation

Load the `command-guide` skill (plugin `xonovex-skill-command`) and perform its
**create** operation with these arguments. The skill is the source of truth for the
procedure, output, safety, and harness-specific format.
