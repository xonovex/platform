# Revise

## Goal

Produce one traceable inline successor from explicit feedback while preserving the
pinned source, feedback provenance, and unresolved items.

## Procedure

- [ ] Resolve the exact source revision and every named feedback binding.
- [ ] Resolve constraints, perspectives, method, and implementation with provenance.
- [ ] Classify feedback as applied, deferred, conflicting, or rejected with reason.
- [ ] Produce a new inline result without editing or persisting the source.
- [ ] Return an inline `OperationResult` linking the successor to its source and
      feedback.

Revise always uses `effect.mode: inspect`. To store the successor, pass its exact
result to Publish. Publishing to the same logical provider reference requires an
expected-revision precondition and creates a provider-native successor; it never
rewrites the pinned source revision.

## Error handling

- Return blocked on missing feedback or an unpinned mutable source.
- Preserve conflicting feedback rather than choosing by prompt order.
- Return partial when a useful successor exists but material feedback remains
  unresolved.
