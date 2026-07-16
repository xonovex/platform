---
description: Declare, update, contain, recover, or close an urgent provider-native Incident with explicit authority and reporting applicability
allowed-tools:
  - Read
  - Bash
  - Glob
  - Grep
  - Task
  - AskUserQuestion
  - Skill
argument-hint: "<declare|update|contain|recover|close> [incident-reference] [--severity <level>] [--authorization <reference>] [--break-glass <reference>] [--provider <selection>]"
---

# /xonovex-workflow:incident-run — Run Incident Response

## Arguments

- Action (required): `declare`, `update`, `contain`, `recover`, or `close`.
- `incident-reference` (required except for declaration): Opaque provider-native Incident reference.
- `--severity` (optional): Current severity under the selected incident policy.
- `--authorization`, `--break-glass` (conditional): Privileged-response authority.
- `--provider` (optional): Incident/evidence provider selection.

## Delegation

Load `workflow-guide` and perform **incident-run**. Load `agent-governance-guide` for
privileged actions and emergency access, then soft-select incident, monitoring, security,
privacy, legal, resilience, and provider skills. Agents cannot fabricate reporting
applicability, accountable decisions, or notifications.
