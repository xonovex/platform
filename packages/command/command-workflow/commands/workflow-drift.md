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
- `--profile <reference>` (optional): Intended profile state, resolved to a shipped reference profile in the owning plane's `assets/profiles/` library.
- `--baseline <reference>` (optional): Provider-native configuration or evidence reference to compare.

## Delegation

Load the `agent-governance-guide` skill (plugin
`xonovex-skill-agent-governance`) and perform its **evaluate drift** operation with these
arguments. The skill is the source of truth for read-only discovery, classification,
conformance, evidence, output, remediation preview, and gotchas — do not restate them.
For an `AgentRun`, evaluate the runtime's correlated `oversight.control` signals and
compare applied/observed policy references. Treat `status.effectiveAutonomy` and
`status.containment` as the authoritative demotion/kill-switch result. Any approved
remediation run must enter through the authenticated `AgentTrigger` POST surface; do not
create a second trigger path.
