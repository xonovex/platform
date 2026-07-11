# appendix: The Discovery Framing

BDD tooling is NOT a software-testing tool — it tests people's understanding of how yet-to-be-written software should behave. Hold that framing and every other decision (how many scenarios, what to stub, where TDD fits) follows; lose it and the practice collapses into a slow, brittle UI-test suite.

## The three practices

- **Discovery** — what it COULD do: a structured conversation over concrete examples (see discovery-three-amigos).
- **Formulation** — what it SHOULD do: document the agreed examples in Gherkin (see gherkin-reference).
- **Automation** — what it ACTUALLY does: connect the examples to the system as guide-rails. Automation is the THIRD practice, not the purpose.

## Two jobs, two tools

- **Cucumber / scenarios** = write the RIGHT code (build what the business asked for).
- **Unit tests** = write the code RIGHT (correctly, exhaustively, fast).

Complementary, not competing: a handful of living-documentation scenarios sit on top of many unit tests. BDD does not replace unit testing.

## Drive the core, stub the edges

The developer makes 2-5 agreed examples concrete in Gherkin, the business confirms them, then drives the CORE DOMAIN with ordinary TDD (**tdd-guide**) while external services, queues, and databases are STUBBED (mechanics owned by **testing-guide**) — no UI, no browser automation. End-to-end-through-the-UI scenarios take tens of minutes, break on any layout change, and can't localise the failure. The result is FAR more unit tests than scenarios, and executable LIVING DOCUMENTATION (see specification-by-example) that pins down the right behaviour before it is built.
