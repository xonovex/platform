# test-smells-and-fixtures: Reading a Bad Test, and Building Clean Fixtures

Naming a smell points at the fix. Most setup-half smells come from how the fixture is built, so clean fixture-creation prevents a whole class.

## The smell families

- **Obscure Test**: you cannot tell what it does or why: from too _much_ setup burying the one value that matters, or too _little_ (magic numbers, helper calls hiding what's exercised). Fix: move irrelevant setup into named helpers; make the value under test loud and local.
- **Mystery Guest**: a specific Obscure Test: it reads/asserts against an external resource not visible in the body (shared fixture file, seeded row, a record another test created). Fix: construct the data _inside_ the test (or via a Builder) so it's self-contained and Independent.
- **Eager Test**: verifies too much in one method; on failure you can't tell which behaviour broke. Fix: one test per logical outcome, each behaviour-named with a single reason to fail (see [test-structure-and-first.md](test-structure-and-first.md)).
- **Fragile Test**: breaks on changes unrelated to what it checks. Two common causes:
  - **Sensitive Equality**: asserting against a whole stringified/serialized object, so an unrelated field/format/order change reddens the bar. Fix: assert on the specific fields.
  - **Indirect Testing**: exercising the SUT through a distant collaborator, so a change anywhere on the path breaks it. Fix: test the unit directly. (Over-mocking is a third Fragile cause, see [state-vs-behaviour-and-what-to-mock.md](state-vs-behaviour-and-what-to-mock.md).)
- **Erratic Test**: flaky: different results with no code change, from uncontrolled non-determinism (wall clock, seeds, time zones, concurrency, network, order-dependence on leftover state). Violates FIRST Repeatable/Independent. Fix: inject a fixed clock and seed, double external seams, no shared mutable fixture.
- **Tautological Test**: green for the wrong reason; nothing it asserts can fail. The tell: you cannot name an input or change that would make it red. A common form is the **After-the-fact baseline**: a golden/characterization test generated from already-changed code; it prevents future drift but did not prove the change preserved behaviour. Fix: first name the concrete input/change that must fail, then capture golden baselines _before_ the change.

Smells breed each other: a bloated fixture makes setup Obscure → people lean on a shared external fixture → Mystery Guest _and_ Erratic. Fixing the root (usually the fixture and the unit's size) clears several at once.

## Fixture-creation: Test Data Builder vs Object Mother

- **Object Mother**: factory methods returning ready-made canonical objects (`anOrder()`, `aPaidOrder()`). Every new variation needs a new method, so the Mother grows unmanageable.
- **Test Data Builder**: a fluent per-field builder specifying **only the values this test cares about**, defaulting the rest: `anOrder().withTotal(1250).withStatus("paid").build()`.

Prefer the **Test Data Builder**: it scales with data variation and makes the relevant field loud. (They combine: a Builder with sensible defaults, exposed through a couple of Mother-style starting points.)

```ts
// BAD: Object Mother forces a new method per variation, hides the relevant field
const order = anOverdueOrderWithTwoLinesAndAExpiredCard();
// GOOD: Builder states only what this test cares about
const order = anOrder().withStatus("overdue").withCard(expired()).build();
```

## Don't use fixtures to hide a design smell

If building the object under test takes a twenty-call Builder or an entire object graph, the fixture is not the problem: the production design is. Fix the design (smaller aggregates, fewer collaborators); the coupling diagnosis is owned by **connascence-guide**.

## Cross-references

- Doubles a fixture wires up: [test-double-taxonomy.md](test-double-taxonomy.md). Why over-mocking makes a test Fragile: [state-vs-behaviour-and-what-to-mock.md](state-vs-behaviour-and-what-to-mock.md). FIRST qualities the Erratic smell violates: [test-structure-and-first.md](test-structure-and-first.md).
- The coupling vocabulary behind "Fragile because over-coupled" is owned by **connascence-guide**; a read-only smell audit that _detects_ missing/over-mocked tests is owned by **code-quality-guide**.
