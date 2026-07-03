---
description: "Pre-plan/draft: settle decisions one at a time — walk known open ones or discover them by questioning you"
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Task
  - AskUserQuestion
  - Write
  - Skill
argument-hint: "[topic-or-plan-file] [--save-to <file>]"
---

# /xonovex-workflow:plan-decide — Settle Decisions One at a Time

> Lifecycle: research → **decide** → create → revise ⇄ critique → subplans-create → continue → update → validate

## Arguments

- `topic-or-plan-file` (optional): A feature idea/topic, or a research/plan file. Auto-detects queued open decisions first (research findings in the current conversation, then the most recent plan); with none queued, discovers decisions instead.
- `--save-to <file>` (optional): Save the consolidated agreed-direction summary to a file

## Modes

- **Walk** (open decisions are queued): present each known decision as a full brief — files involved, explanation, options with pros and cons, recommendation — one per message, recording your call before the next.
- **Discover** (nothing queued): surface unknown decisions by walking the design tree one question at a time, exploring the codebase to self-answer what it can, until shared understanding is reached.

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-decide** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
