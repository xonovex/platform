---
description: Research a bounded question with evidence provenance, uncertainty, executor limits, and a provider-native result
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
argument-hint: "<question-or-native-reference> [--scope <scope>] [--profile <reference>] [--provider <selection>] [--method <selection>] [--executor <class>]"
---

# /xonovex-workflow:research-run — Run Research

## Arguments

- `question-or-native-reference` (required): Research question or opaque native input reference.
- `--scope` (optional): Explicit inclusion, exclusion, retrieval-window, or stop boundaries.
- `--profile`, `--provider`, `--method`, `--executor` (optional): Independent axis selections.

## Delegation

Load `plan-guide` (plugin `xonovex-skill-plan`) and perform **research-run**. Load
`workflow-guide` (plugin `xonovex-skill-workflow`) for result and provider contracts.
Keep evidence/provenance distinct from synthesis, expose uncertainty, bound adaptive
exploration, and publish a canonical Research result through the selected provider.
