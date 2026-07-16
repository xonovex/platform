---
description: Record an authorized decision against an exact native revision without fabricating human or qualified authority
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

# /xonovex-workflow:decision-accept — Accept Decision

## Arguments

- `native-reference` (required): Opaque proposed Decision reference.
- `--revision` (required when not provider-implied): Exact revision under decision.
- `--authority-reference` (optional only when provider context proves authority): Actor/role authority reference.
- `--provider` (optional): Decision provider.

## Delegation

Load `plan-guide` and perform **decision-accept**; load `workflow-guide` for native result
contracts. Verify actor, qualification, scope, and freshness, then bind the decision to the
exact revision. A model may prepare the brief but cannot supply the authority action.
