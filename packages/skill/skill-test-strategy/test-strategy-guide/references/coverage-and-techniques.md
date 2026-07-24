# Coverage and Techniques

## Allocate Coverage

Choose the cheapest level that can expose the failure with sufficient realism, then
add higher-level evidence for interactions and user-visible behavior. Avoid duplicating
identical assertions at every level.

Select only relevant test types:

- functional examples, state, boundary, decision, combinatorial, and property tests;
- component, contract, integration, system, acceptance, and end-to-end checks;
- exploratory, usability, accessibility, compatibility, localization, and content;
- performance, capacity, concurrency, soak, resilience, recovery, and migration;
- security verification derived from threats and versioned requirements;
- deployment, observability, canary, rollback, and post-release verification.

Use the focused domain skill for test design or execution. This skill owns the overall
allocation and gap analysis, not every specialist technique.

## Execution Design

Define representative environments, configuration differences, data sources,
privacy controls, reset and cleanup, dependencies or simulations, fault controls,
observable signals, and expected oracles. Identify what cannot be represented outside
production.

Automate stable, valuable, repeatable checks at the appropriate layer. Preserve
exploration for uncertainty and human judgment. Define ownership, feedback time,
quarantine criteria, and repair expectations for flaky checks.
