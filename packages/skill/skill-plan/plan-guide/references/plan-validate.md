# plan-validate: Validate Plan Achievement

Verify a plan document's objectives have been fully achieved, or validate the most recently stated goal in the current conversation. Read-only — never modifies plan files.

## Goal

- Validate plan completion with objective tests
- Check subplan status (if applicable)
- Report clear pass / fail with evidence
- Never modify files

## Core Workflow

### With Plan Document

1. **Load plan** — read document, extract success criteria from phases
2. **Parse metadata** — extract status, phase, dependencies from frontmatter
3. **Check subplans** — if subplans exist, read and report their status
4. **Design tests** — create validation tests for each phase criterion
5. **Execute validation checks** — run the project's standard checks in order (skip any that don't apply):
   - **Type check** (e.g. `tsc --noEmit`, `mypy`, `go vet`) — 0 errors
   - **Lint** (e.g. `eslint`, `ruff`, `golangci-lint`) — 0 errors, 0 warnings
   - **Build** (e.g. `npm run build`, `cargo build`, `go build`) — succeeds
   - **Tests** (e.g. `vitest`, `pytest`, `go test`) — all pass

   Detect the toolchain from the project (`package.json` scripts, Moon tasks, Makefile, language config) instead of hardcoding commands.

6. **Report** — clear pass / fail per phase with evidence and recommendations

**IMPORTANT:** Read-only; use `plan-update` to update status.

### Without Plan Document (Current Work)

1. **Identify goal** — review conversation to extract objective and success criteria
2. **Design tests** — create validation tests for each criterion
3. **Execute** — run tests, collect evidence
4. **Report** — clear pass / fail with evidence and recommendations

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

- Plan file not found / invalid format → error
- Plan already marked complete → warning
- All tests pass → info (recommend `plan-update`)

## Gotchas

- Detecting toolchain via `package.json` only misses Moon/Makefile-driven projects — check both
- "Tests pass" without checking coverage misses gaps in _what_ is tested — read success criteria, not just exit codes
- A subplan still `in_progress` blocks parent validation — surface the offending subplan explicitly
- Treating warnings as passes hides slow regressions — lint warnings count toward FAIL unless the project explicitly tolerates them
