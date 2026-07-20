---
description: Assess any exact workflow result against pinned criteria with explicit applicability, evaluator origin, evidence, and freshness
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - Task
  - AskUserQuestion
  - Skill
argument-hint: "<subject-reference> [--revision <native-revision>] --criteria <reference> [--profile <reference>] [--provider <selection>]"
---

# /xonovex-workflow:assessment-run — Run an Assessment

## Arguments

- `subject-reference` (required): Opaque reference to the subject being assessed.
- `--revision` (optional): Exact native subject revision.
- `--criteria` (required): Pinned framework, policy, risk, security, accessibility, AI, supply-chain, privacy, legal, or other criterion set.
- `--profile`, `--provider` (optional): Assessment context and result provider.

## Delegation

Perform **assessment-run** with installed criterion, testing, scanner, policy, and provider
skills selected for the request. Preserve evidence origin and qualification; fail visibly
when required expertise or enforcement is unavailable.
