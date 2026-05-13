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

# /xonovex-workflow:plan-validate – Validate Plan Achievement

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
5. **Execute Validation Checks** - Run the project's standard checks in order (skip any that don't apply):
   - **Type check** (e.g. `tsc --noEmit`, `mypy`, `go vet`) — 0 errors
   - **Lint** (e.g. `eslint`, `ruff`, `golangci-lint`) — 0 errors, 0 warnings
   - **Build** (e.g. `npm run build`, `cargo build`, `go build`) — succeeds
   - **Tests** (e.g. `vitest`, `pytest`, `go test`) — all pass

   Detect the toolchain from the project (`package.json` scripts, Moon tasks, Makefile, language config) instead of hardcoding commands.

6. **Report** - Clear pass/fail result with evidence per phase and recommendations

**IMPORTANT**: This command is read-only and never modifies plan files. Use `/xonovex-workflow:plan-update` to update status.

### Without Plan Document (Current Work)

1. **Identify Goal** - Review conversation to extract objective and success criteria
2. **Design Tests** - Create validation tests for each criterion
3. **Execute** - Run tests, collect evidence
4. **Report** - Clear pass/fail result with evidence and recommendations

## Success Criteria

- All phase requirements validated
- All subplans complete (if subplans exist)
- All toolchain checks pass (typecheck / lint / build / tests)
- Implementation matches specification

## Output

```
Validation Report

Phase 1: Setup PASS
Phase 2: Implementation IN PROGRESS
  FAIL Error handling incomplete (src/core.ts:45)

Files: PASS typecheck | FAIL lint (2 errors) | PASS test

Result: NOT COMPLETE
```

## Error Handling

- Error if plan file not found or invalid format
- Warning if plan already marked complete
- Info if all tests pass (recommend `/xonovex-workflow:plan-update`)

## Examples

```bash
# Validate current conversation goal
/xonovex-workflow:plan-validate

# Validate specific plan
/xonovex-workflow:plan-validate plans/auth.md
/xonovex-workflow:plan-validate plans/auth.md --detailed
```

## Gotchas

- Detecting toolchain via `package.json` only misses Moon/Makefile-driven projects — check both
- "Tests pass" without checking coverage misses gaps in _what_ is tested — read success criteria, not just exit codes
- A subplan still `in_progress` blocks parent validation — surface the offending subplan explicitly
- Treating warnings as passes hides slow regressions — lint warnings count toward FAIL unless the project explicitly tolerates them
