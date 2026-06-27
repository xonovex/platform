# gherkin: Gherkin Keyword Reference

Gherkin is the plain-language notation for capturing an agreed example as an executable scenario. A feature file groups related scenarios, each a concrete example built from steps. The keywords exist so a single document is readable by business people AND runnable against the system — that dual readability is the whole reason to use the grammar rather than free prose.

## Primary keywords

- **Feature** — names and groups related scenarios; the opening line and free-text description say what capability the file is about. One feature per file.
- **Rule** — represents one business rule that the scenarios beneath it illustrate. Use it to cluster the examples that belong to the same rule so the document reads as "rule, then its examples".
- **Scenario** (alias **Example**) — a single concrete example, made of steps. This is the unit that runs.
- **Scenario Outline** (alias **Scenario Template**) — runs the same scenario once per row of an **Examples** (alias **Scenarios**) table, substituting `<placeholder>` values from the columns. Use it only when the SAME behaviour repeats over a small set of value combinations.
- **Background** — a set of `Given` steps run before EVERY scenario in the file, so shared preconditions are written once.

## Step keywords

- **Given** — known state: the preconditions that must hold before the action. Put the world into a knowable state; do not describe user interaction here.
- **When** — the single event or action under examination. Keep one essential `When` per scenario.
- **Then** — the expected, observable outcome. A `Then` is an ASSERTION: it compares an actual result against an expected one. It must describe something the business can observe, not an internal database row.
- **And** / **But** — continue the previous step type so several Givens or Thens read naturally. They take the meaning of the keyword above them; `But` is only cosmetic for a negative.

## Declarative over imperative

Write steps DECLARATIVELY — state intent and outcome, not the UI mechanics. Imperative steps ("click the #submit button", "type 4 into the qty field") couple the scenario to the interface, bloat the file, and obscure the rule. Declarative steps survive a redesign and keep the example about behaviour.

```gherkin
# BAD — imperative, UI-coupled, no rule visible
Scenario: Checkout
  Given I open "/cart"
  When I click "Add coupon"
  And I type "SAVE10" into "#code"
  And I click "Apply"
  And I click "Pay"
  Then I see ".total" contains "$45.00"
```

```gherkin
# GOOD — declarative, states the rule's behaviour
Feature: Order checkout

  Rule: A valid coupon reduces the order total by its percentage

    Background:
      Given a cart of items totalling 50.00 EUR

    Scenario: A 10 percent coupon is applied
      Given a valid coupon "SAVE10" worth 10 percent
      When the customer checks out
      Then the order total should be 45.00 EUR

    Scenario Outline: Coupons reduce the total by their percentage
      Given a valid coupon worth <percent> percent
      When the customer checks out
      Then the order total should be <total> EUR

      Examples:
        | percent | total |
        | 10      | 45.00 |
        | 25      | 37.50 |
```

## Keep it small

A scenario with many `When` steps is usually two scenarios; a `Then` that asserts five unrelated things is an Eager-style example — keep one rule per scenario and 2-5 scenarios per `Rule`. Exhaustive value coverage and edge cases belong in unit tests driven by **tdd-guide**, not in more outlines here. The vocabulary the steps are written in (the names for "order", "coupon", "taxi") is the ubiquitous language owned by **ddd-guide**; the step-definition glue code and any stubbing of external services is owned by **testing-guide**. Scenarios are first the OUTPUT of a discovery conversation (see discovery-three-amigos) and only second a regression suite.
