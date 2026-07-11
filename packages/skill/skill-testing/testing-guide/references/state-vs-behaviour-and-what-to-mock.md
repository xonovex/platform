# state-vs-behaviour-and-what-to-mock: Verify the Outcome, Mock the Seam

- **State verification** — after the action, is the _result_ correct? Inspect the return value or the state of the SUT/Fake.
- **Behaviour verification** — did the SUT make the _right calls_ on its collaborators (`gateway.charge(1250, 'EUR')` exactly once)?

State verification survives a refactor that changes internal calls but keeps the result; behaviour verification does not. Prefer state by default; reach for behaviour only when there is no observable state to check.

## Only mocks insist on behaviour

A Dummy, Stub, Spy, or Fake can all be used with state verification. A Spy _enables_ behaviour verification but the test decides whether to assert on it; a **Mock** is the only double that fails on its own when the expected calls don't happen. So "should this be a Mock?" reduces to "is the call itself the thing under test, with no observable result to assert?"

## A stub answers queries; a mock verifies commands

- A collaborator the SUT **queries** for a value (fare table, clock, config reader) → **Stub**: supply the answer, verify resulting state. Never assert a query was made.
- A collaborator the SUT **commands** with no return worth checking (charge a card, enqueue a dispatch, send an email) → **Mock** or **Spy**: the call _is_ the effect.

## What to mock — the architectural seams

- True **indirect outputs**: payment gateway, email/notification service, dispatch queue, outbound HTTP.
- Slow/non-deterministic/external: network, wall clock, filesystem, real DB (often a **Fake** in-memory store rather than a Mock).

## What NOT to mock

- **Value objects** and plain data (`Money`, `Address`, a date range) — construct the real thing.
- **The SUT itself** — mocking part of the SUT tests the mock; if you feel the urge, the SUT does too much — split it.
- **Internal collaborators with no interesting behaviour** — private helpers, pure mappers; let them run.
- **Things you don't own** — third-party libraries; mocking encodes your _assumptions_ about their API. Wrap behind a seam you own and double that.

## Over-mocking is a coupling signal

A long list of doubles, or asserting on a precise sequence of internal calls, reddens the bar on any honest refactor even though behaviour is unchanged — a symptom of coupling in the production design. This skill flags the symptom; the diagnosis is owned by **connascence-guide**. The remedy is fewer collaborators and a smaller unit, not a cleverer fixture.

## Cross-references

- The five doubles: [test-double-taxonomy.md](test-double-taxonomy.md). Fixture smells over-mocking produces: [test-smells-and-fixtures.md](test-smells-and-fixtures.md).
- Classical vs mockist as a development _style_ is owned by **tdd-guide**; the coupling vocabulary by **connascence-guide**.
