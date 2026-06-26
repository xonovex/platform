---
description: Create a new slash command from a completed task or goal
allowed-tools:
  - Read
  - Write
  - Glob
  - TodoWrite
  - AskUserQuestion
  - Skill
argument-hint: "[description] [--name <name>] [--interactive]"
---

# /xonovex-utility:slashcommand-create — Create Slash Command from Task

## Arguments

- `description` (required): Brief description of what the task accomplished
- `--name` (optional): Command name (auto-generated from description if not provided)
- `--interactive` (optional): Ask clarifying questions about arguments, validation, output, and error handling

## Delegation

Load the `command-guide` skill (plugin `xonovex-skill-command`) and perform its
**create** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
