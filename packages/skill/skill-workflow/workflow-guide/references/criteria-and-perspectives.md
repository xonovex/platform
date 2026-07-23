# Criteria and Perspectives

## Meanings

A perspective adds questions, evidence needs, and advisory checks. A criterion states
something that can be evaluated. Perspective selection never grants authority and
never makes its suggested criteria binding.

Use repeatable perspectives. Canonicalize equivalent names, deduplicate shared
questions and evidence, and preserve each selection's provenance.

## Criterion authority

| Source                              | Binding rule                                            |
| ----------------------------------- | ------------------------------------------------------- |
| Explicit caller criterion           | Binding unless explicitly marked advisory.              |
| Exact authoritative parent artifact | Binding only after authority and revision are verified. |
| Mandatory runtime policy            | Binding regardless of method or perspective.            |
| Selected method                     | Advisory until its proposal is explicitly accepted.     |
| Perspective or role lens            | Advisory by default.                                    |
| Model inference                     | Advisory by default.                                    |

An advisory failure is a finding, not a failed acceptance gate. Overall validation
uses binding criteria only; report advisory pass, fail, and blocked results
separately.

## Assisted resolution

Assisted is the default:

- [ ] Collect explicit, inherited, policy, suggested, and inferred candidates.
- [ ] Assign a stable local ID and provenance to every candidate.
- [ ] Verify authority and exact revision before marking inherited criteria binding.
- [ ] Add high-confidence advisory checks when useful, but keep them advisory.
- [ ] Require explicit acceptance before promoting a suggestion.
- [ ] Preserve contradictions and ask or block when they affect a binding outcome.
- [ ] Record accepted, declined, and pending proposals in `OperationResult`.

Strict uses only explicit and verified authoritative criteria and reports missing
coverage. Automatic may select high-confidence advisory checks, but it still cannot
invent a binding validation, merge, release, or deployment gate.

## Selection precedence

For a single-valued semantic choice such as method, resolve in this order:

1. Explicit caller selection.
2. Explicitly accepted proposal from the chosen workflow or method.
3. Reported unambiguous inference from authoritative context.
4. Unresolved: ask, use a policy-declared safe default, or return blocked.

Mandatory policy is not a lower-priority suggestion; it constrains every level.
Never resolve conflicting facts by prompt order. Compare authority, scope, freshness,
exact revision, and evidence, and expose any remaining conflict.

## Perspective resolution

For each selected, inherited, suggested, or inferred perspective, record:

- value and provenance;
- why it applies;
- questions it added;
- criterion IDs it proposed;
- evidence it requires;
- supporting skill resolution, including visible unavailability.

If a specialist skill is unavailable, retain the perspective. Continue only when its
checks are advisory and useful baseline handling remains; block when policy makes the
specialist evidence mandatory.
