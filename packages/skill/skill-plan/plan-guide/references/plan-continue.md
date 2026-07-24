# plan-continue: Continue One Plan

Inspect, preview, or apply implementation of one explicit inline plan or provider-native plan reference plus an optional native revision. Default to `inspect`, operate on only the first actionable target, return the result inline, and stop.

## Core Workflow

1. **Resolve effect and input** — accept `inspect`, `preview`, or `apply`, defaulting to `inspect`. Use the explicit plan supplied by the caller. Let the selected provider resolve an opaque reference and optional revision; do not glob for an arbitrary plan or infer provider identity from a reference shape.
2. **Select one target** — when the plan relates ordered children, scan metadata only until the first child with unfinished tasks or unmet criteria. Treat status as a hint, verify it against evidence, and do not require a particular status value.
3. **Load the target** — resolve only that target fully, including tasks, dependencies, blockers, success criteria, sources, and `skills_to_consult`.
4. **Baseline** — discover the actual project toolchain and inspect applicable typecheck, lint, build, tests, and integration checks. Run non-mutating checks when useful and record unavailable or not-applicable categories.
5. **Consult skills** — load every applicable `skills_to_consult` capability and its relevant progressive references before implementation. Report any mandatory unavailable capability.
6. **Honor the effect** — for `inspect`, report the target, current evidence, blockers, and applicable checks without proposing edits. For `preview`, describe the exact intended changes, validation, and external effects without applying them. Only `apply` may implement the target; preserve unrelated worktree changes and resolve warnings at their source.
7. **Verify** — after `apply`, re-read every task and success criterion, run applicable validation, and distinguish new failures from baseline failures. For `inspect` or `preview`, state which verification would run without claiming it passed.
8. **Return one result** — report the selected effect, target, evidence, limitations, descriptive progress, and remaining work inline. Stop without persisting plan state or advancing to another child.

## Gotchas

- Reconstruct provider-native inputs after context loss; conversation memory is not persistent identity.
- A `complete` status does not prove tasks or criteria are satisfied, and another status does not block work requested by the caller.
- Passing tests do not prove every success criterion or Definition of Done item.
- Automatically continuing to the next child silently expands scope.
- Inspect and preview must not edit files, update providers, or mark plan tasks complete.
- A request to persist progress is a separate Publish operation.
