# Workspace Cleanup

1. Resolve the exact named targets, optional target revisions, associated references,
   provider, applicable capabilities, and retry identity.
2. Detect dirty, unmerged, active, or otherwise protected state.
3. Preview the precise removal set and recovery information.
4. Require explicit `apply`; accept protected targets only when the caller explicitly
   authorizes force after seeing the risk.
5. Remove only the previewed set, then verify and report every observed effect.

For provider-native removal, require the supplied idempotency key when supported.
Reconcile unknown removal state before retrying and never widen the target set.

Workspace Cleanup is the only workspace operation that removes resources.
