---
name: release-readiness-guide
description: "Use when assessing whether an exact software release candidate is sufficiently evidenced and controlled for a proposed rollout. Triggers on release readiness, go or no-go evidence, candidate identity, artifact provenance, release criteria, rollout strategy, canary analysis, deployment verification, rollback or roll-forward, migration sequencing, change window, release communications, or residual release risk, even when the user doesn't say release readiness."
---

# Release Readiness

Integrate evidence and controls for one immutable release candidate into a descriptive
recommendation without changing the candidate, approving a protected gate, or
deploying it.

## Essentials

- **Pin the candidate** — identify source revision, build, artifacts and digests,
  provenance, dependencies, configuration, migrations, target environments, and change
  set
- **Verify evidence freshness** — map product, quality, security, accessibility,
  compliance, architecture, and operational evidence to the exact candidate or justify
  valid inheritance
- **Plan a controlled rollout** — define stages, populations, timing, concurrency,
  feature controls, migration order, compatibility, and blast-radius limits
- **Set observable decision rules** — name baseline, canary or stage signals,
  guardrails, evaluation windows, thresholds, ownership, and proceed/pause/abort rules
- **Prepare recovery and communication** — prove rollback or roll-forward,
  data recovery, operator access, runbooks, contacts, user and stakeholder messages,
  and support readiness
- **Return a bounded recommendation** — ready, conditionally ready, not ready, or
  blocked with evidence, exceptions, residual risk, approver, and verification plan

## Gotchas

- Evidence for a previous build is stale unless the unchanged scope is demonstrated.
- A passed pipeline says only that configured checks passed for the identified inputs.
- Rollback is not credible when data or contract changes make the prior version unsafe.
- A canary needs representative exposure, stable comparison, enough time, and explicit
  decision rules.
- The recommendation does not satisfy a provider approval or authorize deployment.

## Example

```text
Candidate: rc-18, source abc123, image digest sha256:..., provenance verified.
Evidence: Critical quality and security criteria pass on rc-18; accessibility evidence
          inherits only for unchanged screens; payment sandbox limitation remains.
Rollout: Internal traffic, then 5%, 25%, 100% with 30-minute minimum stages.
Guardrails: Error ratio +0.5%, p95 +100 ms, duplicate-charge signal >0; any breach aborts.
Recovery: Feature disable plus compatible rollback; migration restore drill linked.
Recommendation: Conditionally ready after on-call drill evidence is attached.
```

## Progressive Disclosure

- Read [references/candidate-and-evidence.md](references/candidate-and-evidence.md) - Load when identifying the candidate, artifacts, provenance, change set, dependencies, evidence freshness, criteria, or exceptions
- Read [references/rollout-and-recovery.md](references/rollout-and-recovery.md) - Load when designing staged rollout, canary analysis, migrations, guardrails, rollback, roll-forward, recovery, communications, or support
- Read [references/decision-and-verification.md](references/decision-and-verification.md) - Load when writing a readiness recommendation, recording residual risk, handing off to approvers, or defining post-deployment verification
