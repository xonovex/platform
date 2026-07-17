# Development Execution Contracts

Development consumes exact Planning revisions and publishes one independently resolvable Development result per assignment. Apply [results.md](results.md) and [providers.md](providers.md) without making Git, files, worktrees, agents, or one harness universal.

## Inputs and assignments

Resolve the parent and child Planning references, native revisions, dependencies, execution groups, success criteria, validation requirements, selected workspace provider, target development revision, profile, and executor constraints. Each assignment records its Planning source, scope, owner or executor, dependency group, isolated workspace reference, starting revision, permitted side effects, validation, cancellation, and result provider.

Concurrent assignments require distinct workspace references or a provider guarantee that prevents cross-assignment mutation. A dependency edge blocks pickup until the prerequisite Development result is available at the expected revision. Do not infer readiness from a shared branch name or runtime session.

## Executor selection

A work shape selects among the deterministic, model, and agent classes. Classify the assignment by taking the first match, which yields the least adaptive shape that fits:

1. a tool or script can produce and verify the result from exact inputs — `mechanical`;
2. otherwise, the whole input set is known and can be supplied up front and the output is schema-checkable — `bounded-transform`;
3. otherwise, the work must discover its own inputs through branching tool use — `adaptive`.

| Work shape          | Selected executor            | Required boundary                                                                                     |
| ------------------- | ---------------------------- | ----------------------------------------------------------------------------------------------------- |
| `mechanical`        | Deterministic script or tool | Exact inputs, tool/version, dry run where supported, idempotency, validation                          |
| `bounded-transform` | Bounded model                | Fixed inputs, output schema, validator, budget, retry ceiling, non-authoritative origin               |
| `adaptive`          | Bounded agent                | Purpose, result contract, tools/data/filesystem/network, depth, token/cost/time budgets, cancellation |

A shape outside these three literals is rejected rather than defaulted, and a model or agent whose required boundary is incomplete is rejected as unbounded.

`human` and `external` are separate classes that no work shape selects. Choose them orthogonally when authority or ownership — not difficulty — requires it: accountable judgment, approval, or another system of record. Qualification is an attribute of the human actor record, not a class of its own. Use **agent-governance-guide** for the executor class definitions, the actor record, and the semantics of an `--executor` request.

## Semantic intents and runtime adapters

Development may request session, tool, capability, result-publication, validation, subagent, workspace, and cancellation intents. Harness adapters own native event names, handler mechanics, ordering, concurrency, timeouts, permission behavior, and capability evidence. Workflow-only execution remains valid unless the selected profile explicitly requires a supported enforcement guarantee.

## Workspaces, consolidation, and abandonment

- Workspace providers own create, resolve, mutate, switch, snapshot, and remove operations. Git worktrees are one selectable adapter.
- Every Development result binds its starting and resulting workspace or subject revision, changes, validation, unresolved findings, executor origin, and partial-failure state.
- Consolidation consumes exact Development result revisions, applies their changes to a selected development workspace, resolves conflicts explicitly, and reruns required validation. It publishes another Development result.
- Consolidation never claims Acceptance, changes an accepted/protected integration target, or becomes Integration merely because it combines branches or workspaces.
- Abandonment publishes the reason, partial changes/evidence, validation state, reusable learning, cleanup state, and whether retry or replacement is allowed. It does not erase the assignment.

## Concurrency, retries, and failures

Declare ordering, maximum concurrency, idempotency key, retry ceiling, cancellation, and cleanup before execution. Preserve successful child results when siblings fail. A partial failure remains `partial` or `failed`; it is never coerced to success. Retry only failed or explicitly invalidated work, against a freshly resolved starting revision, using a stable idempotency key or provider-native duplicate protection. Reconcile duplicate or uncertain provider outcomes before another side effect.
