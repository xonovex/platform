# tdd: The Red-Green-Refactor Cycle

TDD is a rhythm of three states repeated in tight loops: write a small test that fails (**red**), make it pass as quickly as possible (**green**), then eliminate the duplication you created getting there (**refactor**). The mantra names the order, and the order is not negotiable.

## The two rules

TDD rests on two rules:

1. Write production code only to make a failing test pass.
2. Eliminate duplication.

Everything else — small steps, fast feedback, an emerging design — falls out of obeying these two. The overall goal they serve is **clean code that works**: "works" comes from rule 1 (every line is demanded by a test), "clean" comes from rule 2 (the refactor step that removes duplication).

## The three states and what each allows

- **Red — write a failing test.** Write the smallest test that expresses the next behaviour you want, for code that does not exist or does not yet do that. It must fail, and fail for the right reason (a missing behaviour, not a typo). Seeing red proves the test can fail — a test that was green before you wrote any code tests nothing. You are allowed to invent the API you wish you had here; you are not allowed to write production code.
- **Green — make it pass, fast.** Do the least that turns the bar green, committing whatever sins are necessary: a hard-coded constant, a duplicated value, an ugly conditional. Speed matters more than cleanliness, because a passing test is a safe platform to refactor from. You are not allowed to add behaviour no test demands.
- **Refactor — remove duplication.** With the test green, change the structure of the code without changing its behaviour: remove the duplication you just introduced (notably the constant now duplicated between the test and the production code), rename, extract, inline. The tests stay green the whole time and tell you the moment a change alters behaviour. You are not allowed to add new behaviour in this step — that needs a new red test first.

## Why the refactor step is non-optional

Going green legitimately produces duplication — a faked constant, a copy-pasted branch, a literal that appears in both the test and the code. That debt is taken on deliberately to reach green fast. The refactor step is where it is paid back. Skip it and the duplication compounds across cycles until the design ossifies and further tests get hard to write. **Green-without-refactor is not TDD**; it is just writing code after a test.

"Refactor" has a precise meaning here: change internal structure while behaviour stays identical and the tests stay green. It is not "rewrite", not "optimize for speed", and not "slip in the next feature". If the behaviour changes, that change must have been driven by its own failing test.

## Small steps

The size of each step is a tuning dial, not a fixed law. When the code is obvious, take big steps (write the real implementation directly). When you are unsure or the last red surprised you, shrink the steps until each one is trivially safe. The discipline is to always be able to go green within a minute or two — if you cannot, the step was too big; back out and split it.

## BAD -> GOOD

```
BAD — no failing test, behaviour added speculatively, never refactored
  function discountedTotal(items, rate) {
    // wrote the whole thing, then bolted a test under it that passed first try
    return items.reduce((s, i) => s + i.price, 0) * (1 - rate)
      + maybeLoyaltyBonus(items);   // no test demands this branch
  }

GOOD — each behaviour driven by a red, made green, then cleaned
  // RED:   discountedTotal([], 0.1)            === 0      -> fails
  // GREEN: return 0
  // RED:   discountedTotal([{price:500}], 0)   === 500    -> fails
  // GREEN: return items[0].price
  // RED:   discountedTotal([a,b], 0)           === sum     -> forces the fold
  // GREEN: items.reduce((s,i)=>s+i.price,0)
  // RED:   discountedTotal([{price:100}], 0.1) === 90      -> forces the rate
  // GREEN: subtotal * (1 - rate)
  // REFACTOR: extract subtotal(); no loyalty bonus until a test asks for one
```

## Cross-references

- Choosing how to go green (obvious implementation / fake it / triangulation) — see [green-bar-strategies.md](green-bar-strategies.md).
- Keeping a test list and letting the API emerge — see [test-list-and-design.md](test-list-and-design.md).
- The anatomy of the individual failing test you write (AAA structure, FIRST qualities, naming, test doubles) — owned by **testing-guide**.
- The quality bar the refactor step aims at (good OO design, low coupling) — owned by **oop-guide** and **connascence-guide**.
