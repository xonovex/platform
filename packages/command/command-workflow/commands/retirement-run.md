---
description: Plan, execute, verify, or roll back authorized Retirement of exact models, data, credentials, features, APIs, infrastructure, dependencies, or provider configuration
allowed-tools:
  - Read
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "<plan|execute|verify|rollback> <subject-reference...> [--revision <native-revision>] --authorization <reference> [--provider <selection>]"
---

# /xonovex-workflow:retirement-run — Retire Lifecycle Resources

## Arguments

- Mode (required): `plan`, `execute`, `verify`, or `rollback`.
- `subject-reference...` (required): Exact resources selected for retirement.
- `--revision` (repeatable): Current native revisions or resource versions.
- `--authorization` (required for target-changing modes): Fresh exact-scope authorization.
- `--provider` (optional): Protected resource provider selection.

## Delegation

Load `workflow-guide` and perform **retirement-run**. Load `agent-governance-guide` plus
the selected resource/provider skills. Use explicit protected retirement, revocation, and
deletion capabilities; missing local state or a successful request is not verification.
