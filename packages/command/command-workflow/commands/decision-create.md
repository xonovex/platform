---
description: Prepare and record one evidence-grounded decision while keeping evidence, recommendation, and authority separate
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - Task
  - AskUserQuestion
  - Skill
argument-hint: "<question-or-native-reference> [--profile <reference>] [--provider <selection>] [--method <selection>] [--authority-reference <reference>]"
---

# /xonovex-workflow:decision-create — Create Decision

## Arguments

- `question-or-native-reference` (required): One decision question or opaque lifecycle input reference.
- `--profile`, `--provider`, `--method` (optional): Independent selections.
- `--authority-reference` (optional only when provider context proves it): Intended actor/role authority.

## Delegation

Load `plan-guide` and perform **decision-create**. Present one evidence-grounded decision
brief at a time. A model or agent may synthesize options and recommend, but only the
required human or qualified actor can supply authority.
