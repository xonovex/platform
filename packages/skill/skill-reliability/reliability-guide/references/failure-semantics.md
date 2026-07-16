# Failure Semantics

## Use one end-to-end deadline

Propagate a deadline or remaining budget across dependencies. Set connection, request, queue, execution, model/tool, and cleanup bounds inside it. On expiry, cancel work where supported, stop launching downstream work, and record whether side effects may still complete.

Timeout is an unknown outcome for side-effecting operations until authoritative state is reconciled.

## Retry only safe operations

Retry transient failures only when the operation is idempotent or protected by a stable idempotency key and reconciliation. Bound attempts and total time, use exponential backoff with jitter, honor server guidance, cap concurrency, and stop on permanent, authorization, validation, stale-version, or cancellation errors.

Record attempt count and one logical operation identity without treating retries as independent successes.

## Declare concurrency and ordering

State whether calls are serial, parallel, reentrant, duplicate-safe, commutative, partition-ordered, or globally ordered. Test simultaneous invocation, same-key races, cancellation, restart, and delayed completion. Never infer serial behavior from local test order or hook registration order.

Use compare-and-swap, version preconditions, transactions, leases, queues, or reconciliation according to the state owner. Preserve partial outcomes rather than reporting the batch as wholly successful.

## Bound overload

Set concurrency, queue, memory, token, cost, and provider-rate budgets. Prefer admission control, backpressure, load shedding, priority, circuit breaking, and graceful degradation over unbounded queues and retry storms.

Protect critical policy, rollback, emergency-disable, health, and recovery paths from ordinary workload saturation. A telemetry or advisory sink may shed load and fail visibly without blocking a mandatory control path.

## Make failure policy explicit

For every dependency and module select:

- `fail-closed` when proceeding would violate a mandatory guarantee and adequate enforcement exists;
- `fail-visible` when work may continue but the degraded guarantee and evidence gap must be explicit;
- `advisory` when loss of the optional signal cannot block the action.

Test the chosen behavior during policy, identity, configuration, evidence, telemetry, network, model, tool, and provider outages. Never silently translate unknown into allow or success.
