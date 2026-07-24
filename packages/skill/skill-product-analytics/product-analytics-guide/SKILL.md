---
name: product-analytics-guide
description: "Use when defining, validating, or interpreting quantitative product and service measurement for an outcome or decision. Triggers on metric trees, KPI definitions, funnels, cohorts, retention, adoption, HEART metrics, guardrails, instrumentation plans, event taxonomies, data quality, dashboards, experiment metrics, A/B analysis, causal claims, or product performance reviews, even when the user doesn't say product analytics."
---

# Product Analytics

Turn product questions into auditable measurement and proportionate decisions without
optimizing proxy metrics, confusing correlation with causation, or collecting data
without a justified purpose.

## Essentials

- **Start with the decision** — name the user or business outcome, decision, population,
  baseline, alternatives, time horizon, and action thresholds
- **Define a balanced measurement model** — connect goals to signals and metrics,
  include leading and lagging outcomes, segments, guardrails, and qualitative evidence
- **Specify metrics exactly** — define event or record sources, numerator, denominator,
  eligibility, exclusions, identity, attribution, window, aggregation, and owner
- **Design governed instrumentation** — version event contracts, minimize data,
  address consent and retention, validate environments and identity, and plan backfill
  or schema changes
- **Prove data quality before interpretation** — monitor completeness, validity,
  uniqueness, timeliness, consistency, bot or test traffic, missingness, and drift
- **Match analysis to the claim** — distinguish descriptive, diagnostic, predictive,
  and causal conclusions; predefine experiment hypotheses, assignment, metrics,
  guardrails, duration, and decision rules

## Gotchas

- A dashboard visualizes definitions; it does not make them correct.
- A rising engagement proxy can coexist with worse user outcomes or exclusion.
- Metric definitions must survive retries, cross-device identity, delayed events, and
  changes to eligibility.
- Statistical significance does not establish practical value, trustworthy execution,
  or an acceptable guardrail result.
- Product analytics informs discovery and prioritization; it does not replace user
  research or make roadmap decisions.

## Example

```text
Decision: Keep, revise, or remove the new recovery flow after four weeks.
Outcome: Eligible users complete recovery without assisted support.
Primary metric: Completed eligible recoveries / started eligible recoveries in 24 h.
Segments: Access need, device class, account age, and recovery route.
Guardrails: Fraud-confirmed rate, repeat attempts, support contacts, and p95 latency.
Quality: Reconcile starts and completions daily; exclude staff and synthetic traffic.
Rule: Keep only if completion improves materially and no guardrail crosses its limit.
```

## Progressive Disclosure

- Read [references/measurement-model.md](references/measurement-model.md) - Load when framing the decision, defining goals, signals, metrics, funnels, cohorts, guardrails, segments, baselines, or thresholds
- Read [references/instrumentation-and-quality.md](references/instrumentation-and-quality.md) - Load when specifying events, identity, attribution, governance, privacy, data contracts, validation, monitoring, or metric changes
- Read [references/experiments-and-decisions.md](references/experiments-and-decisions.md) - Load when interpreting product evidence, designing or analyzing controlled experiments, assessing causality, or recording a product recommendation
