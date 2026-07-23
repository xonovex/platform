# Review

## Goal

Return evidence-linked judgment about one exact subject without changing the subject
or persisting the report.

## Procedure

- [ ] Pin the subject and every evidence or criteria binding independently.
- [ ] Resolve repeatable perspectives, criteria authority, method, and implementation
      with provenance.
- [ ] Examine the subject under the selected questions and evidence requirements.
- [ ] Report findings, severity where relevant, evidence, and uncertainty without
      revising the subject.
- [ ] Return an inline review-report `OperationResult` with no effects.

Review may discuss each criterion but does not promise the criterion-by-criterion
pass/fail contract owned by Validate. A role lens adds suggestions only.

## Error handling

- Return blocked before reviewing an unpinned mutable subject.
- Keep missing advisory evidence visible without turning it into a binding failure.
- Block only the affected conclusion when binding evidence is unavailable or
  contradictory.
