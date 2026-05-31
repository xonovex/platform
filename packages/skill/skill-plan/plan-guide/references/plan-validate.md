# plan-validate: Validate Plan Achievement

Verify a plan document's objectives have been fully achieved: success criteria pass
AND the code structurally matches what the plan describes. Read-only — never modifies files.

## Goal

- Run success criteria checks (build, test, grep commands)
- Audit code vs plan structure (types, functions, naming, data flow)
- Check cross-subplan consistency (shared concepts described the same way)
- Report clear pass/fail with evidence
- Never modify files

## Core Workflow

### With Plan Document

1. **Load plan** — read document, extract tasks, success criteria, file lists
2. **Parse metadata** — extract status, dependencies from frontmatter
3. **Check subplans** — if subplans exist, read and report their status
4. **Criteria checks** — run each success criterion command and report pass/fail:
   - Type check, lint, build, tests — detect toolchain from project, don't hardcode
   - Grep checks from the plan (e.g. "zero direct calls outside X")
5. **Structural audit** — for each plan task, search the codebase to verify:
   - Types/functions mentioned in the plan exist in the code
   - Files created/removed as stated
   - Naming matches (plan says `foo_t`, code doesn't still use the old name `bar_t`)
   - Data flow matches (plan says "A reads from B", code doesn't have A reading from C)
   - Report deviations as: "plan says X, code has Y"
6. **Cross-subplan consistency** — shared concepts named the same way across subplans,
   no contradictions, dependencies reference things that exist
7. **Report** — per-criterion PASS/FAIL, per-task match/deviation, consistency issues

**IMPORTANT:** Read-only; use `plan-update` to update status.

### Without Plan Document (Current Work)

1. **Identify goal** — review conversation to extract objective and success criteria
2. **Design tests** — create validation tests for each criterion
3. **Execute** — run tests, collect evidence
4. **Report** — clear pass/fail with evidence and recommendations

## Gotchas

- Criteria checks alone miss structural drift — code can pass all greps but implement something different from what the plan describes
- Cross-subplan consistency issues compound silently — one wrong name in an early subplan propagates through all later ones
- "Tests pass" without checking coverage misses gaps — read success criteria, not just exit codes
- A subplan still `in_progress` blocks parent validation — surface it explicitly
