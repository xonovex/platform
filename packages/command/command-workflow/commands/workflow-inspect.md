---
description: Inspect effective workflow capabilities, native result handoffs, profile topology, evidence, and completion gaps
allowed-tools:
  - Bash
  - Read
  - Glob
  - Grep
  - Skill
argument-hint: "[result-or-profile] [--profile <reference>] [--provider <name>]"
---

# /xonovex-workflow:workflow-inspect — Inspect Composable Workflow State

## Arguments

- `result-or-profile` (optional): Opaque native result reference, local result, or profile to inspect; defaults to the current workflow context.
- `--profile <reference>` (optional): Explicit profile reference or identity, resolved to a shipped reference profile in `workflow-guide/assets/profiles/`; an integrated reference follows its governance-facet cross-reference into `agent-governance-guide/assets/profiles/`.
- `--provider <name>` (optional): Explicit provider capability for resolving the input; fail visibly if unavailable.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**inspect** operation with these arguments. The skill is the source of truth for result
semantics, provider-native handoffs, profile topology, completion, output, and gotchas —
do not restate them.
