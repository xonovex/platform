# tdd: The Test List and Letting Tests Drive Design

## The test list (to-do list)

Before starting, write down every test you can think of — operations, edge cases, variants, error conditions. You **list** them; you do not write them. The list is a parking lot for test _ideas_.

- Pick one item, drive it red -> green -> refactor, cross it off.
- New case mid-cycle: add it to the list, don't chase it now. Same for a refactoring you can't do yet.
- Stop when the list is empty.

One item at a time preserves the cycle. Batch-writing ten failing tests up front gives ten red bars at once, no green platform to refactor from, and a pile of half-specified behaviour.

## The API as a design output

Because the test is written first, you are your API's first client before any implementation exists. Awkwardness (a constructor needing five collaborators, an assertion reaching through three objects) surfaces at the cheapest moment. Change the API you wish you had, then make it real; the module's shape accretes from real usage, not speculation.

TDD tells you _when_ to improve the design (every green bar) and gives you a safety net; it does not tell you _what_ good design is — that bar is owned by **oop-guide** and **connascence-guide**.

## Classical vs mockist as a design-pressure choice

Orthogonal to test-first; it's the collaborator-substitution style, and it pushes the design differently:

- **Classical (Detroit)** — real collaborators wherever practical, a double only where the real thing is awkward (slow, non-deterministic, external). Assert the _result_ (output/state).
- **Mockist (London)** — a mock for any collaborator with interesting behaviour; assert the expected _calls_.

Tradeoff: mockist tests couple to the _implementation_ (change how the unit collaborates and they break even when observable behaviour is unchanged) but reward smaller/faster fixtures, precise failure localization, and pressure toward clear roles/interfaces. Classical tests couple only to the result, so they survive internal refactoring and give better integration coverage, but a failure points at a cluster of objects and large graphs make setup heavier. Neither is "more modern"; pick per situation. Don't conflate the axes — you can do classical or mockist TDD test-first.

## Cross-references

- The cycle each list item runs through — see [red-green-refactor.md](red-green-refactor.md).
- Making an item go green — see [green-bar-strategies.md](green-bar-strategies.md).
- State- vs behaviour-verification, what to mock, the test-double taxonomy — **testing-guide**.
- Driving the outer acceptance loop from agreed examples before the inner TDD loop — **bdd-guide**.
