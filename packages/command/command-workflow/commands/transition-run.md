---
description: Plan, execute, verify, or roll back an exact-context Transition with protected privileged actions
allowed-tools:
  - Read
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "<plan|execute|verify|rollback> --source <native-reference> --target <native-reference> [--authorization <reference>] [--plan <reference>] [--provider <selection>]"
---

# /xonovex-workflow:transition-run — Run a Transition Operation

## Arguments

- Mode (required): `plan`, `execute`, `verify`, or `rollback`.
- `--source`, `--target` (required): Exact provider-native transition contexts.
- `--authorization` (required for target-changing modes): Fresh exact-scope authorization.
- `--plan`, `--provider` (optional): Prior Transition plan and native adapter selection.

## Delegation

Load `workflow-guide` and perform **transition-run** with the selected provider skill. If
the caller supplied a composition, load `agent-governance-guide` and apply only its
selected controls and evidence behavior. The command does not derive controls from the
target, executor, or host.
