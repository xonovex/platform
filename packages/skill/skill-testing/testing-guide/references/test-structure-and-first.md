# test-structure-and-first: AAA / Four-Phase Shape and the FIRST Qualities

## Arrange-Act-Assert

1. **Arrange** — build the SUT and inputs, wire up doubles, set starting state.
2. **Act** — invoke the one behaviour under test, ideally a single call.
3. **Assert** — verify the outcome.

Separate the blocks visually (a blank line). If you cannot tell where Arrange ends and Act begins, the test does too much.

## The Four-Phase Test and the teardown distinction

The Four-Phase Test is the same shape plus a fourth phase: **Setup → Exercise → Verify → Teardown**. Setup/Exercise/Verify equal Arrange/Act/Assert; the difference is explicit **Teardown** releasing resources Setup acquired. AAA is **three** phases, not four — a true unit test works on in-memory objects the GC reclaims. Reach for explicit Teardown only when Setup grabbed a real external resource, which a fast unit test usually should not.

## One action, one logical assertion

The Act phase is a single call; multiple actions test a sequence, not a unit. The Assert phase checks **one logical outcome** (a single reason to fail) — that can be several `expect` lines describing one result, but not unrelated results bolted together. Verifying many things at once is an Eager Test (see [test-smells-and-fixtures.md](test-smells-and-fixtures.md)).

## The FIRST qualities

- **Fast** — milliseconds; slowness comes from real I/O, so double the seams.
- **Independent** (also **Isolated**) — runs in any order; no shared mutable fixture leaks.
- **Repeatable** — same result every run/environment; no wall clock, random seed, network, or time zone (inject a clock; seed randomness).
- **Self-validating** (also **Self-verifying**) — a boolean pass/fail with no manual interpretation.
- **Timely** — written close in time to the production code, ideally just before it.

The "I" (Independent / Isolated) and "S" (Self-validating / Self-verifying) both circulate with the same intent; neither spelling is the only correct one.

## Naming a test

Say what behaviour holds, not how the code works. Shape: **subject + condition + expected outcome** (`rejects a card charge over the per-ride limit`). Avoid names tied to method names (`testCharge1`) or internals (`callsGatewayThenSavesReceipt`) — they break on rename/refactor. When a test fails, its name alone should explain what stopped being true.

## BAD vs GOOD

```ts
// BAD — no phase separation, two actions, unrelated assertions, internal-coupled name
test("testFlow", () => {
  const c = new Cart();
  c.add(item("A", 500));
  c.add(item("B", 700));
  expect(c.count()).toBe(2);
  c.applyDiscount(0.1); // ...a second action
  expect(c.total()).toBe(1080); // ...an unrelated outcome
  expect(c.callsToRecalc).toBe(2); // ...an internal call
});

// GOOD — one behaviour, AAA blocks, one logical outcome, behaviour-named
test("applies a 10% discount to the cart subtotal", () => {
  const cart = cartWith(item("A", 500), item("B", 700)); // Arrange
  const total = cart.totalWithDiscount(0.1); // Act
  expect(total).toBe(1080); // Assert
});
```

## Cross-references

- Choosing/naming the Arrange-phase doubles: [test-double-taxonomy.md](test-double-taxonomy.md). Whether Assert checks state or calls: [state-vs-behaviour-and-what-to-mock.md](state-vs-behaviour-and-what-to-mock.md). Smells violating FIRST and fixture helpers: [test-smells-and-fixtures.md](test-smells-and-fixtures.md).
- _When_ to write the test (red-green-refactor) is owned by **tdd-guide**.
