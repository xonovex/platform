---
description: Independently stress-test an exact Planning revision and publish separate findings
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - Task
  - Skill
argument-hint: "<native-reference> [--revision <native-revision>] [--provider <selection>] [--mode red-team|pre-mortem|falsify|steelman|all]"
---

# /xonovex-workflow:plan-critique — Critique Planning Result

## Arguments

- `native-reference` (required): Opaque Planning reference.
- `--revision` (required when not provider-implied): Exact revision to critique.
- `--provider` (optional): Provider for the critique result.
- `--mode` (optional): Adversarial lens; defaults to red-team plus pre-mortem.

## Delegation

Load `plan-guide` and perform **plan-critique** in fresh independent context; load
`workflow-guide` for exact-revision and native publication contracts. Publish separate
findings and do not revise, approve, or implement the Planning result.
