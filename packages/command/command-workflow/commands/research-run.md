---
description: Research a bounded question with evidence provenance, uncertainty, and a provider-native result
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - Task
  - WebSearch
  - WebFetch
  - AskUserQuestion
  - Skill
argument-hint: "<question-or-native-reference> [--scope <scope>] [--profile <reference>] [--provider <selection>] [--method <selection>]"
---

# /xonovex-workflow:research-run — Run Research

## Arguments

- `question-or-native-reference` (required): Research question or opaque native input reference.
- `--scope` (optional): Explicit inclusion, exclusion, retrieval-window, or stop boundaries.
- `--profile`, `--provider`, `--method` (optional): Explicit selections.

## Delegation

Load `plan-guide` (plugin `xonovex-skill-plan`) and perform **research-run**. Keep
evidence/provenance distinct from synthesis, expose uncertainty, and bound adaptive
exploration.
