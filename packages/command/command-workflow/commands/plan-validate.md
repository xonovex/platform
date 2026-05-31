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

Verify that a plan document's objectives have been fully achieved: both that success
criteria pass (builds, tests, grep checks) AND that the code structurally matches what
the plan describes (types, functions, data flow). Read-only — never modifies files.

## Goal

- Run success criteria checks (build, test, grep commands)
- Audit code vs plan structure (does the code match what the plan says?)
- Check cross-subplan consistency (shared concepts described the same way)
- Report clear pass/fail with evidence
- Never modify files

## Arguments

`/plan-validate [plan-file] [--detailed]`

- `plan-file` (optional): Path to plan document (if omitted, validates current conversation goal)
- `--detailed` (optional): Comprehensive analysis with full evidence trail

## Core Workflow

### With Plan Document

1. **Load Plan** — read plan document, extract tasks, success criteria, file lists
2. **Parse Metadata** — extract status, dependencies from frontmatter
3. **Check Subplans** — if subplans exist, read and report their status
4. **Criteria Checks** — run each success criterion command and report pass/fail:
   - **Type check** (e.g. `tsc --noEmit`, `cmake --build`, `go vet`) — 0 errors
   - **Lint** (e.g. `eslint`, `ruff`) — 0 errors, 0 warnings
   - **Build** — succeeds
   - **Tests** — all pass
   - **Grep checks** — any grep-based criteria from the plan (e.g. "zero direct calls outside X")

   Detect the toolchain from the project, don't hardcode commands.

5. **Structural Audit** — for each task in the plan, search the codebase to verify:
   - Types/structs mentioned in the plan exist in the code
   - Functions/kernels mentioned in the plan exist and have the described signature
   - File paths mentioned in the plan match actual files (created/removed as stated)
   - Naming matches (plan says `foo_t`, code doesn't still use the old name `bar_t`)
   - Data flow matches (plan says "A reads from B", code doesn't have A reading from C instead)

   Report deviations as: "plan says X, code has Y"

6. **Cross-Subplan Consistency** (if parent plan with subplans) — check that shared
   concepts are described consistently across subplans:
   - Same naming for the same types/streams/functions
   - No contradictory statements (e.g. one subplan says "camera is an entity", another says "camera is a struct")
   - Dependencies between subplans reference things that actually exist in the depended-on subplan

7. **Report** — per-criterion PASS/FAIL, per-task structural match/deviation, cross-subplan consistency issues

**IMPORTANT**: This command is read-only and never modifies plan files. Use `/xonovex-workflow:plan-update` to update status.

### Without Plan Document (Current Work)

1. **Identify Goal** — review conversation to extract objective and success criteria
2. **Design Tests** — create validation tests for each criterion
3. **Execute** — run tests, collect evidence
4. **Report** — clear pass/fail result with evidence and recommendations

## Output

```
Validation Report: plans/feature.md

Criteria Checks:
  [PASS] Build: clean
  [PASS] Tests: 12/12 pass
  [FAIL] Grep: "zero direct destroy calls" — 2 hits outside drain pass

Structural Audit:
  [MATCH] user_service_t exists in services/user_service.h
  [DEVIATION] plan says "reads from cache", code reads from database directly
  [MISSING] plan lists validation_middleware, file does not exist

Cross-Subplan Consistency:
  [CONFLICT] subplan 02 says "UserDTO", subplan 04 says "UserResponse"

Result: NOT COMPLETE (1 criteria fail, 1 deviation, 1 missing, 1 conflict)
```

## Error Handling

- Error if plan file not found or invalid format
- Warning if plan already marked complete
- Info if all checks pass (recommend `/xonovex-workflow:plan-update`)

## Examples

```bash
/xonovex-workflow:plan-validate
/xonovex-workflow:plan-validate plans/auth.md
/xonovex-workflow:plan-validate plans/auth.md --detailed
```

## Gotchas

- Detecting toolchain via `package.json` only misses Moon/Makefile-driven projects — check both
- "Tests pass" without checking coverage misses gaps — read success criteria, not just exit codes
- A subplan still `in_progress` blocks parent validation — surface it explicitly
- Criteria checks alone miss structural drift — the code can pass all greps but implement something different from what the plan describes
- Cross-subplan consistency issues compound silently — one wrong name in an early subplan propagates through all later ones
