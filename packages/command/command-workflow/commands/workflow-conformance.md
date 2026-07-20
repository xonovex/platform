---
description: Validate either a lifecycle workflow contract or an executable composition
allowed-tools:
  - Bash
  - Read
  - Glob
  - Grep
  - Skill
argument-hint: "<subject> [--kind <lifecycle|composition>] [--registry <registry.json>]"
---

# /xonovex-workflow:workflow-conformance — Validate a Workflow Contract

## Arguments

- `subject` (required): Lifecycle result, provider handle, invocation, or native reference.
- `--kind` (optional): `lifecycle` or `composition`; infer only when the subject is unambiguous.
- `--registry` (required for a composition): Trusted plugin registry used to resolve the invocation.

## Delegation

- For `lifecycle`, load `workflow-guide` and perform its conformance operation.
- For `composition`, load `agent-governance-guide`, validate the invocation against the
  registry, and explain the resolved plugins, capability gate, and enforcement points.

Do not assemble profiles across these contract types. Lifecycle semantics and executable
composition are independent and may be used separately.
