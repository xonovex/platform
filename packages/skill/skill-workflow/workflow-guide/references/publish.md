# Publish

1. Resolve one exact subject, its optional revision, one explicit destination, binding
   preconditions, the expected destination revision, and applicable capabilities.
2. When the subject is forwarded context, check it carries no secrets or detail the
   destination's audience should not see.
3. Report the exact external write in the mode [effects.md](effects.md) assigns.
4. Before `apply`, require explicit authority, the expected destination revision when
   the destination exposes one, and a stable idempotency key when the provider
   supports one.
5. Reconcile the exact destination immediately before `apply`. An identical write
   already present is a no-op; a destination that has moved blocks.
6. Apply only the previewed write, then verify the observed destination reference and
   revision.
7. Report failed or unknown effects without claiming total success.

A provider-native note or pull-request comment is a valid destination for selected
context. Return its native reference so later operations can supply it through
`--context` or a handoff. Never publish all session context implicitly.

Where the provider offers no idempotency key, name the non-idempotent retry boundary
and reconcile the destination before retrying an unknown write.

Publication persists a result; it does not imply approval. See
[governance.md](governance.md).
