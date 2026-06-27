---
description: >-
  Interrogate the user to surface unknown decisions before a plan exists — walk
  the design tree one question at a time, explore the codebase to self-answer,
  recommend an answer per question, until shared understanding is reached
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Task
  - AskUserQuestion
  - Write
  - Skill
argument-hint: "[topic-or-file] [--save-to <file>]"
---

# /xonovex-workflow:plan-interrogate — Interrogate to Surface Unknown Decisions

## Arguments

- `topic-or-file` (optional): The feature idea, design direction, or a research/plan file to interrogate. Auto-detects the current conversation's direction first, then the most recent `plans/*.md`.
- `--save-to <file>` (optional): Save the consolidated shared-understanding summary to a file.

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-interrogate** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
