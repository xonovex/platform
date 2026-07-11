# gherkin: Gherkin Keyword Reference

The grammar exists so ONE document is both readable by business people and runnable against the system.

## Keywords and non-obvious rules

- **Feature** — one per file; the free-text description names the capability.
- **Rule** — clusters the scenarios that illustrate one business rule; document reads "rule, then its examples".
- **Scenario** (alias **Example**) — one concrete example; the unit that runs.
- **Scenario Outline** (alias **Scenario Template**) — runs once per row of an **Examples** (alias **Scenarios**) table, substituting `<placeholder>` columns. Use only when the SAME behaviour repeats over a few value combinations.
- **Background** — `Given` steps run before EVERY scenario in the file (shared preconditions written once).
- **Given / When / Then** — known state / the single event under examination / an observable outcome assertion. Keep one essential `When`; `Then` must be business-observable, not an internal DB row.
- **And / But** — take the meaning of the keyword above; `But` is purely cosmetic.

## Declarative over imperative

State intent and outcome, not UI mechanics — declarative steps survive a redesign and keep the example about the rule.

```gherkin
# GOOD — declarative
Feature: Order checkout

  Rule: A valid coupon reduces the order total by its percentage

    Background:
      Given a cart of items totalling 50.00 EUR

    Scenario Outline: Coupons reduce the total by their percentage
      Given a valid coupon worth <percent> percent
      When the customer checks out
      Then the order total should be <total> EUR

      Examples:
        | percent | total |
        | 10      | 45.00 |
        | 25      | 37.50 |
```

Imperative steps (`When I click "#submit"`, `type "SAVE10" into "#code"`) couple the file to the interface and hide the rule.

## Ownership

Keep one rule per scenario, 2-5 scenarios per `Rule`; drive exhaustive value/edge coverage into **tdd-guide** unit tests, not more outlines. The vocabulary ("order", "coupon", "taxi") is the ubiquitous language owned by **ddd-guide**; step-definition glue and external stubbing is owned by **testing-guide**. Scenarios are first the OUTPUT of a discovery conversation (see discovery-three-amigos), second a regression suite.

- Feed each step only the data the rule needs; incidental or hardcoded values obscure the behaviour under examination.
- Keep steps to single digits per scenario — more signals a scenario doing too much.
- The `.feature` suite runs every build in CI as living regression, so any drift between spec and system fails immediately.
