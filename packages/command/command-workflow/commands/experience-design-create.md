---
description: Create an optional provider-native Experience Design result with independent methods and policy intents
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

# /xonovex-workflow:experience-design-create — Create Experience Design

## Arguments

- `subject-or-native-reference` (required): Subject or opaque Discovery/Research/Formulation/Decision reference.
- `--profile`, `--provider`, `--method`, `--executor` (optional): Independent selections.

## Delegation

Load `plan-guide` (plugin `xonovex-skill-plan`) and perform
**experience-design-create**. Load `workflow-guide` (plugin
`xonovex-skill-workflow`) for result and provider contracts. Use compatible installed
research, interaction, content, prototyping, and accessibility capabilities as selected;
publish a proposed Experience Design result without making it mandatory for every profile.
