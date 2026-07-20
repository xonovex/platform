---
description: Expand an approved Planning revision into provider-native child Planning results and execution groups
allowed-tools:
  - Write
  - Read
  - Bash
  - Glob
  - Grep
  - TaskCreate
  - TaskUpdate
  - Skill
argument-hint: "<native-reference> [--revision <native-revision>] [--provider <selection>] [--by-phase] [--dry-run]"
---

# /xonovex-workflow:plan-subplans-create — Create Child Planning Results

## Arguments

- `native-reference` (required): Opaque approved parent Planning reference.
- `--revision` (required when not provider-implied): Exact approved revision.
- `--provider` (optional): Child result provider.
- `--by-phase` (optional): Group by explicit lifecycle/profile phases instead of logical components.
- `--dry-run` (optional): Preview native child publications and relationships.

## Delegation

Load `plan-guide` and perform **plan-subplans-create**. Publish child Planning results and
execution groups, then stop. Files, line numbers, Git configuration, and worktrees are
optional persistence/workspace behavior.
