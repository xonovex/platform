# tdd: The Red-Green-Refactor Cycle

Two rules generate everything else:

1. Write production code only to make a failing test pass.
2. Eliminate duplication.

"Works" comes from rule 1 (every line is demanded by a test), "clean" from rule 2 (the refactor step). Non-obvious constraints:

- **Red must fail for the right reason**: a missing behaviour, not a typo. A test green before you write any code tests nothing. Invent the API you wish you had; write no production code.
- **Green is allowed to cheat**: hard-coded constant, duplicated value, ugly conditional. Speed over cleanliness; a passing test is the platform to refactor from. Add no behaviour a test doesn't demand.
- **Refactor means change structure while behaviour and the green bar stay identical**: remove the constant now duplicated between test and code, rename, extract, inline. It is NOT "rewrite", "optimize", or "slip in the next feature"; new behaviour needs its own red test first.

**Green-without-refactor is not TDD.** Going green deliberately incurs duplication (a faked constant, a copied branch); the refactor step is where it is paid back. Skip it and duplication compounds until the design ossifies.

Step size is a tuning dial: take big steps when the code is obvious, shrink them when the last red surprised you. Always be able to go green within a minute or two, if you can't, the step was too big; back out and split.

```
// RED:   discountedTotal([], 0.1)          === 0    -> fails
// GREEN: return 0
// RED:   discountedTotal([{price:500}], 0) === 500  -> fails
// GREEN: return items[0].price
// RED:   discountedTotal([a,b], 0)         === sum   -> forces the fold
// GREEN: items.reduce((s,i)=>s+i.price,0)
// RED:   discountedTotal([{price:100}],0.1)=== 90    -> forces the rate
// GREEN: subtotal * (1 - rate)
// REFACTOR: extract subtotal(); no loyalty bonus until a test asks for one
```

## Cross-references

- Choosing how to go green, see [green-bar-strategies.md](green-bar-strategies.md).
- Test list and letting the API emerge, see [test-list-and-design.md](test-list-and-design.md).
- Anatomy of the individual test (AAA, FIRST, naming, doubles): **testing-guide**.
- The quality bar the refactor aims at: **oop-guide**, **connascence-guide**.
