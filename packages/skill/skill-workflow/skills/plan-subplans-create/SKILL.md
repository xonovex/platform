---
description: "Generate detailed subplans from an approved parent plan, flagging steps that can run in parallel. Use when the user asks to break down, expand, or generate subplans from a plan. Keywords: plan, subplans, breakdown, parallel execution, expand plan, detailed steps."
---

# /xonovex-workflow:plan-subplans-create – Generate Detailed Subplans from Parent Plan

Generate detailed implementation subplans from an approved parent plan with parallel execution detection based on file/package dependencies.

## Prerequisites

Run before using this command:

- `/xonovex-workflow:plan-create` - Create and get approval for the parent plan

This command requires an approved parent plan. It does NOT perform codebase exploration - it relies on the parent plan context.

## Goal

- Read approved parent plan and extract proposed subplan structure
- Generate detailed child plans with implementation steps, code snippets, validation
- Detect parallel execution groups based on file/package overlap analysis
- Auto-associate with feature worktree via git config
- Save child plans and STOP (user runs /xonovex-workflow:plan-continue when ready)

## Core Workflow

**IMPORTANT: Do NOT use EnterPlanMode. Do NOT use Task/Explore agents - rely on parent plan context.**

1. **Read parent plan**: Load and parse approved parent plan
2. **Validate status**: Ensure parent plan is approved (status: `pending-approval` or `approved`)
3. **Extract context**: Goals, technology choices, proposed subplans, dependencies
4. **Identify child plans**: Use proposed subplan structure from parent plan
5. **Generate detailed content**: Objective, tasks with code snippets, file paths/line numbers, validation steps (typecheck/lint/build/test/integration), success criteria
6. **Analyze dependencies**: File overlap analysis -> parallel groups (independent), sequential (overlapping/dependent)
7. **Write child plan files**: Save to `<plan-dir>/<feature-name>/subplan-*.md`
8. **Update parent plan**: Add parallel_groups and dependencies.subplans
9. **Auto-associate worktree**: Set `git config branch.<branch>.plan` if in feature worktree
10. **Show summary**: Display created child plans and execution strategy; STOP (no implementation)

## Implementation Details

**Splitting**: Logical grouping (default) or by phase markers (if the user requests phase-based splitting)

**Dependency Detection**: File overlap -> Sequential; No overlap -> Parallel; Explicit deps -> Sequential with tracking

**Child Plan Frontmatter**: `type: plan`, `has_subplans: false`, `parent_plan`, `parallel_group`, `status: pending`, `dependencies: {plans: [], files: []}`, `skills_to_consult: [skill-names]`, `validation: {type_check: pending, lint: pending, build: pending, tests: pending, integration: pending}`

**Skills to Consult**: Every child plan MUST include `skills_to_consult` array listing applicable coding guidelines (e.g., `typescript-guidelines`, `testing-guidelines`, `hono-guidelines`). This ensures implementers consult project conventions before coding.

**Child Plan Sections**: Objective, Tasks (numbered with file paths, code snippets, actions), Validation Steps (typecheck, lint, build, test - all must pass), Success Criteria (checklist), Files Modified/Created (list), Dependencies (required plans), Estimated Duration

## Output

```
Created child plans for: plans/feature-name.md

Child Plans Created:
- plans/feature-name/01-add-library.md (pending)
- plans/feature-name/02-create-component.md (pending)
- plans/feature-name/03-integrate.md (pending)
- plans/feature-name/04-add-tests.md (pending)

Execution Strategy:
- Parallel Group 1: 01-add-library, 02-create-component
- Sequential Group 2: 03-integrate (depends on Group 1)
- Sequential Group 3: 04-add-tests (depends on Group 2)

Updated parent plan with parallel execution groups

Next Steps:
1. Review all child plans
2. Verify parallel execution groups and dependencies
3. Create workspace: invoke plan-worktree-create for feature-name
4. Start implementation: invoke plan-continue
5. Track progress: check overall plan completion status
```

## Error Handling

- Error: parent plan doesn't exist, parent plan not approved, child plans already exist, output dir fails
- Warning: >10 child plans (consider consolidation), circular dependencies, excessive file overlap
