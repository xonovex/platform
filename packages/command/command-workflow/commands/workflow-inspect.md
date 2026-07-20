---
description: Explain a workflow composition without executing it
allowed-tools:
  - Bash
  - Read
  - Skill
argument-hint: "<registry.json> [invocation.json]"
---

# /xonovex-workflow:workflow-inspect — Explain a Composition

## Arguments

- `registry.json` (required): Trusted executor, control, and evidence plugin registry.
- `invocation.json` (optional): Invocation file; when omitted, read the invocation from standard input.

## Delegation

Inspect the invocation with the shared workflow runtime. Report the selected executor,
controls and their observe/enforce modes, evidence sinks and their failure modes,
available capabilities, missing required capabilities, and exact enforcement points. Do
not execute plugins, infer a maturity level, or add defaults.
