# plan-continue: Continue One Plan

Resume implementation from one explicit inline plan or provider-native plan reference plus an optional native revision. Start immediately, complete only the first actionable target, return changes and evidence, and stop.

## Core Workflow

1. **Resolve input** — use the explicit plan supplied by the caller. Let the selected provider resolve an opaque reference and optional revision; do not glob for an arbitrary plan or infer provider identity from a reference shape.
2. **Select one target** — when the plan relates ordered children, scan metadata only until the first child with unfinished tasks or unmet criteria. Treat status as a hint, verify it against evidence, and do not require a particular status value.
3. **Load the target** — resolve only that target fully, including tasks, dependencies, blockers, success criteria, sources, and `skills_to_consult`.
4. **Baseline** — discover the actual project toolchain and run applicable typecheck, lint, build, tests, and integration checks before changes. Record unavailable or not-applicable categories.
5. **Consult skills** — load every applicable `skills_to_consult` capability and its relevant progressive references before implementation. Report any mandatory unavailable capability.
6. **Execute** — track the target's pending work, implement all of it, preserve unrelated worktree changes, and resolve warnings at root cause.
7. **Verify** — re-read every task and success criterion, run applicable validation, and distinguish new failures from baseline failures.
8. **Return one result** — report changes, evidence, limitations, descriptive progress, and remaining work. Persist only when the caller requested a provider destination, then stop without advancing to another child.

## Gotchas

- Reconstruct provider-native inputs after context loss; conversation memory is not persistent identity.
- A `complete` status does not prove tasks or criteria are satisfied, and another status does not block work requested by the caller.
- Passing tests do not prove every success criterion or Definition of Done item.
- Automatically continuing to the next child silently expands scope.
