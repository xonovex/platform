# Validate

## Goal

Return reproducible criterion-by-criterion evidence about one exact subject without
changing the subject or persisting the evidence result.

## Procedure

- [ ] Pin subject, criteria sources, supporting evidence, and revisions.
- [ ] Resolve criterion provenance and binding status under assisted resolution.
- [ ] Evaluate every selected criterion as pass, fail, or blocked and record evidence
      freshness and reproduction information.
- [ ] Compute the binding validation outcome from binding criteria only; keep advisory
      outcomes separate.
- [ ] Return an inline evidence `OperationResult` with no domain persistence effect.

Validation does not accept, publish, revise, merge, release, or deploy. An advisory
failure cannot silently fail a binding gate.

## Error handling

- Return blocked when no binding criteria can be resolved for a required gate.
- Mark only an unevaluable criterion blocked and identify missing evidence.
- Return blocked on an unpinned mutable subject or changed authoritative criteria
  revision.
