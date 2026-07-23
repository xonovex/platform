# Publish

## Goal

Preview or apply persistence or transmission of one exact domain result or artifact
to one explicit provider destination.

## Procedure

- [ ] Resolve the exact artifact binding or immediately preceding result and verify
      its source revisions.
- [ ] Resolve the independent destination provider, opaque reference, and expected
      revision.
- [ ] Derive the provider adapter and produce the exact write preview.
- [ ] On apply, require deterministic authorization, an idempotency key, and unchanged
      concurrency preconditions.
- [ ] Apply once, verify the provider receipt, and return the native locator and
      revision in `OperationResult`.

Publish is the only domain-result persistence operation. It never implies approval,
acceptance, readiness, merge, or release. A publish destination may be the same
logical reference as an input only when the expected revision pins the successor
write.

## Error handling

- Return blocked on a missing destination, mutable unpinned artifact, ambiguous
  provider, denied authority, or stale expected revision.
- Never fall back to another provider or local file.
- On an unknown write outcome, retain the idempotency key and reconcile before retry.
