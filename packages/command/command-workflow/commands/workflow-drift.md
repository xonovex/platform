---
description: Compare intended and observed governance versions, capabilities, policy, configuration, exceptions, and evidence freshness
allowed-tools:
  - Bash
  - Read
  - Glob
  - Grep
  - Skill
argument-hint: "[scope] [--profile <reference>] [--baseline <reference>]"
---

# /xonovex-workflow:workflow-drift — Evaluate Governance Drift

## Arguments

- `scope` (optional): Native configuration, repository, harness, provider, or environment scope; defaults to the current environment.
- `--profile <reference>` (optional): Intended profile state.
- `--baseline <reference>` (optional): Provider-native configuration or evidence reference to compare.

## Delegation

Load the `agent-governance-guide` skill (plugin
`xonovex-skill-agent-governance`) and perform its **evaluate drift** operation with these
arguments. The skill is the source of truth for read-only discovery, classification,
conformance, evidence, output, remediation preview, and gotchas — do not restate them.
