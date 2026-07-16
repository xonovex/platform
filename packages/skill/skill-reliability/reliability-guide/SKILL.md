---
name: reliability-guide
description: "Use when designing, assessing, or operating reliability and resilience for services, automation, agents, policy systems, governance modules, and privileged changes. Triggers on SLOs or error budgets, failure semantics, deadlines, retries, idempotency, concurrency, backpressure, degraded modes, capacity, dependency outages, backup and restore, continuity, canary updates, rollback, emergency disable, incident recovery, corrective action, operational metrics, or safe retirement, even when the user doesn't say 'reliability'."
---

# Reliability and Resilience

Design for observable user and control outcomes through failure, change, recovery, and retirement without turning telemetry into individual surveillance.

## Essentials

- **Define outcome objectives** — bind service and control outcomes to users, indicators, targets, windows, dependencies, capacity, evidence, and response, see [references/service-objectives.md](references/service-objectives.md)
- **Specify failure semantics** — set deadlines, retries, idempotency, concurrency, ordering, backpressure, cancellation, overload, and degraded behavior explicitly, see [references/failure-semantics.md](references/failure-semantics.md)
- **Engineer recovery** — define backup, restore, failover, rollback, continuity, emergency disable, recovery objectives, and authoritative verification, see [references/recovery-and-continuity.md](references/recovery-and-continuity.md)
- **Change progressively** — pin candidate/rollback versions, canary representative cohorts, monitor success/abort criteria, authorize, verify, and detect drift, see [references/recovery-and-continuity.md](references/recovery-and-continuity.md)
- **Operate with ownership** — prepare support, escalation, incident, communication, corrective action, learning, and retirement before failure, see [references/operations-and-learning.md](references/operations-and-learning.md)
- **Measure systems, not people** — combine quantitative and qualitative outcomes with purpose, aggregation, privacy, interpretation limits, and anti-gaming counter-metrics, see [references/operations-and-learning.md](references/operations-and-learning.md)

## Gotchas

- Availability without correctness, freshness, authorization, or safe failure can make an unreliable control look healthy.
- Retries amplify load and duplicate side effects unless bounded by deadlines, backoff, jitter, idempotency, and reconciliation.
- A successful API response is not recovery or rollback verification; re-read authoritative state and exercise the recovered outcome.
- Telemetry outage is itself a degraded state. It does not always block useful work, but it must never silently become evidence of health.
- High deployment frequency or low alert counts can be gamed. Balance delivery, reliability, security, user, and recovery outcomes and review incentives.

## Example

```text
Outcome: policy decisions correct and fresh for 99.95% of governed operations
Bounds: 2 s deadline · one jittered retry · idempotency key · bounded queue · cancel on expiry
Failure: stale mandatory policy fails closed; telemetry outage fails visible; advisory sink sheds load
Change: 5% canary · explicit abort thresholds · pinned rollback · independent emergency disable
Recovery: authoritative state reread + allow/deny probes + evidence-correlation verification
```

## Progressive Disclosure

- Read [references/service-objectives.md](references/service-objectives.md) - Load when defining SLOs, indicators, error budgets, capacity, dependency objectives, evidence freshness, or governance-effectiveness measures
- Read [references/failure-semantics.md](references/failure-semantics.md) - Load when designing timeouts, retries, idempotency, duplicate handling, concurrency, ordering, backpressure, cancellation, overload, or degraded modes
- Read [references/recovery-and-continuity.md](references/recovery-and-continuity.md) - Load when planning backup/restore, failover, continuity, canaries, rollback, emergency disable, drift, or recovery verification
- Read [references/operations-and-learning.md](references/operations-and-learning.md) - Load when defining support, ownership, incidents, corrective action, operational metrics, governed learning, or retirement
