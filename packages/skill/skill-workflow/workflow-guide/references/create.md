# Create

## Goal

Produce one new inline domain result from exact inputs without changing or
persisting the source or result.

## Procedure

- [ ] Normalize the subject, named inputs, output kind, method, perspectives, and
      criteria.
- [ ] Resolve exact bindings and derived implementation selections.
- [ ] Apply assisted criterion resolution; method and perspective proposals remain
      advisory until accepted.
- [ ] Create an inline result that preserves source provenance and satisfies evaluated
      binding criteria.
- [ ] Return the human summary and inline `OperationResult`; leave `effects` empty.

Create always uses `effect.mode: inspect`. A request for a file, tracker item,
document, or other persisted destination requires a later Publish operation.

## Error handling

- Return blocked on an ambiguous subject, provider, required kind, or binding
  criterion.
- Report unavailable required skills or adapters without substitution.
- Preserve useful partial output and uncertainty rather than claiming completion.
