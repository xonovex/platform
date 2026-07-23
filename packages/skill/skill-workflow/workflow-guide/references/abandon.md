# Abandon

## Goal

Stop work on one subject and return an inline abandonment record containing reason,
partial output, evidence, uncertainty, and retry boundary.

## Procedure

- [ ] Resolve the exact subject, revision, reason, partial results, and evidence.
- [ ] Inspect known effects and classify each as applied, failed, skipped, or unknown.
- [ ] State what is recoverable, what remains unresolved, and whether retry is safe.
- [ ] Preserve every source and provider resource.
- [ ] Return the abandonment `OperationResult` inline with `effect.mode: inspect`.

Abandon does not persist the record, change provider status, clean resources, or
authorize compensation. A host may checkpoint the record administratively under
policy, or the caller may Publish it separately as a domain artifact.

## Error handling

- Return blocked on a missing subject or reason.
- Require reconciliation before retry when an earlier write outcome is unknown.
- Never describe retained or uninspected state as removed.
