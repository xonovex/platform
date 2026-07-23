# Effects, Authority, and Idempotency

## Effect modes

- `inspect` reads and reasons without proposing or applying a provider mutation.
- `preview` resolves exact targets, preconditions, and proposed effects but applies
  none.
- `apply` requests the previewed effects; it does not itself grant authority.

Allowed modes are operation-specific in
[contracts.md](contracts.md). Unsupported modes are invalid requests, not fallbacks.

## Deterministic ownership

The runtime and provider boundary own authentication, authorization, approval
enforcement, policy, budgets, audit logging, retries, idempotency coordination, and
kill switches. A skill explains the requested work but cannot grant or self-assert
any of those controls.

Before apply:

- [ ] Resolve exact targets and current revisions.
- [ ] Produce or verify the effect preview.
- [ ] Verify authorization and required approvals outside model judgment.
- [ ] Establish an idempotency key or provider-native equivalent for retryable writes.
- [ ] Recheck optimistic-concurrency preconditions.
- [ ] Stop if the preview changed after approval.

Record each effect separately with target, mode, status, authorization reference,
idempotency key, provider receipt, and error where available. Never convert requested
or proposed into applied without observed evidence.

## Unknown outcomes

When a write times out or loses its response:

1. Mark the effect `unknown`, not failed or applied.
2. Preserve the same idempotency key and exact target.
3. Reconcile provider state or receipt before any retry.
4. Resume after the last known-safe effect; do not replay confirmed effects.
5. Return partial or blocked when reconciliation cannot establish the outcome.

Approval, a ready decision, passing validation, or an explicit apply request never
widens the exact approved scope.
