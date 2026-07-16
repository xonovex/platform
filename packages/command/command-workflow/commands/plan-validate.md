---
description: Validate an exact Planning revision and its Definition of Done without mutating it
allowed-tools:
  - Bash
  - Read
  - Grep
  - Glob
  - Task
  - TaskCreate
  - TaskUpdate
  - Write
  - Skill
argument-hint: "<native-reference> [--revision <native-revision>] [--provider <selection>] [--publish-evidence]"
---

# /xonovex-workflow:plan-validate — Validate Planning Result

## Arguments

- `native-reference` (required): Opaque Planning reference.
- `--revision` (required when not provider-implied): Exact revision to validate.
- `--provider` (optional): Provider used to resolve the result and optionally publish validation evidence.
- `--publish-evidence` (optional): Publish the read-only validation report through the selected provider.

## Delegation

Load `plan-guide` and perform **plan-validate**; load `workflow-guide` for exact-revision
and evidence contracts. Check every criterion and Definition of Done item independently,
report stale/missing evidence, and do not revise status or claim Acceptance.
