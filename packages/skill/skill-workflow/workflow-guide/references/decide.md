# Decide

## Goal

Resolve one question into a descriptive inline outcome and rationale while keeping
evidence, recommendation, outcome, and operational authority distinct.

## Procedure

- [ ] Resolve the exact question, options, evidence, criteria, revisions, and
      perspectives.
- [ ] Separate binding criteria from advisory considerations.
- [ ] Compare alternatives and record rationale, assumptions, uncertainty, and
      dissenting evidence.
- [ ] Record the descriptive outcome without applying a gate or protected mutation.
- [ ] Return the inline decision `OperationResult` with no effects.

Decide never approves, rejects, merges, promotes, releases, deploys, or publishes.
Deterministic provider and organization policy owns any later protected action.

## Error handling

- Return blocked on an unclear question or materially missing binding evidence.
- Keep a recommendation distinct when no authoritative outcome was supplied.
- If a protected action is requested, report that it remains a separate authorized
  operation.
