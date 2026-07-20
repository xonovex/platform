---
description: Preflight an exact accepted revision and protected target without performing Integration
allowed-tools:
  - Read
  - Bash
  - Glob
  - Grep
  - Skill
argument-hint: "<accepted-reference> [--revision <native-revision>] --target <native-target> --authorization <reference> [--profile <reference>] [--provider <selection>]"
---

# /xonovex-workflow:integration-validate — Preflight Integration

## Arguments

- `accepted-reference` (required): Opaque accepted Deliverable Publication reference.
- `--revision` (optional): Exact accepted source revision.
- `--target` (required): Protected target and current provider-native revision.
- `--authorization` (required): Exact Acceptance or other authorization reference.
- `--profile`, `--provider` (optional): Workflow profile and native enforcement/provider selection.

## Delegation

Perform **integration-validate** with the selected target-provider skill. If a composition
is supplied, explain its selected controls separately. This operation is read-only and
cannot mutate or reserve the target.
