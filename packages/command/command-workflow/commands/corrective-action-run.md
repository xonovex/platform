---
description: Plan, execute, verify, or close a provider-native Corrective Action with effectiveness and learning evidence
allowed-tools:
  - Read
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "<plan|execute|verify|close> <source-reference> [--revision <native-revision>] [--action <reference>] [--provider <selection>]"
---

# /xonovex-workflow:corrective-action-run — Run Corrective Action

## Arguments

- Mode (required): `plan`, `execute`, `verify`, or `close`.
- `source-reference` (required): Exact Incident, finding, Observation, or assurance result.
- `--revision` (optional): Exact source revision.
- `--action` (optional): Existing Corrective Action reference for revise/verify/close.
- `--provider` (optional): Result and evidence provider selection.

## Delegation

Perform **corrective-action-run** with selected root-cause, verification, provider, and
domain skills. Resolve any caller-selected executable composition independently; the
command does not add a control or maturity requirement. Closure requires the evidence
selected by the caller.
