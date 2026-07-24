---
name: architecture-evaluation-guide
description: "Use when evaluating whether a proposed or existing software architecture can satisfy its business drivers and competing quality goals before or during delivery. Triggers on architecture review, quality-attribute scenarios, ATAM, utility trees, architecture risks, sensitivity points, tradeoffs, fitness evidence, architectural alternatives, or an evaluation recommendation, even when the user doesn't say architecture evaluation."
---

# Architecture Evaluation

Evaluate architecture fitness through concrete scenarios and traceable evidence rather
than preference, diagram style, or an unweighted checklist.

## Essentials

- **Pin the evaluation subject** — identify the architecture revision, scope, system
  context, dependencies, assumptions, constraints, and available evidence
- **Elicit business drivers** — connect stakeholder outcomes and constraints to a
  prioritized set of product and system quality goals
- **Make qualities testable** — express each important quality as source, stimulus,
  environment, artifact, response, and measurable response threshold
- **Trace decisions to scenarios** — inspect how architectural approaches support or
  threaten each scenario, using models, prototypes, measurements, or implementation
  evidence where needed
- **Expose interactions** — record risks, non-risks, sensitivity points, and tradeoff
  points instead of declaring an architecture simply good or bad
- **Return bounded findings** — provide severity, evidence, affected scenarios,
  uncertainty, mitigation or experiment, owner, and re-evaluation boundary

## Gotchas

- A quality label such as scalable, secure, or maintainable is not an evaluable
  requirement.
- A familiar pattern is not evidence that its consequences fit this system.
- Architecture evaluation can start with a concept, but confidence must match the
  evidence available at that revision.
- One stakeholder group cannot supply all business drivers or operational scenarios.
- The evaluation informs an ADR or delivery decision; it does not make or enforce it.

## Example

```text
Subject: Checkout architecture at decision revision adr-42/draft-3.
Driver: Preserve conversion while isolating payment-provider failures.
Scenario: During a provider timeout burst of 20%, unaffected payments retain p95 under
          800 ms and failed attempts are safely retryable without duplicate charge.
Approach: Synchronous adapter plus idempotency store and bounded circuit breaker.
Finding: Retry safety is supported; latency isolation lacks load evidence.
Tradeoff: Shorter timeout limits queue growth but may reject recoverable requests.
Next evidence: Failure-injection test across the stated burst and latency thresholds.
```

## Progressive Disclosure

- Read [references/drivers-and-scenarios.md](references/drivers-and-scenarios.md) - Load when defining scope, stakeholders, business drivers, quality goals, utility trees, or measurable scenarios
- Read [references/tradeoff-analysis.md](references/tradeoff-analysis.md) - Load when tracing architectural approaches, comparing alternatives, modeling qualities, or identifying risks and tradeoffs
- Read [references/findings-and-follow-up.md](references/findings-and-follow-up.md) - Load when recording findings, confidence, mitigations, experiments, decisions, or re-evaluation boundaries
