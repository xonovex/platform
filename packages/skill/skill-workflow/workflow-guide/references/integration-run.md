# integration-run: Execute an Authorized Integration

## Core workflow

1. Resolve a fresh Integration preflight or repeat its checks against the exact accepted
   revision and current target revision.
2. Revalidate authorization, evidence, policy/profile, conditions, expiry, target state,
   external enforcement, credentials, provenance, and rollback through
   [operational-contracts.md](operational-contracts.md).
3. Ask the selected protected provider/external adapter to perform the explicit Integration
   capability. Never merge, write, deploy, or mutate the target through an ordinary tool
   path that bypasses that capability.
4. On success, partial failure, denial, timeout, cancellation, conflict, or ambiguous
   provider state, preserve the native decision, enforcement, mutation, and audit
   references. Fail closed where the profile requires it.
5. Re-resolve the target, verify the integrated revision and required invariants, and invoke
   the declared rollback capability when its trigger is met. Rollback is a new authorized
   action with its own outcome and evidence.
6. Publish an Integration result with accepted source revision, starting and resulting
   target revisions, authorization, actor/executor, action, outcome, verification,
   rollback/recovery, and residual gaps.
