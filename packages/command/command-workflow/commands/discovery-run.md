---
description: Discover a problem or opportunity without forcing one method, provider, or artifact shape
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
argument-hint: "<subject-or-native-reference> [--profile <reference>] [--provider <selection>] [--method <selection>]"
---

# /xonovex-workflow:discovery-run — Run Discovery

## Arguments

- `subject-or-native-reference` (required): Problem, opportunity, scope, or opaque native input reference.
- `--profile`, `--provider`, `--method` (optional): Explicit selections; otherwise use profile/environment resolution.

## Delegation

Load `plan-guide` (plugin `xonovex-skill-plan`) and perform **discovery-run**. Persist the
Discovery result as requested; do not assume files, Git, user stories, Gherkin, or one
agent session.
