---
description: Revise an exact Solution Design revision while preserving feedback, evidence freshness, and supersession
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - Skill
argument-hint: "<native-reference> [--revision <native-revision>] [--feedback <reference>...] [--provider <selection>]"
---

# /xonovex-workflow:solution-design-revise — Revise Solution Design

## Arguments

- `native-reference` (required): Opaque Solution Design reference.
- `--revision` (required when not provider-implied): Exact subject revision.
- `--feedback` (repeatable): Opaque critique, evidence, feedback, or decision reference.
- `--provider` (optional): Explicit result-provider selection.

## Delegation

Load `plan-guide` and perform **solution-design-revise**. Preserve the original and
supersession links when supported, and re-evaluate affected evidence.
