---
description: Create an optional provider-native Solution Design result with explicit boundaries, qualities, trade-offs, and risks
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - Task
  - AskUserQuestion
  - Skill
argument-hint: "<subject-or-native-reference> [--profile <reference>] [--provider <selection>] [--method <selection>] [--executor <class>]"
---

# /xonovex-workflow:solution-design-create — Create Solution Design

## Arguments

- `subject-or-native-reference` (required): Subject or opaque lifecycle input reference.
- `--profile`, `--provider`, `--method` (optional): Independent selections.
- `--executor` (optional): Explicit `deterministic`, `model`, `agent`, `human`, or `external` workflow selection; it does not imply controls or maturity.

## Delegation

Load `plan-guide` and perform **solution-design-create**. Select compatible architecture,
security, privacy, data, platform, and technology skills by context; do not make Solution
Design a prerequisite.
