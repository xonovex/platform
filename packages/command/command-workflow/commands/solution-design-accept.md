---
description: Record an authorized accept, reject, or conditional decision for an exact Solution Design revision
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "<native-reference> [--revision <native-revision>] [--authority-reference <reference>] [--provider <selection>]"
---

# /xonovex-workflow:solution-design-accept — Accept Solution Design

## Arguments

- `native-reference` (required): Opaque Solution Design reference.
- `--revision` (required when not provider-implied): Exact revision under decision.
- `--authority-reference` (optional only when provider context proves authority): Actor/role authority reference.
- `--provider` (optional): Decision/result provider.

## Delegation

Load `plan-guide` and perform **solution-design-accept**. Verify any required qualification
and revision binding; a model cannot grant or impersonate authority.
