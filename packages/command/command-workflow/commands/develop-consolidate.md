---
description: Consolidate exact Development results into a validated development workspace without claiming Acceptance or Integration
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
argument-hint: "<development-reference...> --target-workspace <reference> [--revision <native-revision>] [--target-revision <native-revision>] [--conflict-policy <selection>]"
---

# /xonovex-workflow:develop-consolidate — Consolidate Development

## Arguments

- `development-reference...` (required): Opaque source Development references.
- `--revision` (repeatable): Exact native source revisions.
- `--target-workspace` (required): Selected development workspace provider reference.
- `--target-revision` (optional): Exact starting workspace revision; otherwise resolve and record it before mutation.
- `--conflict-policy` (optional): Explicit deterministic or authority-bound conflict strategy.

## Delegation

Perform **develop-consolidate** with the selected workspace and Git skills for native
mechanics. Preserve constituent results, rerun consolidated validation, and stop if the
requested target change exceeds the caller's authorized scope.
