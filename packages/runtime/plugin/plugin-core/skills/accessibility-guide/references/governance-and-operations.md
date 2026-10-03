# Governance and Operations

## Compose an accessibility profile

The profile selects the standard/version/level, scope, applicable criteria, platform support, required user journeys, evaluator classes, environments, assessor roles, independence, evidence freshness, release behavior, exception authority, monitoring, and reassessment triggers.

Map each mandatory requirement to an adequate enforcement point. Advisory design guidance may fail visibly without blocking; a mandatory release criterion needs a protected gate or accountable human control that cannot be bypassed by the same actor performing the work.

## Handle findings and exceptions

Every finding has affected users and journeys, exact revision, criterion, severity, evidence, owner, remediation, target date, regression scope, and verification state.

An exception records exact scope, owner, authorized approver, rationale, affected users, compensating access, start, expiry, evidence, remediation owner/date, review cadence, and reassessment. It cannot silently become the default or claim conformance. Expired, ownerless, evidence-free, or overbroad exceptions fail visibly for mandatory requirements.

## Observe production without surveilling users

Monitor accessible-path availability, support contacts, regressions, unresolved finding age, exception age, assessment coverage/freshness, remediation lead time, and repeated component failures. Aggregate to product and journey outcomes; do not rank individual contributors or collect disability information without a defined purpose, consent or other valid authority, minimization, access, retention, and deletion controls.

Telemetry can identify a possible barrier but does not replace direct accessibility evaluation. Keep raw user content and assistive-technology details out of telemetry by default.

## Operate change and incidents

Treat a critical inaccessible journey, loss of alternative access, broken authentication, or inaccessible emergency communication as an operational incident according to impact. Preserve evidence, restore access, communicate through accessible channels, verify recovery with the affected path, and create corrective actions.

Reassess after material design, content, component, dependency, platform, policy, or support changes. Retire obsolete rules, tools, waivers, and compatibility assumptions only after proving replacement coverage, migrating evidence, disabling old enforcement, and verifying no user path was lost.
