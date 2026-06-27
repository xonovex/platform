# appendix: The Discovery Framing

The single most-misunderstood thing about BDD tooling: it is NOT a software-testing tool. It tests people's understanding of how yet-to-be-written software should behave. Hold that framing and every other decision — how many scenarios, what to stub, where TDD fits — follows from it. Lose it and the practice collapses into a slow, brittle UI-test suite.

## Why BDD exists

BDD grew out of confusion in TDD about where to start, what to test, and what to call tests. Renaming "test" to "behaviour" and starting example names with "should" made intent clearer and turned the question into "what should this system do?". The Given-When-Then template — given some initial context, when an event occurs, then ensure some outcomes — emerged as a shared language for the analysis itself, a way to phrase a requirement that a business person and a machine can both read.

## The three practices

BDD is three practices, each answering a different question:

- **Discovery** — what it COULD do: a structured conversation over concrete examples (see discovery-three-amigos).
- **Formulation** — what it SHOULD do: document the agreed examples in a human- and computer-readable medium (Gherkin, see gherkin-reference).
- **Automation** — what it ACTUALLY does: connect the examples to the system as guide-rails.

Automation is the THIRD practice, not the purpose. The outcome is documentation automatically checked against the system.

## The discovery workflow

1. Pop a story and run a discovery workshop (~20 minutes) with business and IT present.
2. Talk through concrete examples in plain language; thumbs-up/down splits a too-big story or sends an unclear one back.
3. A developer, with a tester, makes 2-5 of the examples concrete in Gherkin; the business confirms them.
4. The developer drives the CORE DOMAIN with ordinary TDD (**tdd-guide**) while external services, queues, and databases are STUBBED — no UI, no browser automation.
5. The result is FAR more unit tests than Gherkin scenarios.

## Two jobs, two tools

- **Cucumber / scenarios** = write the RIGHT code (build the thing the business asked for).
- **Unit tests** = write the code RIGHT (build it correctly, exhaustively, fast).

These are complementary, not competing. A handful of living-documentation scenarios sit on top of many unit tests; BDD does not replace unit testing.

```text
# BAD — scenarios driven end-to-end through the UI
Every scenario launches a browser, logs in, clicks through pages.
Suite takes 40 minutes, breaks on any layout change, and when it fails
nobody can tell whether the rule or the button moved.

# GOOD — scenarios drive the core domain, externals stubbed
The taxi-assignment scenario calls the ranking domain directly.
The location service and the dispatch queue are stubbed (mechanics owned
by testing-guide). The single living-documentation scenario passes in
milliseconds, sitting above dozens of TDD-driven unit tests for the
distance/rating edge cases.
```

The output of the whole loop is executable LIVING DOCUMENTATION (see specification-by-example) that prevents defects up front by pinning down the right behaviour before it is built — not a regression net bolted on afterwards.
