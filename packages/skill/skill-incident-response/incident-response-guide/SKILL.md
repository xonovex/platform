---
name: incident-response-guide
description: "Use when declaring, coordinating, mitigating, communicating, recovering from, or learning from an active production or security incident. Triggers on incident commander, severity, blast radius, outage, degradation, containment, mitigation, responder roles, status updates, incident timeline, operational handoff, recovery verification, rollback, post-incident review, or repeated operational failure, even when the user doesn't say incident response."
---

# Incident Response

Reduce user harm through clear command, controlled mitigation, reliable communication,
and verified recovery while preserving evidence and avoiding speculative changes.

## Essentials

- **Declare early from observed impact** — name the incident, severity, affected users,
  services, regions, start estimate, current symptoms, and uncertainty
- **Assign explicit response roles** — one incident commander coordinates; operations
  changes the system; communications updates stakeholders; specialists advise
- **Mitigate before diagnosing perfectly** — choose the safest reversible action that
  reduces impact while preserving evidence needed for recovery and investigation
- **Keep one effect log** — record timestamp, actor, hypothesis, command or provider
  action, target, revision, expected result, observed result, and rollback
- **Communicate on a cadence** — state impact, scope, actions, current status,
  uncertainty, next update, and support guidance without unsupported causes or ETAs
- **Verify recovery and learn** — confirm user journeys and telemetry, watch for
  recurrence, preserve follow-up ownership, and analyze system conditions without blame

## Gotchas

- Waiting for root cause before declaring or mitigating extends harm.
- More responders without clear roles increase conflicting changes and communication
  load.
- “Metrics are green” is not recovery until affected user journeys and dependencies are
  verified.
- Rolling back can increase harm when data or external side effects are not backward
  compatible; preview the exact recovery effect.
- A post-incident review that stops at individual error misses the conditions that made
  the failure likely or hard to detect.

## Example

```text
Incident: INC-1842 · checkout failures in EU · severity 1
Impact: 62% payment attempts fail; no evidence of duplicate capture.
Roles: incident commander / operations lead / communications lead / payment specialist
Mitigation: disable new routing path at config revision 81 after preview and authority.
Verify: completion rate, provider events, order/payment consistency, synthetic journey.
Next update: 15 minutes; cause remains under investigation.
```

## Progressive Disclosure

- Read [references/declare-and-coordinate.md](references/declare-and-coordinate.md) - Load when declaring an incident, setting severity, assigning roles, scoping impact, or handing command to another responder
- Read [references/mitigate-and-recover.md](references/mitigate-and-recover.md) - Load when selecting or applying mitigation, containment, rollback, recovery, or verification actions
- Read [references/communication-and-learning.md](references/communication-and-learning.md) - Load when writing status updates, maintaining the timeline, closing an incident, or producing follow-up learning
