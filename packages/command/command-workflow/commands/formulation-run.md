---
description: Formulate candidate behavior, requirements, examples, and constraints with a neutral default method
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - Task
  - AskUserQuestion
  - Skill
argument-hint: "<subject-or-native-reference> [--profile <reference>] [--provider <selection>] [--method <selection>] [--executor <class>]"
---

# /xonovex-workflow:formulation-run — Run Formulation

## Arguments

- `subject-or-native-reference` (required): Discovery/Research input, subject, or opaque native reference.
- `--profile`, `--provider`, `--method` (optional): Independent axis selections; neutral is the default method.
- `--executor` (optional): Explicit `deterministic`, `model`, `agent`, `human`, or `external` workflow selection; it does not imply controls or maturity.

## Delegation

Load `plan-guide` (plugin `xonovex-skill-plan`) and perform **formulation-run**. User
stories, BDD, and example mapping are selectable installed methods, not hard dependencies.
