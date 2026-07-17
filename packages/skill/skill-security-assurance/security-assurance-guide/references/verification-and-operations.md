# Verification and Security Operations

## Build layered assurance evidence

For each claim, identify the least adaptive authoritative source and an independent challenge where impact warrants it. Combine configuration inspection, exact-revision tests, intended allow/deny probes, scanners, provenance verification, access-control tests, human review, provider evidence, and operational exercises.

Record subject, policy/control/evaluator/environment versions, inputs, origin, actor/independence, time, coverage, limitations, findings, conflicts, and native references. A summary dashboard never replaces the underlying evidence.

## Exercise failure and adversarial paths

Test malformed and hostile input, direct and indirect injection, privilege expansion, secret and sensitive-data canaries, missing enforcement, stale versions, dependency and policy-provider outage, telemetry loss, duplicate and concurrent execution, ordering, timeout, cancellation, partial changes, rollback failure, exception abuse, unauthorized emergency exceptions, and fresh-context evidence reconstruction.

A test proves something only if a concrete bad case makes it fail. Include known-denied and known-failing fixtures; avoid baselines generated only after the implementation change.

## Handle exceptions and emergency exceptions

An exception records exact scope, control, owner, authorized approver, rationale, compensating controls, affected subjects, evidence, start, expiry, review, remediation, and revocation. An emergency exception additionally records emergency reason, explicit invocation, authoritative access evidence, notification, containment, time-limited access, revocation, and post-event review.

Expired, ownerless, unreviewed, evidence-free, overbroad, or repeatedly renewed records fail closed for mandatory controls. Review patterns of use for systemic problems without turning individual usage into surveillance.

## Respond to control and security incidents

Prepare ownership, severity, communication, evidence protection, containment, eradication, recovery, disclosure/notification decision, external coordination, and continuity before an incident. During response, preserve exact affected versions, decisions, enforcement, provider evidence, actions, authority, and timeline.

Verify recovery and control restoration independently. Create corrective actions with owners, target versions, effectiveness measures, evidence, expiry, and closure authority. Capture candidate lessons, but promote policy or module changes only through reviewed, canaried where executable, reversible change control.

## Operate assurance sustainably

Measure system-level control coverage, evidence freshness, false positives, bypass attempts, exception age, vulnerability age/exposure, rollback and recovery success, incident outcomes, drift, user friction, and unresolved gaps. Define purpose, audience, aggregation, access, retention, interpretation limits, and counter-metrics. Never rank individuals or auto-punish from security telemetry.
