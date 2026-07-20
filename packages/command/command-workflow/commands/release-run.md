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

Perform **release-run** with the selected automation/provider skill and its native
authorization mechanism.
