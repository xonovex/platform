---
description: Inspect or lifecycle-manage governance modules through native adapters with preview, authorization, verification, rollback, and drift checks
allowed-tools:
  - Bash
  - Read
  - Glob
  - Grep
  - Skill
argument-hint: "<list|inspect|enable|disable|upgrade|remove> [module] [--apply] [--profile <reference>]"
---

# /xonovex-workflow:workflow-modules — Manage Governance Modules

## Arguments

- `list|inspect|enable|disable|upgrade|remove` (required): Module lifecycle operation.
- `module` (optional): Module identity or native reference; required except for `list`.
- `--apply` (optional): Apply an already previewed and authorized change; without this flag changing operations stop after preview.
- `--profile <reference>` (optional): Profile whose requirements must remain satisfied.

## Delegation

Load the `agent-governance-guide` skill (plugin
`xonovex-skill-agent-governance`) and perform its **manage modules** operation with these
arguments. The skill is the source of truth for trust, dependencies, conflicts,
permissions, preview, authorization, native adapters, verification, rollback, evidence,
and gotchas. Fail visibly when no native adapter can perform or verify the requested
operation; do not edit a universal registry.
