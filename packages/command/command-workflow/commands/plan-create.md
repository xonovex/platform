---
description: Publish a high-level provider-native Planning result from resolved lifecycle references for review
allowed-tools:
  - Write
  - Read
  - Bash
  - Glob
  - Grep
  - Task
  - TaskCreate
  - TaskUpdate
  - AskUserQuestion
  - Skill
argument-hint: "[subject-or-source-reference] [--source <reference>...] [--profile <reference>] [--provider <selection>] [--method <selection>] [--dry-run]"
---

# /xonovex-workflow:plan-create — Create Planning Result

## Arguments

- `subject-or-source-reference` (optional): Subject or opaque Research/Decision/Formulation/Design reference; defaults to resolved current lifecycle context.
- `--source` (repeatable): Additional opaque native source reference.
- `--profile`, `--provider`, `--method` (optional): Independent workflow-axis selections.
- `--dry-run` (optional): Preview the provider-native publication without applying it.

## Delegation

Load `plan-guide` and perform **plan-create**. Publish one high-level Planning result
through the requested persistence and stop before child plans or implementation.
