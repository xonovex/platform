# Continue: One Plan

Inspect, preview, or apply implementation of one explicit inline plan or provider-native plan reference plus an optional native revision. Default to `inspect`, operate on only the first actionable target, return the result inline, and stop.

## Core Workflow

1. **Resolve effect and input**: accept `inspect`, `preview`, or `apply`, defaulting to `inspect`. Use the explicit plan supplied by the caller. Let the selected provider resolve an opaque reference and optional revision; do not glob for an arbitrary plan or infer provider identity from a reference shape.
2. **Select one target**: when the plan relates ordered children, scan metadata only until the first child with unfinished tasks or unmet criteria. Treat status as a hint, verify it against evidence, and do not require a particular status value.
3. **Load the target**: resolve only that target fully, including tasks, dependencies,
   blockers, success criteria, sources, carried decisions, and
   `skills_to_consult`. Resolve every active opaque context reference through an
   explicitly selected provider; unresolved, stale, conflicting, or invalid context
   blocks before edits.
4. **Baseline**: discover the actual project toolchain and inspect applicable typecheck, lint, build, tests, and integration checks. Run non-mutating checks when useful and record unavailable or not-applicable categories.
5. **Consult skills**: load every applicable `skills_to_consult` capability and its relevant progressive references before implementation. Report any mandatory unavailable capability.
6. **Honor the effect**: for `inspect`, report the target, current evidence, blockers, and applicable checks without proposing edits. For `preview`, describe the exact intended changes, validation, and external effects without applying them. Only `apply` may implement the target; apply carried decisions as constraints, not instructions or evidence, preserve unrelated worktree changes, and resolve warnings at their source.
7. **Verify**: after `apply`, re-read every task and success criterion, run applicable validation, and distinguish new failures from baseline failures. For `inspect` or `preview`, state which verification would run without claiming it passed.
8. **Return one result**: report the selected effect, target, exact implementation
   revision, decisions retained or newly made, evidence, limitations,
   descriptive progress, and remaining work inline. Stop without persisting plan state
   or advancing to another child.

## Example

```text
Continue (apply): plans/order-import-backpressure/01-reader-pull.md

Target: subplan 01, the first with unfinished tasks (02 untouched)
Baseline: app:test 214/214 before edits; lint clean
Implemented: tasks 1 to 5; revision f4e5d6
Verification: app:test 219/219; app:ci-check green; criterion 2's
  counter asserts zero drops under the spike fixture
Criteria: 3 of 4 met; criterion 4 (staging soak) needs the staging
  feed, recorded unmet with its blocker named
Remaining: subplan 02, not started, per the one-target rule
```

The report separates baseline from new results, ties each criterion to
its evidence, and stops at one target.

## Gotchas

- Reconstruct provider-native inputs after context loss; conversation memory is not persistent identity.
- Retain invalidated or superseded context for history but do not apply it.
- A `complete` status does not prove tasks or criteria are satisfied, and another status does not block work requested by the caller.
- Passing tests do not prove every success criterion or Definition of Done item.
- Automatically continuing to the next child silently expands scope.
- Inspect and preview must not edit files, update providers, or mark plan tasks complete.
- A request to persist progress is a separate Publish operation.
