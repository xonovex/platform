# integration-validate: Preflight an Integration

## Core workflow

1. Resolve the accepted Deliverable Publication revision, current target revision, intended
   action, Acceptance reference/revision, prerequisite evidence, and current policy/profile.
2. Apply [operational-contracts.md](operational-contracts.md) to revalidate actor, subject,
   target, evidence, policy, conditions, and expiry bindings without changing the target.
3. Resolve an external-enforcement/provider adapter that maps the Integration intent to a
   protected target-side gate, declares bypass and failure behavior, and can return native
   decision, enforcement, mutation, and verification evidence.
4. Validate least-privilege credential release, immutable source/artifact provenance,
   idempotency, conflict and concurrency behavior, rollback readiness, cancellation, and
   partial-failure handling.
5. Publish a preflight result containing `ready`, `blocked`, or `stale`, every resolved
   binding, unsupported guarantee, required remediation, and adapter capability reference.

Preflight is read-only and does not reserve the target or make later drift safe. Execution
must re-resolve every binding immediately before the protected mutation.
