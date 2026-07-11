# specification-by-example: Specification by Example and Living Documentation

Turn agreed examples into the project's specification AND its test suite at once. Specification by Example, ATDD, Story-Test-Driven Development, and BDD are one practice differing in emphasis only — do not split them into rival methodologies.

## The seven process patterns

1. **Derive scope from goals** — start from the business goal, let it decide scope.
2. **Specify collaboratively** — write the spec together (three-amigos perspectives), not handed down.
3. **Illustrate using examples** — concrete realistic examples, not abstract prose.
4. **Refine the specification** — distil examples to precise, unambiguous specs; drop incidental detail, name the rule.
5. **Automate validation without changing the specification** — wire examples to run WITHOUT rewording the agreed spec to suit tooling.
6. **Validate frequently** — run continuously so document/system drift fails immediately.
7. **Evolve a living documentation system** — the validated specs are the single trusted current description of the system.

## Why "living"

When examples are both spec and automated check, "specification" and "test" become one artefact. Because it runs against the system every build, a stale statement fails — the documentation cannot silently rot.

Pattern 5 is the load-bearing discipline: if automation rewords "a 10% coupon reduces a 50.00 order to 45.00" into `assert discountEngine.apply(...) == 45`, the business reader can no longer recognise their rule and the doc is dead. Keep the agreed Gherkin wording; step definitions (owned by **testing-guide**) bind that unchanged text to the system.

Keep specs declarative (see gherkin-reference) and refine ruthlessly — a few illustrative examples per rule; exhaustive cases go to **tdd-guide**/**testing-guide**. The vocabulary is the ubiquitous language owned by **ddd-guide**; the scoping story by **user-stories-guide**; the WHY-this-tests-understanding framing in appendix-discovery.
