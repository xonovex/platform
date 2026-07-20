---
description: Execute, verify, or recover an exact-revision Release through controlled automation and a protected environment
allowed-tools:
  - Read
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "<execute|verify|rollback|recover> <integrated-reference> [--revision <native-revision>] --target <environment> --authorization <reference> [--provider <selection>]"
---

# /xonovex-workflow:release-run — Run a Controlled Release

## Arguments

- Mode (required): `execute`, `verify`, `rollback`, or `recover`.
- `integrated-reference` (required): Exact Integration result or immutable artifact reference.
- `--revision`, `--target` (required): Immutable source revision/digest and protected release target.
- `--authorization` (required): Fresh release authorization.
- `--provider` (optional): Controlled automation/protected-environment adapter.

## Delegation

Load `workflow-guide` and perform **release-run** with the selected automation/provider
skill. If the caller supplied a composition, load `agent-governance-guide` and apply only
its selected control modes and evidence behavior. Do not infer approval or maturity from
the host or executor.
