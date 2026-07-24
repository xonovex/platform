# Effects and Authority

## Effect Modes

| Mode      | Permitted behavior                                                       |
| --------- | ------------------------------------------------------------------------ |
| `inline`  | Return a result without external mutation                                |
| `inspect` | Read external state without mutation                                     |
| `preview` | Resolve and report the exact effects that apply would attempt            |
| `apply`   | Perform only the explicitly authorized and previously understood effects |

Create, review, revise, decide, validate, abandon, and workspace abandon return inline
results. Execute may inspect, preview, or apply. Publish, workspace create, workspace
merge, and workspace cleanup default to preview and may apply.

## Plan, Validate, Execute

For preview or apply:

1. Resolve exact targets, expected revisions, preconditions, and the requested effect
   mode.
2. Produce an exact effect plan that names additions, changes, removals, and external
   writes.
3. Validate target identity, current state, authority, preconditions, and recovery
   information.
4. Stop on ambiguity, stale revisions, missing required capability, or failed
   preconditions.
5. Apply only the validated plan, then verify and report observed state.

Do not infer apply from urgency, a trusted executor, a prior preview, a successful
validation, or a decision outcome.

A selected domain procedure cannot change the effect mode. In `inspect` or `preview`,
omit its mutation steps. If its useful result inherently requires mutation, return a
blocked result that names `apply` as the retry boundary.

## Effect Reporting

Classify each attempted effect:

- **planned** — included in preview but not attempted
- **applied** — verified in observed state
- **failed** — rejected or known not to have occurred
- **unknown** — submission may have happened but the observed outcome is unavailable

Never report total success when any effect failed or remains unknown. Reconcile unknown
effects before retrying, reusing the same idempotency key when the provider supports
one.

## Safety Boundary

- Keep destructive targets exact; do not widen paths, references, or resource sets.
- Treat opaque provider references as data for the selected provider capability.
- Preserve recovery information before destructive apply operations.
- Stop when completing the operation would require new authority or a materially
  broader effect.
