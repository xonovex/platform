---
description: Create an optional provider-native Experience Design result with independently selected methods and providers
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - Task
  - AskUserQuestion
  - Skill
argument-hint: "<subject-or-native-reference> [--profile <reference>] [--provider <selection>] [--method <selection>]"
---

# /xonovex-workflow:experience-design-create — Create Experience Design

## Arguments

- `subject-or-native-reference` (required): Subject or opaque Discovery/Research/Formulation/Decision reference.
- `--profile`, `--provider`, `--method` (optional): Explicit selections.

## Delegation

Load `plan-guide` (plugin `xonovex-skill-plan`) and perform
**experience-design-create**. Use compatible installed research, interaction, content,
prototyping, and accessibility capabilities as selected; do not make Experience Design
mandatory.
