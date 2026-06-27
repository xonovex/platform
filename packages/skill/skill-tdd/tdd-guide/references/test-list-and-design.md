# tdd: The Test List and Letting Tests Drive Design

TDD plans with a **test list** and designs through the act of testing. The list keeps the cycle from sprawling, and the production API is treated as an output of the tests rather than a blueprint drawn before them.

## The test list (to-do list)

Before starting, write down every test you can think of for the behaviour you are building — operations that should work, edge cases, variants, error conditions. Crucially, you **list** them; you do not write them. The list is a parking lot for test _ideas_.

As you work the list:

- Pick one item, drive it through red -> green -> refactor, then cross it off.
- When a new case occurs to you mid-cycle, add it to the list instead of chasing it now.
- When a refactoring you cannot do yet becomes necessary, add it to the list too.
- Stop when the list is empty.

Working one item at a time is what preserves the cycle. Batch-writing ten failing tests up front gives you ten red bars at once, no safe green platform to refactor from, and a pile of half-specified behaviour — it abandons the rhythm that makes TDD work.

## The API as a design output

Because the test is written first, you experience your own API as its first client before any implementation exists. Awkwardness shows up immediately: a constructor that needs five collaborators, an assertion that has to reach through three objects, a setup that takes twenty lines. Those are design signals delivered at the cheapest possible moment. You change the API you wish you had, then make it real. Over many cycles the shape of the module — its functions, parameters, and seams — accretes from real usage rather than from speculation.

The refactor step is where that emergent design is actually improved. TDD tells you _when_ to improve the design (every green bar) and gives you a safety net (the tests) for doing so; it does not tell you _what_ good design is. That bar — cohesion, coupling, responsibilities — is owned elsewhere (see Cross-references).

## Classical vs mockist as a design-pressure choice

How you handle a unit's collaborators in its tests is an orthogonal choice from test-first, and it pushes the design in different directions:

- **Classical (Detroit) TDD** uses real collaborators wherever practical and substitutes a test double only where the real thing is awkward (slow, non-deterministic, or external). Tests focus on the _result_ of the behaviour — given inputs, assert the output or resulting state — not on how the result was reached.
- **Mockist (London) TDD** uses a mock for any collaborator with interesting behaviour, and asserts the expected _calls_ between the unit and its collaborators.

The consequence is a coupling tradeoff. Mockist tests state the expected interactions explicitly, so they are coupled to the _implementation_: change how the unit collaborates and the tests break even when the observable behaviour is unchanged. They reward you with smaller, faster fixtures and immediate feedback about which unit failed, and they pressure the design toward clearly defined roles and interfaces between collaborators. Classical tests are coupled only to the result, so they survive internal refactoring and give better integration coverage, but a failure points at a cluster of objects rather than one, and large object graphs make setup heavier.

Neither is "more modern" or correct by default; pick per situation. And do not conflate the axes: test-first is the _rhythm_, classical-vs-mockist is the _collaborator-substitution style_. You can do classical TDD test-first, and you can do mockist TDD test-first.

## BAD -> GOOD

```
BAD — batch-written tests, no list, mockist style asserting internals reflexively
  it('charges and notifies', () => {
    // ten tests written before any code, all red at once
    expect(gateway.charge).toHaveBeenCalledWith(1250, 'EUR')   // couples to HOW
    expect(logger.info).toHaveBeenCalled()                      // tests an internal call
  })

GOOD — a worked list, one item at a time, asserting results
  // test list:
  //   [x] empty order totals 0
  //   [x] one line totals its price
  //   [ ] two lines sum
  //   [ ] a paid order is marked confirmed
  it('marks a paid order confirmed', () => {
    const order = placeOrder([{ price: 1250 }])
    const result = pay(order, fakeGateway({ outcome: 'approved' }))
    expect(result.status).toBe('confirmed')   // the RESULT, survives refactoring
  })
```

## Cross-references

- The cycle each list item runs through — see [red-green-refactor.md](red-green-refactor.md).
- How to make a single item go green — see [green-bar-strategies.md](green-bar-strategies.md).
- The mechanism behind classical vs mockist — state-verification vs behaviour-verification, what to mock and what not to, and the full test-double taxonomy — owned by **testing-guide**.
- The quality bar the refactor step aims at (responsibilities, cohesion, coupling) — owned by **oop-guide** and **connascence-guide**.
- Driving the outer acceptance loop from agreed examples before the inner TDD loop — owned by **bdd-guide**.
