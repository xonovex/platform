# plan-subplans-create: Generate Detailed Subplans from Parent Plan

Generate detailed implementation subplans from an approved parent plan, with parallel-execution detection based on file/package overlap.

## Prerequisites

An approved parent plan (run `plan-create` first). This command does NOT perform codebase exploration; it relies on the parent plan's context.

## Goal

- Read approved parent plan; extract proposed subplan structure
- Generate detailed child plans with implementation steps, code snippets, validation
- Detect parallel execution groups via file/package overlap
- Auto-associate with feature worktree via git config
- Save child plans and STOP (user runs `plan-continue` when ready)

## Core Workflow

**IMPORTANT: Do NOT switch into a plan-authoring mode. Do NOT delegate to codebase-exploration agents — rely on parent plan context.**

1. **Read parent plan** — load and parse
2. **Validate status** — parent must be `pending-approval` or `approved`
3. **Extract context** — goals, technology choices, proposed subplans, dependencies
4. **Identify child plans** — use proposed subplan structure
5. **Generate detailed content** — objective, tasks with code snippets, file paths/line numbers, validation steps (typecheck/lint/build/test/integration), success criteria
6. **Analyze dependencies** — file overlap → sequential; no overlap → parallel; explicit deps → sequential with tracking
7. **Write child plan files** — save to `<plan-dir>/<feature-name>/subplan-*.md`
8. **Update parent plan** — add `parallel_groups` and `dependencies.subplans`
9. **Auto-associate worktree** — `git config branch.<branch>.plan` if in feature worktree
10. **Show summary** — created plans + execution strategy; STOP (no implementation)

## Implementation Details

**Splitting:** logical grouping (default), or by phase markers if user requests phase-based splitting

**Dependency detection:** file overlap → sequential; no overlap → parallel; explicit deps → sequential with tracking

**Child plan frontmatter:** `type: plan`, `has_subplans: false`, `parent_plan`, `parallel_group`, `status: pending`, `dependencies: {plans: [], files: []}`, `skills_to_consult: [skill-names]`, `validation: {type_check: pending, lint: pending, build: pending, tests: pending, integration: pending}`

**Skills to consult:** every child plan MUST include `skills_to_consult` so implementers know project conventions

**Child plan sections:** Objective, Tasks (numbered with file paths, code snippets, actions), Validation Steps, Success Criteria (checklist), Files Modified/Created, Dependencies, Estimated Duration

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
```

## Error Handling

- Parent plan doesn't exist / not approved / child plans already exist → error
- Output dir creation fails → error
- > 10 child plans → warning (consider consolidation)
- Circular dependencies / excessive file overlap → warning

## Gotchas

- Approving the parent plan is mandatory — generating subplans against `draft` parent skips review
- File-overlap analysis runs against the parent plan's listed files only — if the parent doesn't enumerate files clearly, parallel detection produces false-parallels
- A child plan without `skills_to_consult` will skip project conventions during implementation — never empty
- > 10 child plans usually signals the parent plan is too broad — split into multiple parent plans
