---
name: testing-guide
description: "Use when writing or reviewing a single good test independent of any framework — structuring it as Arrange-Act-Assert / Four-Phase, meeting FIRST, naming it, choosing and naming the right test double (Dummy / Stub / Spy / Mock / Fake), deciding what to mock and what not to, telling state from behaviour verification, and spotting test smells. Triggers on test double, mock vs stub vs fake vs spy, what should I mock, over-mocking, AAA, FIRST, fragile/flaky/obscure test, Object Mother, Test Data Builder, fixture setup — even when the user doesn't say 'testing'. Skip the test-first rhythm and red-green-refactor (see **tdd-guide**), acceptance/example specification and Gherkin (see **bdd-guide**), the coupling theory behind over-mocking (see **connascence-guide**), and any framework-specific API (see **vitest-guide**)."
---

# Writing a Single Good Test

How to structure, name, populate, and verify one test — and how to pick the right test double — independent of any framework. It governs the shape of an individual test, not the rhythm in which you write tests.

## Essentials

- **Structure every test as Arrange-Act-Assert** - set up, run one action, verify the outcome, see [references/test-structure-and-first.md](references/test-structure-and-first.md)
- **Make each test meet FIRST** - Fast, Independent, Repeatable, Self-validating, Timely, see [references/test-structure-and-first.md](references/test-structure-and-first.md)
- **Name the right double for the job** - Dummy / Stub / Spy / Mock / Fake are distinct, not synonyms, see [references/test-double-taxonomy.md](references/test-double-taxonomy.md)
- **Prefer state verification; mock only true outputs** - assert outcomes, reserve call-checking for external commands, see [references/state-vs-behaviour-and-what-to-mock.md](references/state-vs-behaviour-and-what-to-mock.md)
- **Mock at architectural seams, not internals** - double the payment gateway, never value objects or the system under test, see [references/state-vs-behaviour-and-what-to-mock.md](references/state-vs-behaviour-and-what-to-mock.md)
- **Read a failing test as a smell report** - fragile, obscure, eager, erratic tests each name a fix, see [references/test-smells-and-fixtures.md](references/test-smells-and-fixtures.md)

## Gotchas

- "Mock" is not a synonym for "test double". A Mock holds pre-programmed call expectations and is verified; most things people call mocks are actually Stubs that answer queries with canned state.
- A Stub answers queries (state); a Mock verifies commands (behaviour). Asserting on calls when state would do couples the test to the implementation and makes it Fragile.
- AAA is three phases, not four. Explicit Teardown belongs to the Four-Phase Test and is usually unnecessary for a true unit test.
- A Spy is not a Mock: a Spy passively records calls for the test to assert later; a Mock owns the expectations itself and fails on an unexpected or missing call.
- A Fake is not a Stub: a Fake has a real, working, shortcut implementation (an in-memory store); a Stub only returns hard-coded answers with no logic.
- A test that needs many doubles or asserts on internal calls is a coupling signal — fix the production design, and take the coupling diagnosis to **connascence-guide**, not a bigger fixture.
- FIRST's "I" (Independent / Isolated) and "S" (Self-validating / Self-verifying) both circulate; same intent, neither spelling is wrong.

## Example

```
// Arrange — one Dummy, one Stub, one Spy, one Mock, one Fake
const auditId   = "ignored";                       // Dummy: fills the signature, never read
const fares     = stubFare({ cityCenter: 1250 });  // Stub: canned answer (Responder)
const receipts  = spyReceipts();                   // Spy: records each send(...)
const gateway   = mockGateway().expect("charge", 1250, "EUR"); // Mock: command expectation
const cards     = inMemoryCardStore();             // Fake: real but shortcut storage

// Act — exactly one action
const confirmation = chargeRide(trip, gateway, receipts, fares, cards, auditId);

// Assert — state first, then the one behaviour expectation
expect(confirmation.status).toBe("paid");          // state verification
expect(receipts.sent).toHaveLength(1);             // spy: outcome, not call order
gateway.verify();                                  // behaviour verification of the command
```

## Progressive Disclosure

- Read [references/test-double-taxonomy.md](references/test-double-taxonomy.md) - Load when choosing or naming a double, or settling mock vs stub vs fake vs spy
- Read [references/state-vs-behaviour-and-what-to-mock.md](references/state-vs-behaviour-and-what-to-mock.md) - Load when deciding what to mock, whether to assert on state or calls, or diagnosing over-mocking
- Read [references/test-structure-and-first.md](references/test-structure-and-first.md) - Load when structuring a test (AAA / Four-Phase), naming it, or checking it against FIRST
- Read [references/test-smells-and-fixtures.md](references/test-smells-and-fixtures.md) - Load when a test is fragile/obscure/eager/erratic, or when building fixtures with a Builder or Object Mother
