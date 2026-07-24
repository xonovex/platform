# Mitigate and Recover

## Mitigation

1. Stabilize observation: exact targets, revisions, telemetry freshness, and known
   evidence gaps.
2. Rank candidate actions by expected harm reduction, time, reversibility, data risk,
   blast radius, required authority, and verification quality.
3. Preview the exact effect, preconditions, rollback or compensation, and stop
   conditions.
4. Authorize one controlled action through the accountable system.
5. Let the operations role apply it while others observe and record.
6. Verify the expected user and system response before applying another change.
7. Reconcile unknown effects before retrying, using the same idempotency key when
   supported.

Prefer feature disablement, traffic reduction, failover, capacity relief, or rollback
only when the current evidence and data compatibility make that option safer than
continued impact.

## Recovery

Recovery requires:

- the affected user journeys succeed from representative locations and identities;
- data and external side effects reconcile;
- dependencies, queues, caches, replicas, and delayed work are healthy;
- error, latency, saturation, and business indicators remain stable for a defined
  observation window;
- temporary controls and system divergence are documented;
- recurrence triggers and fallback actions are ready.

Close mitigation only after observed recovery. Remove temporary controls later through
separate planned effects.
