---
description: Verify that a plan or current work has been fully achieved
allowed-tools:
  - Bash
  - Read
  - Grep
  - Glob
  - Task
  - TaskCreate
  - TaskUpdate
  - TaskList
argument-hint: "[plan-file] [--detailed]"
---

# /plan-validate – Validate Plan Achievement

Verify that a plan document's objectives have been fully achieved, or validate that the most recently stated goal in the current conversation has been completed with evidence-based validation.

## Goal

- Validate plan completion with objective tests
- Check subplan status (if applicable)
- Report clear pass/fail with evidence
- Never modify files (read-only)

## Arguments

`/plan-validate [plan-file] [--detailed]`

- `plan-file` (optional): Path to plan document (if omitted, validates current conversation goal)
- `--detailed` (optional): Comprehensive analysis with full evidence trail

## Core Workflow

### With Plan Document

1. **Load Plan** - Read plan document, extract success criteria from phases
2. **Parse Metadata** - Extract status, phase, dependencies from frontmatter
3. **Check Subplans** - If subplans exist, read and report their status
4. **Design Tests** - Create validation tests for each phase criterion
5. **Execute Validation Checks** - Run ALL four mandatory checks in order:
   - Type check: `npx tsc --noEmit` (0 errors)
   - Linting: `npm run lint` (0 errors, 0 warnings)
   - Build: `npm run build` (succeeds)
   - Tests: `npm test` (all pass)
6. **Report** - Clear pass/fail result with evidence per phase and recommendations

**IMPORTANT**: This command is read-only and never modifies plan files. Use `/plan-update` to update status.

### Without Plan Document (Current Work)

1. **Identify Goal** - Review conversation to extract objective and success criteria
2. **Design Tests** - Create validation tests for each criterion
3. **Execute** - Run tests, collect evidence
4. **Report** - Clear pass/fail result with evidence and recommendations

## Success Criteria

- All phase requirements validated
- All subplans complete (if subplans exist)
- All tests pass (typecheck, lint, build, tests)
- Implementation matches specification

## Output

```
Validation Report

Phase 1: Setup PASS
Phase 2: Implementation IN PROGRESS
  FAIL Error handling incomplete (src/core.ts:45)

Files: PASS typecheck | FAIL lint (2 errors) | PASS test

Result: NOT COMPLETE

Next Steps:
1. If PASS: /plan-update - Update plan status to reflect completion
2. If FAIL: Fix identified issues (error handling, lint errors), then re-run validation
3. Check subplans: /plan-status plans/<plan>.md - See which subplans need work
4. Continue work: /plan-continue plans/<plan>.md - Resume implementation for incomplete items
5. Track progress: Monitor validation results in plan status section
```

## Error Handling

- Error if plan file not found or invalid format
- Warning if plan already marked complete
- Info if all tests pass (recommend `/plan-update`)

## Examples

```bash
# Validate current conversation goal
/plan-validate

# Validate specific plan
/plan-validate plans/auth.md
/plan-validate plans/auth.md --detailed
```
