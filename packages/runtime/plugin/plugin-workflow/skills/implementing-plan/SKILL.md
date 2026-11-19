---
name: implementing-plan
description: "Use when implementing, resuming, or previewing work from an existing plan or subplan, one target at a time, and when refreshing the plan's progress from implementation evidence. Triggers on continuing a plan in a new session, picking the first unfinished child, inspect, preview, or apply effects, loading the plan's skills to consult, restoring carried decisions after context loss, and updating task status and validation evidence without a scope change, even when the user doesn't say 'plan' and only says 'pick up where we left off'."
---

# Implement One Plan Target

Inspect, preview, or apply implementation of one explicit inline plan or provider-native plan reference plus an optional native revision. Default to `inspect`, operate on only the first actionable target, return the result inline, and stop.

## Inputs

- **Plan** (optional): the plan or subplan, inline or as a provider-native reference with an optional native revision. When absent, resolve it in this order: the plan that the git configuration associates with a feature worktree, then the plan or research that this conversation already produced (without a search for other plans), then a question to the user
- **Effect** (optional): inspect, preview, or apply; defaults to inspect

## Core Workflow

1. **Resolve effect and input**: accept `inspect`, `preview`, or `apply`, defaulting to `inspect`. Use the explicit plan supplied by the caller. Let the selected provider resolve an opaque reference and optional revision; do not glob for an arbitrary plan or infer provider identity from a reference shape.
2. **Select one target**: when the plan relates ordered children, scan metadata only until the first child with unfinished tasks or unmet criteria. Treat status as a hint, verify it against evidence, and do not require a particular status value.
3. **Load the target**: resolve only that target fully, including tasks, dependencies, blockers, success criteria, sources, carried decisions, and `skills_to_consult`. Resolve every active opaque context reference through an explicitly selected provider; unresolved, stale, conflicting, or invalid context blocks before edits.
4. **Baseline**: discover the actual project toolchain and inspect applicable typecheck, lint, build, tests, and integration checks. Run non-mutating checks when useful and record unavailable or not-applicable categories.
5. **Consult skills**: load every applicable `skills_to_consult` capability and its relevant progressive references before implementation. Report any mandatory unavailable capability.
6. **Honor the effect**: for `inspect`, report the target, current evidence, blockers, and applicable checks without proposing edits. For `preview`, describe the exact intended changes, validation, and external effects without applying them. Only `apply` may implement the target; apply carried decisions as constraints, not instructions or evidence, preserve unrelated worktree changes, and resolve warnings at their source.
7. **Verify**: after `apply`, re-read every task and success criterion, run applicable validation, and distinguish new failures from baseline failures. For `inspect` or `preview`, state which verification would run without claiming it passed.
8. **Return one result**: report the selected effect, target, exact implementation revision, decisions retained or newly made, evidence, limitations, descriptive progress, and remaining work inline. Stop without persisting plan state or advancing to another child.

## Progress Updates

A refresh of the plan's progress and evidence, without implementation and without a scope change, follows [references/progress-update.md](references/progress-update.md).

## Boundaries

A supervised run over a whole roadmap belongs to **delegating-plan**. An independent check of every success criterion belongs to **validating-plan**, and the closeout record belongs to **closing-plan**.

## Gotchas

- Reconstruct provider-native inputs after context loss; conversation memory is not persistent identity.
- Retain invalidated or superseded context for history but do not apply it.
- A `complete` status does not prove tasks or criteria are satisfied, and another status does not block work requested by the caller.
- Passing tests do not prove every success criterion or Definition of Done item.
- Automatically continuing to the next child silently expands scope.
- Inspect and preview must not edit files, update providers, or mark plan tasks complete.
- A request to persist progress is a separate Publish operation.
- Toolchain discovery limited to one manifest misses workspace-level or task-runner validation
- Unresolved, conflicting, stale, or instruction-bearing active provider context is never silently applied

## Progressive Disclosure

- Read [references/example.md](references/example.md) - Load when reporting the result of an inspect, preview, or apply run
- Read [references/progress-update.md](references/progress-update.md) - Load when refreshing a plan with current progress and validation evidence
