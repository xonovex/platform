---
name: product-discovery-guide
description: "Use when framing and testing whether a product opportunity is worth pursuing before delivery planning. Triggers on solution-led feature requests, problem statements, desired outcomes, product hypotheses, assumptions, discovery scope, opportunity comparison, success measures, experiment design, or evidence-based proceed/change/stop recommendations, even when the user doesn't say product discovery."
---

# Product Discovery

Turn an uncertain opportunity into a bounded evidence-backed recommendation without
starting delivery or treating stakeholder confidence as user evidence.

## Essentials

- **Reframe solutions as problems** — identify affected users, desired progress,
  current behavior, pain, scope, and exclusions
- **Separate facts from assumptions** — record evidence origin, freshness, confidence,
  contradictions, and the riskiest unknowns
- **Define measurable outcomes** — pin a baseline, target, population, measurement
  window, guardrails, and decision threshold
- **Test the riskiest assumption first** — choose the smallest ethical research,
  prototype, data, or feasibility check that can change the decision
- **Compare interventions consistently** — assess expected user value, evidence,
  feasibility, cost, risk, reversibility, and whole-journey fit
- **Return a recommendation, not approval** — proceed, change, pause, or stop with
  rationale, unresolved uncertainty, and the next evidence boundary

## Gotchas

- A requested feature is a proposed intervention, not proof of a problem.
- Votes, stakeholder rank, and roadmap position are preference evidence, not user or
  outcome evidence.
- A metric without baseline, population, window, and guardrails invites selective
  interpretation.
- Discovery code and prototypes test assumptions; they are not production foundations
  unless separately planned and validated.
- Proceed means further investment is supported by current evidence; it does not
  approve funding, delivery, or release.

## Example

```text
Opportunity: Customers abandon identity verification before completion.
Evidence: 18% funnel loss at document capture; support contacts confirm confusion.
Unknown: Whether instructions or camera constraints dominate the failure.
Outcome: Reduce eligible-user abandonment to 10% in 30 days without increasing fraud.
Next test: Observe six representative attempts, including constrained devices, then
           compare instruction and capture prototypes against the same tasks.
Decision boundary: Continue only if one intervention improves completion without
                   violating the fraud guardrail.
```

## Progressive Disclosure

- Read [references/opportunity-framing.md](references/opportunity-framing.md) - Load when reframing a request, defining the problem, users, journey, scope, or outcome
- Read [references/experiments-and-evidence.md](references/experiments-and-evidence.md) - Load when inventorying assumptions, choosing a discovery method, or evaluating evidence quality
- Read [references/prioritization-and-handoff.md](references/prioritization-and-handoff.md) - Load when comparing opportunities, recording a recommendation, or handing discovery into product definition
