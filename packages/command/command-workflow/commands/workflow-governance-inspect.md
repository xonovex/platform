---
description: Inspect effective agent governance policies, modules, enforcement guarantees, authority, evidence, exceptions, and gaps
allowed-tools:
  - Bash
  - Read
  - Glob
  - Grep
  - Skill
argument-hint: "[scope] [--profile <reference>] [--include-advisory]"
---

# /xonovex-workflow:workflow-governance-inspect — Inspect Effective Governance

## Arguments

- `scope` (optional): Repository, project, user, session, external system, or native configuration reference; defaults to the current environment.
- `--profile <reference>` (optional): Explicit governance profile reference, resolved to a shipped reference profile in `agent-governance-guide/assets/profiles/`.
- `--include-advisory` (optional): Include advisory modules alongside evidence-producing and enforcing modules.

## Delegation

Load the `agent-governance-guide` skill (plugin
`xonovex-skill-agent-governance`) and perform its **inspect** operation with these
arguments. The skill is the source of truth for policy, module, authority, evidence,
exception, enforcement, output, and gotchas — do not restate them.
