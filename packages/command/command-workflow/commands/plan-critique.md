---
description: "Draft: adversarially stress-test the written plan before approval; reports findings for plan-revise to fix"
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

> Lifecycle: research → decide → create → revise ⇄ **critique** → subplans-create → continue → update → validate

## Arguments

- `plan-file` (optional): Path to the plan document (auto-detects from git config or the most recent plan in `plans/`).
- `--mode <mode>` (optional): Which adversarial lens(es) to run — `red-team`, `pre-mortem`, `falsify`, `steelman`, or `all`. Default: `red-team` + `pre-mortem`.

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-critique** operation with these arguments. Run this as a fresh session /
independent agent that did not author the plan — self-critique defends instead of
attacks. The skill is the source of truth for the procedure, output format, and
gotchas — do not restate them.
