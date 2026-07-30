# Effects and Authority

## Effect Modes

| Mode      | Permitted behavior                                                       |
| --------- | ------------------------------------------------------------------------ |
| `inspect` | Read external state without mutation                                     |
| `preview` | Resolve and report the exact effects that apply would attempt            |
| `apply`   | Perform only the explicitly authorized and previously understood effects |

The mode governs external state; where the result goes is a separate question.

## Modes Per Operation

This table is the single source of the default. No other file, command, or README
states it; they cite this one.

| Operation                                                            | Accepts                       | Default   |
| -------------------------------------------------------------------- | ----------------------------- | --------- |
| Create, review, revise, decide, validate, abandon, workspace abandon | `inspect` only, no `--effect` | `inspect` |
| Execute                                                              | every mode                    | `inspect` |
| Publish, workspace create, workspace merge, workspace cleanup        | every mode                    | `preview` |

The `inspect`-only operations may read the subject and its context wherever those
live, including through a provider capability, and they return their result to the
caller without writing anywhere.

## Plan, Validate, Execute

For preview or apply:

1. Resolve exact targets, available source revisions, expected destination revisions,
   preconditions, retry identity, and the requested effect mode.
2. Produce an exact effect plan that names additions, changes, removals, and external
   writes.
3. Validate target identity, current state, authority, preconditions, and recovery
   information.
4. Stop on ambiguity, stale revisions, missing required capability, or failed
   preconditions.
5. Apply only the validated plan, then verify and report observed state.

Do not infer apply from urgency, a trusted executor, a prior preview, a successful
validation, or a decision outcome.

A protected operation against provider-native state requires the exact source or
destination revision when that provider exposes one. When revision control is
unsupported, report the target as unversioned and its race consequence; block when the
requested criteria require stale-write protection.

An externally submitted apply operation requires a stable idempotency key when the
provider supports one. When it does not, name the non-idempotent retry boundary and
reconcile observed state before any retry. Local operations may instead use a
provider-independent natural identity when they can prove that repeating the exact
plan is safe.

A selected domain procedure cannot change the effect mode. In `inspect` or `preview`,
omit its mutation steps. If its useful result inherently requires mutation, return a
blocked result that names `apply` as the retry boundary.

## Effect Reporting

Classify each attempted effect:

- **planned**: included in preview but not attempted
- **applied**: verified in observed state
- **failed**: rejected or known not to have occurred
- **unknown**: submission may have happened but the observed outcome is unavailable

Never report total success when any effect failed or remains unknown. Reconcile unknown
effects before retrying, reusing the same idempotency key when the provider supports
one.

## Safety Boundary

- Keep destructive targets exact; do not widen paths, references, or resource sets.
- Treat opaque provider references as data for the selected provider capability.
- Preserve recovery information before destructive apply operations.
- Stop when completing the operation would require new authority or a materially
  broader effect.
