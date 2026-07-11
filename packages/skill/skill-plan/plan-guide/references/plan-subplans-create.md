# plan-subplans-create: Generate Detailed Subplans from Parent Plan

Generate detailed implementation subplans from an approved parent plan, with parallel-execution detection based on file/package overlap. Save child plans and STOP (user runs `plan-continue`).

## Prerequisites

An approved parent plan (run `plan-create` first). Does **NOT** explore the codebase — relies on parent-plan context.

## Core Workflow

**Do NOT switch into plan-authoring. Do NOT delegate to codebase-exploration agents.**

1. **Read parent plan** and validate `status: approved` (run `plan-accept` first)
2. **Extract context** — goals, technology choices, proposed subplans, dependencies
3. **Generate detailed child plans** — objective, tasks with code snippets and file paths/line numbers, validation steps (typecheck/lint/build/test/integration), success criteria
4. **Detect execution groups** — file overlap → sequential · no overlap → parallel · explicit deps → sequential with tracking
5. **Write child files** — `<plan-dir>/<feature-name>/subplan-*.md`
6. **Update parent** — add `parallel_groups` and `dependencies.subplans`
7. **Auto-associate worktree** — `git config branch.<branch>.plan` if in a feature worktree
8. **Show summary**, STOP (no implementation)

Splitting defaults to logical grouping, or by phase markers if the user requests phase-based.

## Child plan shape

- **Frontmatter** — `type: plan`, `has_subplans: false`, `parent_plan`, `parallel_group`, `status: pending`, `dependencies: {plans: [], files: []}`, `skills_to_consult: [skill-names]`, `validation: {type_check: pending, lint: pending, build: pending, tests: pending, integration: pending}`
- **`skills_to_consult`** — every child plan MUST include it; never empty
- **Sections** — Objective, Tasks (numbered, file paths, code snippets, actions), Validation Steps, Success Criteria (checklist), Files Modified/Created, Dependencies, Estimated Duration

## Gotchas

- Approving the parent plan (`plan-accept`) is mandatory — generating subplans against an unapproved parent skips review
- File-overlap analysis runs against the parent plan's listed files only — if the parent doesn't enumerate files clearly, parallel detection produces false-parallels
- A child plan without `skills_to_consult` skips project conventions during implementation — never empty
- > 10 child plans usually signals the parent is too broad — split into multiple parent plans
