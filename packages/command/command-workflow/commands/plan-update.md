---
description: Publish current implementation status and exact-revision validation evidence as a new Planning revision
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - TaskUpdate
  - Skill
argument-hint: "<native-reference> [--revision <native-revision>] [--provider <selection>] [--evidence <reference>...]"
---

# /xonovex-workflow:plan-update — Update Planning Result

## Arguments

- `native-reference` (required): Opaque Planning reference.
- `--revision` (required when not provider-implied): Exact revision to update.
- `--provider` (optional): Result provider.
- `--evidence` (repeatable): Native Development, validation, review, or policy evidence reference.

## Delegation

Load `plan-guide` and perform **plan-update**. Verify tasks and criteria against available
evidence, preserve unavailable validation categories, and do not infer completion from
conversation memory.
