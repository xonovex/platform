---
description: Revise an exact Planning revision from explicit native feedback and publish a new revision
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

# /xonovex-workflow:plan-revise — Revise Planning Result

## Arguments

- `native-reference` (required): Opaque Planning reference.
- `--revision` (required when not provider-implied): Exact subject revision.
- `--feedback` (repeatable): Native feedback, critique, or annotation reference.
- `--provider` (optional): Explicit result provider.

## Delegation

Load `plan-guide` and perform **plan-revise**. Resolve every item, preserve history when
supported, and stop without implementing or transferring approval from the prior result.
