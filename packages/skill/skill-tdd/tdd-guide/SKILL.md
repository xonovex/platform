---
name: tdd-guide
description: "Use when driving code test-first or coaching the red-green-refactor rhythm — writing a failing test before the production code, going green with fake-it / obvious-implementation / triangulation, then refactoring to remove duplication, working a test list one item at a time, and letting the tests grow the design. Triggers on red-green-refactor, test-first, 'write the test first', failing test before code, triangulation, fake it till you make it, test list / to-do list, classical (Detroit) vs mockist (London) TDD, and 'clean code that works' — even when the user doesn't say 'TDD'. Skip the anatomy of a single test (AAA, FIRST, naming, test doubles, mocking, smells -> see **testing-guide**), acceptance-first / example-led Given-When-Then specification (see **bdd-guide**), and the design bar the refactor step aims at (see **oop-guide**, **connascence-guide**)."
---

# Test-Driven Development

Drive production code from a failing test, one small step at a time: red (write a test that fails), green (make it pass by any means), refactor (remove the duplication you just created). The goal is clean code that works — and the design emerges from the tests rather than being drawn up front.

## Essentials

- **Run the loop red -> green -> refactor** - never skip refactor; green-without-refactor is not TDD, see [references/red-green-refactor.md](references/red-green-refactor.md)
- **Write the failing test before the code** - production code exists only to make a red test pass, see [references/red-green-refactor.md](references/red-green-refactor.md)
- **Pick a green-bar strategy per step** - obvious implementation, fake it, or triangulate, see [references/green-bar-strategies.md](references/green-bar-strategies.md)
- **Keep a test list, work one item at a time** - list test ideas, never batch-write them, see [references/test-list-and-design.md](references/test-list-and-design.md)
- **Let the tests grow the design** - the emerging API is an output; judge the refactor against **oop-guide** and **connascence-guide**, see [references/test-list-and-design.md](references/test-list-and-design.md)

## Gotchas

- Triangulation is a fallback, not the canonical move — reach for it only when you cannot yet see the general solution; obvious implementation and fake it are used far more often.
- Skipping refactor leaves the duplication you committed to go green (including a constant duplicated between test and code) — removing it is exactly what the third step is for.
- "Refactor" means change structure without changing behaviour while the test stays green; it is not "rewrite", "optimize", or "add the next feature".
- The test list is a list of test IDEAS, not a batch of tests written up front — write them one at a time to preserve the cycle.
- Test-first is the rhythm; classical vs mockist is an orthogonal design choice — you can do classical TDD test-first, so don't conflate the two.
- Neither classical nor mockist is "more modern" — mockist trades integration coverage and tighter implementation-coupling for faster, smaller fixtures.
- A single good test's anatomy (AAA, FIRST, naming, the test-double taxonomy) belongs to **testing-guide** — this skill owns only the rhythm in which you write it.

## Example

```
Test list (shopping-cart total):
[ ] empty cart totals 0
[ ] one line item totals its price
[ ] two items sum
[ ] a percentage discount applies

RED      test: total([]) === 0                       -> fails (no function)
GREEN    return 0                                    -> obvious implementation
RED      test: total([{price: 500}]) === 500         -> fails
GREEN    return items[0].price                       -> fake it (duplicated constant)
RED      test: total([a, b]) === a.price + b.price    -> fails, forces generalize
GREEN    items.reduce((s, i) => s + i.price, 0)       -> sum; constant duplication gone
REFACTOR name the fold, extract lineTotal()           -> structure only, test stays green
```

## Progressive Disclosure

- Read [references/red-green-refactor.md](references/red-green-refactor.md) - Load when running or coaching the cycle: the two rules, what each of red/green/refactor allows, small steps, and why refactor is non-optional.
- Read [references/green-bar-strategies.md](references/green-bar-strategies.md) - Load when deciding how to make a red test pass: obvious implementation vs fake it vs triangulation and how to choose.
- Read [references/test-list-and-design.md](references/test-list-and-design.md) - Load when planning what to test next or letting tests shape the API: the test list, one-at-a-time discipline, classical vs mockist as a design choice.
