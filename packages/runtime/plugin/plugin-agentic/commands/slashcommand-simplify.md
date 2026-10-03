---
description: >-
  Automatically simplify slash command files by reducing verbosity while maintaining functionality
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - TodoWrite
  - Skill
argument-hint: "[command-file] [--dry-run] [--target-reduction <percent>]"
---

# /xonovex-agentic:slashcommand-simplify - Simplify Slash Command Documentation

Shorten a slash command while preserving its arguments and behavior.

## Arguments

- `command-file` (required): Path to slash command file to simplify
- `--dry-run` (optional): Preview changes without writing file
- `--target-reduction` (optional): Target reduction percentage (default: 50, range: 30-70)

## Delegation

Load the `command-guide` skill (plugin `xonovex-agentic`) and perform its **simplify** operation with these arguments. The skill is the source of truth for the procedure, output format, and gotchas. Do not restate them.
