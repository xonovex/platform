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
- `--profile`, `--provider`, `--method`, `--executor` (optional): Independent selections.

## Delegation

Load `plan-guide` and perform **solution-design-create**; load `workflow-guide` for
result and provider contracts. Select compatible architecture, security, privacy, data,
platform, and technology skills by context, then publish a proposed Solution Design
without making architecture a prerequisite for every profile.
