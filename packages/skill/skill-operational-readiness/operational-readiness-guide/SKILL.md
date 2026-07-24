---
name: operational-readiness-guide
description: "Use when assessing whether a service and its owning team can operate safely and sustainably in production before launch, transfer, or a material change. Triggers on operational or production readiness review, service ownership, SLOs, dependencies, capacity, observability, alerts, on-call, runbooks, playbooks, recovery objectives, backup restore, support model, operational risk, or readiness actions, even when the user doesn't say operational readiness."
---

# Operational Readiness

Assess the enduring ability to operate a service through normal work, change, failure,
and recovery without replacing release-candidate assessment or active incident
response.

## Essentials

- **Pin service scope and ownership** — identify service revision, users, critical
  functions, environments, data, dependencies, owners, support hours, and escalation
- **Define service objectives** — connect user journeys to measurable indicators,
  objectives, error-budget or tolerance policy, capacity assumptions, and compliance
  obligations
- **Verify observability and response** — ensure actionable signals, dashboards,
  alert ownership, runbooks, playbooks, on-call access, communications, and incident
  command paths
- **Exercise failure and recovery** — cover dependency, overload, deployment, data,
  region, credential, security, and operator failure with containment, degraded modes,
  backup, restore, and recovery evidence
- **Review sustainable operation** — assess routine procedures, change safety,
  maintenance, vulnerability and certificate lifecycle, support, staffing, training,
  toil, cost, and third-party commitments
- **Return owned readiness findings** — record ready, conditionally ready, not ready,
  or blocked with evidence, residual risk, action owner, target, and re-review trigger

## Gotchas

- A generic checklist misses risks learned from this service's incidents and near misses.
- A dashboard without user-centered objectives and an owner is not operational readiness.
- A backup is not recovery evidence until restoration and integrity are exercised.
- Runbooks must be usable by the actual responder with available production access.
- Operational readiness evidence feeds release readiness; it does not approve a release.

## Example

```text
Service: Identity verification v3, owned by Identity Platform, 24x7 critical journey.
Objective: 99.9% successful eligible verifications over 28 days; latency p95 < 4 s.
Finding: Provider failure alert exists, but no safe degraded mode or current escalation.
Recovery: Restore drill meets four-hour RTO; data-integrity verification is undocumented.
Recommendation: Conditionally ready after dependency playbook exercise and integrity check.
Re-review: Before general availability or after provider, SLO, or ownership change.
```

## Progressive Disclosure

- Read [references/objectives-and-ownership.md](references/objectives-and-ownership.md) - Load when defining service scope, users, ownership, SLOs, capacity, dependencies, obligations, support, or escalation
- Read [references/failure-and-recovery.md](references/failure-and-recovery.md) - Load when reviewing observability, alerts, runbooks, playbooks, failure modes, degraded operation, backups, restore, or exercises
- Read [references/readiness-review.md](references/readiness-review.md) - Load when running the review, tailoring questions, recording findings, assigning actions, reporting residual risk, or scheduling re-review
