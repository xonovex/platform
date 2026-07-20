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

Perform **transition-run** with the selected provider skill. If the caller supplied a
composition, pass it unchanged to the shared workflow runtime. The command does not
derive controls from the target, executor, or host.
