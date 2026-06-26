---
description: >-
  Research code-simplification opportunities (duplicates, dead code,
  over-abstraction) — read-only report that feeds a follow-up plan; makes no
  changes
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - AskUserQuestion
  - Skill
argument-hint: >-
  [path] [--aspects
  <duplicates,dead-code,dependencies,abstractions,interfaces,config>]
---

# /xonovex-workflow:plan-research-code-simplify — Research Code Simplification Opportunities

## Arguments

- `path` (required): Directory to analyze
- `--aspects` (optional): Aspects to check (comma-separated):
  - `duplicates` - Identical/near-identical functions
  - `utilities` - Logic for shared utilities
  - `types` - Redundant type definitions
  - `constants` - Scattered magic values
  - `patterns` - Repeated code patterns
  - `dead-code` - Unused functions, unreachable branches
  - `dependencies` - Unused imports, redundant packages
  - `abstractions` - Over-engineered patterns
  - `interfaces` - Complex APIs (>5 params, nested config)
  - `config` - Scattered configuration
  - `all` (default) - All aspects

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-research-code-simplify** operation with these arguments. The skill is the source
of truth for the procedure, output format, and gotchas — do not restate them.
