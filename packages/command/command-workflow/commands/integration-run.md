---
description: Execute an authorized exact-revision Integration through a protected external enforcement point
allowed-tools:
  - Read
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "<accepted-reference> [--revision <native-revision>] --target <native-target> --authorization <reference> [--preflight <reference>] [--provider <selection>] [--idempotency-key <key>]"
---

# /xonovex-workflow:integration-run — Execute Integration

## Arguments

- `accepted-reference` (required): Opaque accepted Deliverable Publication reference.
- `--revision` (optional): Exact accepted source revision.
- `--target` (required): Protected target and current native revision.
- `--authorization` (required): Fresh exact-scope authorization reference.
- `--preflight`, `--provider`, `--idempotency-key` (optional): Preflight, native adapter, and safe-retry key.

## Delegation

Perform **integration-run** using the selected provider's native target-changing
capability and authorization mechanism.
