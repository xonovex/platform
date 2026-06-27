# test-structure-and-first: AAA / Four-Phase Shape and the FIRST Qualities

A good test reads in one pass and fails for exactly one reason. Two conventions get you there: a fixed phase structure (Arrange-Act-Assert, or the Four-Phase Test) and a checklist of qualities (FIRST). Together they make a test fast to read, fast to run, and trustworthy when it goes red.

## Contents

- Arrange-Act-Assert
- The Four-Phase Test and the teardown distinction
- One action, one logical assertion
- The FIRST qualities
- Naming a test
- BAD vs GOOD
- Cross-references

## Arrange-Act-Assert

Structure the body of every test in three blocks, in order:

1. **Arrange** — build the SUT and its inputs, wire up any doubles, set the starting state.
2. **Act** — invoke the one behaviour under test, ideally a single call.
3. **Assert** — verify the outcome.

Separate the blocks visually (a blank line is enough). A reader should locate "what is set up", "what is exercised", and "what is checked" without parsing the whole method. If you cannot tell where Arrange ends and Act begins, the test is doing too much.

## The Four-Phase Test and the teardown distinction

The Four-Phase Test names the same shape with a fourth phase: **Setup → Exercise → Verify → Teardown**. Setup/Exercise/Verify are exactly Arrange/Act/Assert under different names. The difference is the explicit **Teardown** phase that releases resources the Setup acquired (a temp file, a real connection).

AAA omits Teardown because a true unit test rarely allocates anything that needs cleaning up — it works on in-memory objects the garbage collector reclaims. Folklore calls AAA a "four-phase pattern" and bolts on a mandatory teardown; that is imprecise. AAA is three phases. Reach for explicit Teardown only when the Setup grabbed a real external resource, which in a fast unit test it usually should not.

## One action, one logical assertion

The Act phase should be a single call. Multiple actions in one test mean you are testing a sequence, not a unit, and the failure won't tell you which step broke.

The Assert phase should check **one logical outcome** — a test should have a single reason to fail. That can be several physical `expect` lines that together describe one result (status is `paid` _and_ one receipt exists _and_ the confirmation id is set), but not several unrelated results bolted into one method. A test that verifies many things at once is an Eager Test (see [test-smells-and-fixtures.md](test-smells-and-fixtures.md)); split it so each failure points at one cause.

## The FIRST qualities

A unit test should be **FIRST**:

- **Fast** — runs in milliseconds, so the whole suite runs constantly. Slowness comes from real I/O — double the seams to keep it fast.
- **Independent** (also written **Isolated**) — stands alone and runs in any order. No test depends on another's leftover state; no shared mutable fixture leaks between tests.
- **Repeatable** — same result every run, in every environment: laptop, CI, offline. No reliance on the wall clock, random seeds, network, or time zone (inject a clock; seed randomness).
- **Self-validating** (also written **Self-verifying**) — produces a boolean pass/fail with no manual interpretation. No reading log output by eye, no "looks right". The test asserts the verdict itself.
- **Timely** — written close in time to the production code, ideally just before it, so the code stays testable and the test stays honest.

The "I" (Independent / Isolated) and the "S" (Self-validating / Self-verifying) both circulate with the same intent; don't present either spelling as the only correct one.

## Naming a test

The name is the test's headline; it should say what behaviour holds, not how the code works. A useful shape is **subject + condition + expected outcome** — for example `assigns the closer taxi when two taxis have equal rating`, or `rejects a card charge over the per-ride limit`. Avoid names tied to method names (`testCharge1`) or to internals (`callsGatewayThenSavesReceipt`) — those break when you rename or refactor and tell a reader nothing about the rule under test. When a test fails, its name alone should explain what stopped being true.

## BAD vs GOOD

```
// BAD — no phase separation, two actions, unrelated assertions, internal-coupled name
test("testFlow", () => {
  const c = new Cart(); c.add(item("A", 500)); c.add(item("B", 700));
  expect(c.count()).toBe(2);          // checks one thing
  c.applyDiscount(0.1);               // ...then a second action
  expect(c.total()).toBe(1080);       // ...and an unrelated outcome
  expect(c.callsToRecalc).toBe(2);    // asserts an internal call
});

// GOOD — one behaviour, AAA blocks, one logical outcome, behaviour-named
test("applies a 10% discount to the cart subtotal", () => {
  // Arrange
  const cart = cartWith(item("A", 500), item("B", 700));
  // Act
  const total = cart.totalWithDiscount(0.1);
  // Assert
  expect(total).toBe(1080);
});
```

## Cross-references

- Choosing and naming the doubles used in the Arrange phase is in [test-double-taxonomy.md](test-double-taxonomy.md).
- Whether the Assert phase checks state or calls is in [state-vs-behaviour-and-what-to-mock.md](state-vs-behaviour-and-what-to-mock.md).
- Smells that violate FIRST (Erratic, Obscure, Eager) and fixture helpers for the Arrange phase are in [test-smells-and-fixtures.md](test-smells-and-fixtures.md).
- _When_ to write the test relative to the code (test-first rhythm, red-green-refactor) is owned by **tdd-guide**; this file owns the shape of the test once you write it.
