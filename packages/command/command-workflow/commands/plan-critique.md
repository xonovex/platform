---
description: >-
  Adversarially critique an existing plan to expose weaknesses before build —
  attack assumptions, run a pre-mortem, hunt the disconfirming case, and
  steelman the weakest part; reports findings for plan-refine to fix
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Task
  - Write
  - Skill
argument-hint: "[plan-file] [--mode red-team|pre-mortem|falsify|steelman|all]"
---

# /xonovex-workflow:plan-critique — Adversarially Critique a Plan

## Arguments

- `plan-file` (optional): Path to the plan document (auto-detects from git config or the most recent plan in `plans/`).
- `--mode <mode>` (optional): Which adversarial lens(es) to run — `red-team`, `pre-mortem`, `falsify`, `steelman`, or `all`. Default: `red-team` + `pre-mortem`.

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-critique** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
