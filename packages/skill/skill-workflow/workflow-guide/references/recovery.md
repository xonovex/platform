# Failure and Recovery

Classify the outcome before continuing:

| Condition                         | Result and boundary                                                           |
| --------------------------------- | ----------------------------------------------------------------------------- |
| Capability timeout before effects | Failed or blocked; safe to retry the operation after availability returns.    |
| Known partial write               | Partial; preserve successful receipts and resume after the successful subset. |
| Unknown write outcome             | Partial or blocked; reconcile with the same idempotency key before retry.     |
| Conflicting evidence              | Block the affected conclusion and preserve both sources.                      |
| Missing retained evidence         | Mark criteria blocked and identify acceptable alternatives.                   |
| Unexpected resource topology      | Stop before effects and re-resolve bindings and policy.                       |
| Concurrent revision change        | Block; report expected and observed revisions.                                |

Always populate `status`, `effects`, `unresolved`, `uncertainty`, `retry`, and
`concurrency` in `OperationResult`. Never report total success when an effect or
binding is unknown.

Recovery guidance may recommend retry, compensation, rollback, replanning, or
escalation. Deterministic policy still enforces retry count, budget, authorization,
and immediate termination conditions.
