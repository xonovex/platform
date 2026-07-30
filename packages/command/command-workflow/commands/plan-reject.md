---
description: "Draft: reject a plan with a reason, setting status rejected and record why, without deleting it"
allowed-tools:
  - Read
  - Edit
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "[plan-file] [reason]"
---

# /xonovex-workflow:plan-reject - Reject a Plan

> Lifecycle: research → decide → create → revise ⇄ critique → **reject** → (revise / discard)

## Arguments

- `plan-file` (optional): Path to plan document (auto-detects from git config or most recent plan in `plans/`)
- `reason` (optional): Why the plan is rejected; prompted for if omitted

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**reject** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas. Do not restate them.
