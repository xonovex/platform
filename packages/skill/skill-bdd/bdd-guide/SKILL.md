---
name: bdd-guide
description: "Use when driving software from concrete agreed examples of behaviour — running three-amigos discovery and example mapping before coding, formulating Given-When-Then scenarios in feature files, and treating them as living documentation that tests the team their shared understanding. ATDD and BDD are one practice here. Triggers on BDD, ATDD, Gherkin, Given-When-Then, Cucumber, feature file, scenario / scenario outline, three amigos, example mapping, specification by example, living documentation, acceptance-test-driven — even when the user doesn not say 'BDD'. Skip the inner red-green-refactor loop that drives the domain logic (see **tdd-guide**), step-implementation / stubbing mechanics and test doubles (see **testing-guide**), the story container and INVEST (see **user-stories-guide**), and the ubiquitous language the scenarios are written in (see **ddd-guide**)."
---

# Behaviour- and Acceptance-Driven Development

Drive software from concrete, agreed examples of behaviour. Discover them in conversation, formulate them as Given-When-Then scenarios, and keep the executable scenarios as living documentation. The acceptance layer answers "are we building the right thing?" — the inner domain logic is still driven test-first by **tdd-guide**. ATDD and BDD are two names for this one practice.

## Essentials

- **Discover before you formulate** - run a three-amigos conversation over real examples first, see [references/discovery-three-amigos.md](references/discovery-three-amigos.md)
- **Map examples to size and split the story** - yellow/blue/green/red cards expose scope and gaps, see [references/discovery-three-amigos.md](references/discovery-three-amigos.md)
- **Write scenarios in Gherkin Given-When-Then** - Feature, Scenario, Background, Outline; declarative steps, see [references/gherkin-reference.md](references/gherkin-reference.md)
- **Keep 2-5 scenarios per rule, push the rest down** - illustrate the rule, then drive exhaustive cases with **tdd-guide** unit tests
- **Treat scenarios as living documentation** - specs and tests become one trustworthy artefact, see [references/specification-by-example.md](references/specification-by-example.md)

## Gotchas

- Gherkin/Cucumber is a **collaboration** tool that tests the team's shared UNDERSTANDING of unwritten software, not a testing tool — automation is the third practice, never the purpose, see [references/appendix-discovery.md](references/appendix-discovery.md).
- BDD does not replace unit testing: drive the core domain with **tdd-guide** and write FAR more unit tests than scenarios; Cucumber writes the right code, unit tests write the code right.
- "Three amigos" means three PERSPECTIVES (business / development / testing), not exactly three people — invite more; the conversation is the point.
- Driving scenarios end-to-end through the UI/browser is slow, volatile, and fails to localise the bug — stub external services/queues/DBs and drive the core domain directly.
- Given-When-Then is the recommended template, not the goal; the essential thing is concrete agreed examples, so do not over-engineer the grammar with deep nesting or imperative "click button X" steps.
- ATDD and BDD are not rival methodologies — same example-driven, acceptance-level idea; do not split the practice in two.

## Example

```gherkin
Feature: Assign the best nearby taxi

  Rule: A closer taxi with a higher rating wins

    Scenario: Close taxi with higher rating is preferred
      Given taxi A with rating 0.8 is 1400m from the customer
      And taxi B with rating 0.9 is 1500m from the customer
      When the customer requests a taxi
      Then taxi B should be assigned
```

## Progressive Disclosure

- Read [references/gherkin-reference.md](references/gherkin-reference.md) - Load when writing or reviewing feature files: keyword meanings, file structure, declarative vs imperative steps.
- Read [references/discovery-three-amigos.md](references/discovery-three-amigos.md) - Load when running discovery: the three perspectives, the workshop, and example-mapping card colours.
- Read [references/specification-by-example.md](references/specification-by-example.md) - Load when establishing the process: the seven patterns, specs-as-tests, and trustworthy living documentation.
- Read [references/appendix-discovery.md](references/appendix-discovery.md) - Load for the underlying framing: testing understanding, scenarios-per-rule, stubbing externals, right-code vs code-right.
