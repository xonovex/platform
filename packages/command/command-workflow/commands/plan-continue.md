---
description: Resume one provider-native Planning result by reconstructing its handle and completing only the first actionable child
allowed-tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - AskUserQuestion
  - Skill
argument-hint: "[native-reference] [--provider-context <context>] [--revision <native-revision>]"
---

# /xonovex-workflow:plan-continue — Continue Planning Result

## Arguments

- `native-reference` (optional only when current provider context identifies one result): Opaque parent/child Planning reference.
- `--provider-context` (optional): Explicit provider selection/context.
- `--revision` (optional): Exact native revision; otherwise resolve current provider revision and verify freshness.

## Delegation

Load `plan-guide` and perform **plan-continue**. Reconstruct state from the selected
persistence after context loss, baseline the actual toolchain, load required skills,
complete one actionable Planning result, record validation/status, and stop.
