# state-vs-behaviour-and-what-to-mock: Verify the Outcome, Mock the Seam

A test verifies a result in one of two ways, and it succeeds or fails depending on what you chose to double. Get both decisions right and the test stays robust under refactoring; get them wrong and it breaks every time the implementation moves.

## Contents

- State verification vs behaviour verification
- Only mocks insist on behaviour
- A stub answers queries; a mock verifies commands
- What to mock
- What NOT to mock
- Over-mocking is a coupling signal
- Cross-references

## State verification vs behaviour verification

- **State verification** asks: after the action, is the _result_ correct? You inspect the return value, or the state of the SUT or a Fake. "After `chargeRide`, the confirmation says `paid` and the in-memory store has one receipt."
- **Behaviour verification** asks: did the SUT make the _right calls_ on its collaborators? You assert that a specific method was invoked with specific arguments. "The SUT called `gateway.charge(1250, 'EUR')` exactly once."

State verification checks the outcome of behaviour; behaviour verification checks how the outcome was achieved. The first survives a refactor that changes the internal calls but keeps the result; the second does not. Prefer state verification by default, and reach for behaviour verification only when there is no observable state to check.

## Only mocks insist on behaviour

Among the five doubles, only a **Mock** insists on behaviour verification — that is its defining job. A Dummy, Stub, Spy, or Fake can all be used with state verification, and usually are. A Spy _enables_ behaviour verification but does not _insist_ on it: the test decides whether to assert on what the Spy recorded. A Mock is the only double that fails on its own when the expected calls don't happen.

So "should this be a Mock?" reduces to "is the call itself the thing under test, with no observable result to assert?" If yes, mock it. If there is a result, verify the result and use a Stub or Fake.

## A stub answers queries; a mock verifies commands

A clean rule for which double a collaborator gets:

- A collaborator the SUT **queries** for a value (a fare table, a clock, a config reader) is a **Stub** — you supply the answer and verify the SUT's resulting state. Never assert that a query was made; queries have no side effects, so the call itself doesn't matter.
- A collaborator the SUT **commands** to do something with no return worth checking (charge a card, enqueue a dispatch, send an email) is a **Mock** or a **Spy** — the call _is_ the effect, so verifying it is verifying the behaviour.

Folklore treats "stub" and "mock" as interchangeable. They are not: a Stub feeds the happy path and supports state verification; a Mock enforces commands and is the only double that does behaviour verification.

## What to mock

Double the things at the **edges** of your system — the architectural seams:

- True **indirect outputs**: the payment gateway, the email/notification service, the dispatch queue, an outbound HTTP call.
- Slow, non-deterministic, or external dependencies: the network, the wall clock, the filesystem, a real database (often a **Fake** in-memory store rather than a Mock).
- Collaborators whose real behaviour would make the test break FIRST (Fast/Repeatable) — see [test-structure-and-first.md](test-structure-and-first.md).

## What NOT to mock

- **Value objects** and plain data — a `Money`, an `Address`, a date range. Construct the real thing; mocking a value with no behaviour adds noise and proves nothing.
- **The system under test itself** — mocking part of the SUT means you are testing the mock, not the code. If you feel the urge, the SUT is doing too much; split it.
- **Internal collaborators with no interesting behaviour** — a private helper, a pure mapper. Let them run for real; doubling them welds the test to the call graph.
- **Things you don't own** — third-party libraries. Mocking them encodes your _assumptions_ about their API; wrap them behind a seam you own and double that seam instead.

## Over-mocking is a coupling signal

When a test needs a long list of doubles, or asserts on a precise sequence of internal calls, the test is not the problem — it is reporting one. Asserting on calls (behaviour verification) where state would do couples the test to _how_ the code works, so any honest refactor turns the bar red even though behaviour is unchanged. That fragility is a symptom of high coupling in the production design.

This skill flags the symptom. The diagnosis — what kind of coupling, and how to reduce it — is owned by **connascence-guide**. The remedy is usually fewer collaborators and a smaller, more cohesive unit, not a cleverer fixture.

## Cross-references

- The five doubles and how to name them are in [test-double-taxonomy.md](test-double-taxonomy.md).
- Fixtures that hide an over-coupled design, and the smells over-mocking produces, are in [test-smells-and-fixtures.md](test-smells-and-fixtures.md).
- Classical (real collaborators) vs mockist (always-double) as a development _style_ is owned by **tdd-guide**; this file owns the underlying state-vs-behaviour mechanism it cross-references.
- The coupling vocabulary behind "fragile because over-coupled" is owned by **connascence-guide**.
