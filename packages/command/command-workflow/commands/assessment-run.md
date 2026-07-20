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
argument-hint: "<subject-reference> [--revision <native-revision>] --criteria <reference> [--profile <reference>] [--provider <selection>] [--executor <class>]"
---

# /xonovex-workflow:assessment-run — Run an Assessment

## Arguments

- `subject-reference` (required): Opaque reference to any canonical workflow result.
- `--revision` (optional): Exact native subject revision.
- `--criteria` (required): Pinned framework, policy, risk, security, accessibility, AI, supply-chain, privacy, legal, or other criterion set.
- `--profile`, `--provider` (optional): Independent selections.
- `--executor` (optional): Explicit `deterministic`, `model`, `agent`, `human`, or `external` workflow selection; it does not imply controls or maturity.

## Delegation

Load `workflow-guide` and perform **assessment-run**. Soft-select installed criterion,
testing, scanner, policy, and provider skills matching the requested assessment. Preserve
evidence origin and qualification; fail visibly when mandatory expertise or enforcement
is unavailable.
