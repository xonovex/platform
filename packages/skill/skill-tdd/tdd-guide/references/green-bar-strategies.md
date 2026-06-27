# tdd: Going Green — Fake It, Obvious Implementation, Triangulation

Once a test is red, there are exactly three ways to make it pass. Choosing between them is a judgement about how sure you are of the real code, and the choice shapes how big each step is.

## The three strategies

- **Obvious Implementation.** When you already know the real code and it is small and clear, just type it. No faking, no detour. This is the default whenever the implementation is obvious — most TDD steps are obvious implementations.
- **Fake It (till you make it).** When you are not yet sure of the real code, return a constant that satisfies the test. The test goes green immediately. Now there is duplication — the same literal appears in the test and in the production code — and removing that duplication during refactor is what drives you toward the real, general implementation. Fake It turns "I don't know how to write this" into "I know how to remove this duplication".
- **Triangulation.** Generalize an implementation only once two or more concrete examples force it. You write one test, fake it; you write a second test with a different value that the fake cannot satisfy; the two examples together pin down the general shape, and you abstract to it. Triangulation is the most conservative move — it refuses to generalize until the evidence demands it.

## The decision flow

```
Is the real implementation obvious to you right now?
  yes -> Obvious Implementation (type the real code)
  no  -> Fake It (return a constant; go green)
           Can you now see how to remove the test<->code duplication?
             yes -> generalize directly during refactor
             no  -> Triangulate (add a second example that forces the general form)
                      still stuck after triangulating?
                        -> step away from the keyboard; the design question is bigger than the next line
```

The order matters: prefer Obvious Implementation, drop to Fake It when unsure, and only reach for Triangulation when even the duplication between two states does not tell you the answer.

## Triangulation is a fallback, not the canonical move

Triangulation circulates as if it were _the_ TDD technique, but the authoritative position is more reserved: it is used only when you are completely unsure how to refactor toward the general solution. If you can already see how to remove the duplication, you write the general solution directly rather than ceremonially adding a second example to "justify" it. Treat Triangulation as the tool of last resort before stepping away — not the default rhythm.

## How Fake It exposes the real implementation

Fake It is not a trick to inflate the test count; it is a mechanism. The faked constant creates a concrete, visible duplication, and the refactor rule ("eliminate duplication") then has something to act on. Following the duplication mechanically often produces the general code without you having to design it in your head.

```
RED:   total([{price: 500}]) === 500
GREEN: return 500                      // Fake It — the literal 500 is duplicated
                                       //   (it lives in the test AND the code)
REFACTOR: the only way to delete the duplicated 500 is to read it from the input
       return items[0].price           // the fake dissolved into the real shape
RED:   total([{price: 500},{price: 300}]) === 800
GREEN/REFACTOR: items.reduce((s,i)=>s+i.price, 0)   // the sum the examples forced
```

## BAD -> GOOD

```
BAD — triangulating reflexively when the code is obvious
  // test: square(3) === 9
  GREEN: return 3 * 3            // faked
  // add a second test only to "earn" the general form
  // test: square(4) === 16
  GREEN: return n * n
  (the implementation was obvious from the first test — the detour was waste)

GOOD — match the strategy to your certainty
  // square is obvious -> Obvious Implementation, one step
  function square(n) { return n * n }

  // a tiered shipping fee is NOT obvious -> Fake It, then let duplication drive
  // test: fee(0)   === 0   -> GREEN return 0
  // test: fee(50)  === 5   -> GREEN return 5   (faked; now two constants)
  // test: fee(150) === 15  -> the three points force fee = weight * 0.1
```

## Cross-references

- The cycle these strategies live inside, and why the duplication they create must be refactored away — see [red-green-refactor.md](red-green-refactor.md).
- Choosing the next test to make red, and letting the API emerge — see [test-list-and-design.md](test-list-and-design.md).
- Structuring the individual test (Arrange-Act-Assert) and its qualities — owned by **testing-guide**.
