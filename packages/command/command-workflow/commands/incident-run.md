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
argument-hint: "<declare|update|contain|recover|close> [incident-reference] [--severity <level>] [--authorization <reference>] [--emergency-exception <reference>] [--provider <selection>]"
---

# /xonovex-workflow:incident-run — Run Incident Response

## Arguments

- Action (required): `declare`, `update`, `contain`, `recover`, or `close`.
- `incident-reference` (required except for declaration): Opaque provider-native Incident reference.
- `--severity` (optional): Current severity under the selected incident policy.
- `--authorization`, `--emergency-exception` (conditional): Privileged-response authority.
- `--provider` (optional): Incident/evidence provider selection.

## Delegation

Perform **incident-run** with selected incident, monitoring, security, privacy, legal,
resilience, and provider skills. Agents cannot fabricate reporting applicability,
accountable decisions, or notifications.
