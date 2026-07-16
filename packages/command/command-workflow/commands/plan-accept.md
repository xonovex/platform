---
description: Record an authorized approval decision against an exact provider-native Planning revision
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

# /xonovex-workflow:plan-accept — Accept Planning Result

## Arguments

- `native-reference` (required): Opaque Planning reference.
- `--revision` (required when not provider-implied): Exact revision under approval.
- `--authority-reference` (optional only when provider context proves authority): Actor/role reference.
- `--provider` (optional): Status/decision provider.

## Delegation

Load `plan-guide` and perform **plan-accept**; load `workflow-guide` for provider-native
result contracts. Verify authority, bind the decision to the exact revision, publish it,
and stop. A model may summarize but cannot fabricate approval.
