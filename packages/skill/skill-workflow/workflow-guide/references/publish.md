# Publish

1. Resolve one exact subject, one explicit destination, binding preconditions, expected
   destination revision, and applicable capabilities.
2. Default to `preview` and report the exact external write.
3. Before `apply`, require explicit authority and a stable idempotency key.
4. Apply only the previewed write, then verify the observed destination reference and
   revision.
5. Report failed or unknown effects without claiming total success.

Publication persists a result; it does not imply approval.
