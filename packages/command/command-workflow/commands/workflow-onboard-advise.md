---
description: Discover an environment and preview compatible governance modules, permissions, data flows, verification, and rollback without applying changes
allowed-tools:
  - Bash
  - Read
  - Glob
  - Grep
  - Skill
argument-hint: "[scope] [--profile <reference>] [--module <name>]"
---

# /xonovex-workflow:workflow-onboard-advise — Advise Governance Onboarding

## Arguments

- `scope` (optional): Environment, repository, organization, harness, CI, or provider scope to discover; defaults to the current environment.
- `--profile <reference>` (optional): Desired governance profile.
- `--module <name>` (optional): Candidate module to assess and preview.

## Delegation

Load the `agent-governance-guide` skill (plugin
`xonovex-skill-agent-governance`) and perform its **advise onboarding** operation with
these arguments. Stop after discovery, assessment, recommendation, and exact preview;
this command never authorizes or applies configuration. The skill is the source of truth
for capabilities, permissions, data flows, verification, rollback, output, and gotchas.
