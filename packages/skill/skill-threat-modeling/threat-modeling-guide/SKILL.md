---
name: threat-modeling-guide
description: "Use when creating, reviewing, or updating a software threat model for a feature, data flow, service, integration, architecture, or exact system revision. Triggers on assets, security objectives, data-flow diagrams, trust boundaries, entry points, attacker goals, abuse cases, STRIDE, attack trees, threat enumeration, risk prioritization, mitigations, residual risk, security requirements, or verifying that a design change addresses identified threats, even when the user doesn't say threat modeling."
---

# Threat Modeling

Make security design risk explicit and actionable without substituting a checklist,
scanner, or generic assurance claim for system-specific reasoning.

## Essentials

- **Pin the model scope** — identify the exact system revision, scenarios, assets,
  security objectives, users, dependencies, environments, assumptions, and exclusions
- **Model how the system works** — show processes, stores, external entities, data
  flows, protocols, identities, privileges, entry points, and trust boundaries
- **Ask what can go wrong** — derive credible abuse and failure paths from the model,
  attacker capabilities, misuse cases, and an explicit enumeration method
- **Prioritize with rationale** — assess consequence, preconditions, reachability,
  existing controls, detectability, uncertainty, and affected security objectives
- **Make mitigations verifiable** — assign owner, design change, prevention or
  detection intent, affected threat, validation evidence, due boundary, and fallback
- **Reassess exact changes** — verify mitigations and update threats, assumptions,
  residual risk, and stale evidence whenever the modeled system changes materially

## Gotchas

- A diagram with no assets, trust boundaries, threats, or decisions is architecture
  documentation, not a threat model.
- A generated threat list detached from actual flows produces noise and false coverage.
- STRIDE is an enumeration aid, not a risk rating or proof that every threat was found.
- Encrypt, validate, log, and use least privilege are design directions until their
  exact target, mechanism, owner, and verification are defined.
- Recording residual risk does not authorize acceptance; the accountable external
  authority owns that effect.

## Example

```text
Scope: checkout-api@91c2, payment callback and order-finalization flow.
Asset/objective: one authorized charge maps to one durable order.
Boundary: provider callback crosses the public edge into a privileged finalizer.
Threat: replayed or reordered callbacks create duplicate fulfillment.
Mitigation: verify signed event identity, persist idempotency key before effects,
            enforce order-state transition, alert on conflicting replays.
Verify: replay, reorder, delay, duplicate, invalid-signature, and outage scenarios.
Residual: provider compromise remains outside application control and needs response.
```

## Progressive Disclosure

- Read [references/scope-and-model.md](references/scope-and-model.md) - Load when defining scope, assets, objectives, data flows, trust boundaries, identities, or assumptions
- Read [references/identify-and-prioritize.md](references/identify-and-prioritize.md) - Load when enumerating abuse paths, applying STRIDE or another method, or prioritizing threats
- Read [references/mitigate-and-verify.md](references/mitigate-and-verify.md) - Load when designing mitigations, deriving security requirements and tests, recording residual risk, or updating a model
