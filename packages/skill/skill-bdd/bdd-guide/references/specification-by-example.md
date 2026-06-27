# specification-by-example: Specification by Example and Living Documentation

Specification by Example is the collaborative method that turns agreed examples into the project's specification AND its test suite at once. The same idea travels under several names — Acceptance-Test-Driven Development, Story-Test-Driven Development, and BDD — and the differences are emphasis only. Treat them as one practice; do not split it into rival methodologies.

## The seven process patterns

The method is built on seven key process patterns that feed each other:

1. **Derive scope from goals** — start from the business goal, not a pre-written feature list; let the goal decide what is in scope.
2. **Specify collaboratively** — write the specification together (the three-amigos perspectives, see discovery-three-amigos), not handed down by one author.
3. **Illustrate using examples** — express each requirement as concrete, realistic examples rather than abstract prose.
4. **Refine the specification** — distil the examples into precise, unambiguous specifications; remove incidental detail and name the rule.
5. **Automate validation without changing the specification** — wire the examples to the system so they run, WITHOUT rewording the agreed specification to suit the tooling.
6. **Validate frequently** — run the executable specifications continuously so drift between document and system is caught immediately.
7. **Evolve a living documentation system** — maintain the validated specifications as the single, trusted, current description of what the system does.

## Specs and tests become one artefact

When examples are both the specification and the automated checks, "specification" and "test" stop being separate documents. The payoff is documentation that cannot silently rot: because it is executed against the system on every run, a stale statement fails the build. This is what makes the documentation LIVING rather than a wiki page that nobody recompiles against — code reflects the documentation, and the documentation reflects the team's shared understanding.

## Keeping living documentation trustworthy

Living documentation is only worth maintaining if it stays trustworthy, which requires discipline:

- Keep specifications declarative and about behaviour, so they describe intent, not interface mechanics (see gherkin-reference).
- Automate validation without distorting the agreed wording — pattern 5 — so the document a business reader signs off is the document that runs.
- Validate frequently so a divergence between document and system surfaces as a failure, not as quiet decay.
- Refine ruthlessly: a few illustrative examples per rule, not an exhaustive dump; exhaustive cases go to **tdd-guide**/**testing-guide** unit tests.

```text
# BAD — automation rewords the spec until it passes
The agreed example said "a 10% coupon reduces a 50.00 order to 45.00".
To make it run, someone rewrote it as "assert discountEngine.apply(...) == 45".
The business reader can no longer recognise their own rule; the doc is dead.

# GOOD — automation honours the agreed wording (pattern 5)
The Gherkin keeps "a valid coupon worth 10 percent ... the order total
should be 45.00 EUR". Step definitions (owned by testing-guide) bind that
unchanged text to the system, so the running spec and the signed-off spec
are the same words, validated on every build.
```

The vocabulary these specifications are written in is the ubiquitous language owned by **ddd-guide**; the story that scopes a specification is owned by **user-stories-guide**; the underlying framing of WHY this tests understanding rather than software is in appendix-discovery.
