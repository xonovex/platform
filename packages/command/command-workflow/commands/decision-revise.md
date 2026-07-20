---
description: Revise an exact Decision result without rewriting its evidence, authority, or supersession history
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "<native-reference> [--revision <native-revision>] [--feedback <reference>...] [--provider <selection>]"
---

# /xonovex-workflow:decision-revise — Revise Decision

## Arguments

- `native-reference` (required): Opaque Decision reference.
- `--revision` (required when not provider-implied): Exact Decision revision.
- `--feedback` (repeatable): Opaque feedback, critique, policy, or evidence reference.
- `--provider` (optional): Explicit result provider.

## Delegation

Load `plan-guide` and perform **decision-revise**. Preserve prior authority and evidence,
and reopen authority when material meaning changed.
