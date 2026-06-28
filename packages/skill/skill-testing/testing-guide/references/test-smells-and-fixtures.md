# test-smells-and-fixtures: Reading a Bad Test, and Building Clean Fixtures

A test that hurts to read, breaks on unrelated changes, or fails at random is reporting a smell. Naming the smell points at the fix. Most smells in the setup half of a test come from how the fixture is built — so clean fixture-creation patterns prevent a whole class of them.

## Contents

- The smell families
- Obscure Test and Mystery Guest
- Eager Test
- Fragile Test: Sensitive Equality and Indirect Testing
- Erratic Test
- How smells cause other smells
- Fixture-creation: Test Data Builder vs Object Mother
- Don't use fixtures to hide a design smell
- Cross-references

## The smell families

Test smells span what a test _does_ (behaviour smells), how its _code_ reads (code smells), and how the _suite_ behaves over time (project smells). The ones you meet most while writing or reviewing a single test:

- **Obscure Test** — you cannot tell what it does or why.
- **Mystery Guest** — it depends on something you can't see in the test.
- **Eager Test** — it verifies too much at once.
- **Fragile Test** — it breaks on changes unrelated to what it checks.
- **Erratic Test** — it passes or fails non-deterministically.
- **Tautological Test** — it passes no matter what the code does; nothing in it can fail.

## Obscure Test and Mystery Guest

An **Obscure Test** is hard to understand — from too _much_ information (pages of irrelevant setup burying the one value that matters) or too _little_ (magic numbers and helper calls that hide what is actually being exercised). Fix it by moving irrelevant setup into named helpers and making the value under test loud and local.

A **Mystery Guest** is a specific Obscure Test: the test reads or asserts against an external resource not visible in the test body — a shared fixture file, a seeded database row, a record another test created. The reader can't see the cause of the result. Fix it by constructing the data the test needs _inside_ the test (or via a Builder), so the test is self-contained and Independent.

## Eager Test

An **Eager Test** verifies too much in one method — several behaviours, or one behaviour plus a handful of incidental checks. When it fails you can't tell which behaviour broke, and the test name can't describe a single rule. Fix it by splitting into one test per logical outcome, each with its own behaviour-naming and a single reason to fail (see [test-structure-and-first.md](test-structure-and-first.md)).

## Fragile Test: Sensitive Equality and Indirect Testing

A **Fragile Test** breaks when you change something it wasn't meant to be checking. Two common causes:

- **Sensitive Equality** — asserting against a whole stringified or fully-serialized object, so an unrelated field, formatting, or ordering change reddens the bar. Fix it by asserting on the specific fields the test is about, not the entire blob.
- **Indirect Testing** — exercising the SUT through a distant collaborator instead of directly, so a change anywhere on the long path breaks the test. Fix it by testing the unit directly. (Over-mocking is a third Fragile cause — asserting on internal calls; see [state-vs-behaviour-and-what-to-mock.md](state-vs-behaviour-and-what-to-mock.md).)

A Fragile Test that breaks on every honest refactor is often reporting coupling in the production code, not a bad assertion — take that diagnosis to **connascence-guide**.

## Erratic Test

An **Erratic Test** gives different results across runs with no code change — a flaky test. Causes are non-determinism the test failed to control: the wall clock, random seeds, time zones, uncontrolled concurrency, network, or order-dependence on another test's leftover state. It violates FIRST's Repeatable and Independent. Fix it by injecting a fixed clock and seed, doubling external seams, and ensuring no shared mutable fixture leaks between tests.

## Tautological Test

## Tautological Test

A **Tautological Test** is green for the wrong reason: nothing it asserts can fail, so it certifies nothing. The tell: you cannot name an input or code change that would make it red. It usually appears in two forms:

- **Scope mismatch**: the assertion is narrower than the rule. For example, an architecture test checks only direct imports while the forbidden dependency is reached transitively.
- **After-the-fact baseline**: a golden or characterization test is generated from already-changed code. It may prevent future drift, but it did not prove the change preserved behavior.

Fix it by first naming the concrete input or change that must fail the test, then aligning the assertion with the rule: check transitive reachability when the rule is transitive; capture golden baselines before the change, or verify the change with pre-existing assertions.

## How smells cause other smells

Smells are not independent; one breeds another. A bloated, hard-to-build fixture makes the setup Obscure; to dodge the bloat people lean on a shared external fixture, which becomes a Mystery Guest _and_ an Erratic/order-dependent test. An Eager Test that checks many things tends to assert against a whole object (Sensitive Equality), which makes it Fragile. So fixing the root — usually the fixture and the unit's size — clears several smells at once. Treat the first smell you spot as a thread to pull, not the whole problem.

## Fixture-creation: Test Data Builder vs Object Mother

Two patterns build the objects a test needs:

- **Object Mother** — a set of factory methods returning ready-made canonical objects (`anOrder()`, `aPaidOrder()`, `anOverdueAccount()`). Convenient at first, but every new data variation needs a new method, so the Mother grows unmanageable and takes on more than one responsibility.
- **Test Data Builder** — a fluent, per-field builder that creates an object specifying **only the values this test cares about**, defaulting the rest: `anOrder().withTotal(1250).withStatus("paid").build()`. Each test states exactly what matters to it and nothing else, which kills Obscure setup and keeps tests Independent.

Prefer the **Test Data Builder**. It scales with data variation where the Object Mother does not, and it makes the relevant field loud and the irrelevant fields quiet. (The two combine well: a Builder seeded with sensible defaults, exposed through a couple of Mother-style starting points.)

```
// BAD — Object Mother forces a new method per variation, hides the relevant field
const order = anOverdueOrderWithTwoLinesAndAExpiredCard();   // what is this test about?

// GOOD — Builder states only what this test cares about
const order = anOrder().withStatus("overdue").withCard(expired()).build();
```

## Don't use fixtures to hide a design smell

If building the object under test takes a Builder twenty calls long, or forces you to construct an entire object graph to test one rule, the fixture is not the problem — the production design is. An aggregate that is too large or too coupled to construct cheaply will make every test that touches it Obscure. Don't paper over it with an ever-cleverer Builder or Mother; fix the design (smaller aggregates, fewer collaborators) and the fixture shrinks with it. The coupling diagnosis is owned by **connascence-guide**.

## Cross-references

- The doubles a fixture wires up are in [test-double-taxonomy.md](test-double-taxonomy.md).
- Why over-mocking makes a test Fragile, and what to mock instead, is in [state-vs-behaviour-and-what-to-mock.md](state-vs-behaviour-and-what-to-mock.md).
- The FIRST qualities (Repeatable, Independent) the Erratic smell violates are in [test-structure-and-first.md](test-structure-and-first.md).
- The coupling vocabulary behind "Fragile because over-coupled" is owned by **connascence-guide**; a read-only smell audit that _detects_ missing or over-mocked tests routes the remedy here, and is owned by **code-quality-guide**.
