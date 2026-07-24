# Tradeoff Analysis

## Trace Architectural Approaches

For each high-priority scenario, identify the decisions, structures, allocation of
responsibility, data paths, dependencies, failure controls, and operating assumptions
that affect the response. Test the claimed causal link using the strongest economical
evidence available:

- reasoning against documented constraints;
- analytical or capacity models;
- executable spikes or prototypes;
- benchmark, failure-injection, or operational evidence;
- implementation conformance checks.

Label evidence freshness and limits. Absence of contradictory evidence is not proof.

## Analysis Outputs

- **Risk** — a decision or uncertainty may prevent a scenario from being met.
- **Non-risk** — available evidence supports the scenario under stated conditions.
- **Sensitivity point** — a small change in a parameter or decision materially changes
  a quality response.
- **Tradeoff point** — one decision affects multiple qualities in competing ways.

Compare alternatives against the same scenarios and evidence standard. Record cost,
complexity, reversibility, migration, and operational consequences without collapsing
them into a context-free score.

Use an ADR capability to record the selected decision after evaluation. Use specialist
architecture or platform skills to design or implement the alternatives.
