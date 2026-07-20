---
description: Discover a problem or opportunity without forcing one method, provider, executor, or artifact shape
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

# /xonovex-workflow:discovery-run — Run Discovery

## Arguments

- `subject-or-native-reference` (required): Problem, opportunity, scope, or opaque native input reference.
- `--profile`, `--provider`, `--method` (optional): Explicit independent axis selections; otherwise use profile/environment resolution.
- `--executor` (optional): Explicit `deterministic`, `model`, `agent`, `human`, or `external` workflow selection; it does not imply controls or maturity.

## Delegation

Load `plan-guide` (plugin `xonovex-skill-plan`) and perform **discovery-run**. Persist the
Discovery result as requested; do not assume files, Git, user stories, Gherkin, or one
agent session.
