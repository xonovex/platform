# Publish

1. Resolve one exact subject, optional subject revision, one explicit destination,
   binding preconditions, expected destination revision, and applicable capabilities.
2. Default to `preview` and report the exact external write.
3. Before `apply`, require explicit authority, the expected destination revision when
   an existing destination exposes one, and a stable idempotency key when the provider
   supports one.
4. Apply only the previewed write, then verify the observed destination reference and
   revision.
5. Report failed or unknown effects without claiming total success.

When the provider does not support idempotency, name the non-idempotent retry boundary
and reconcile the exact destination before retrying an unknown write.

Publication persists a result; it does not imply approval.
