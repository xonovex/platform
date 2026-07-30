# test-double-taxonomy: The Five Doubles and the Reconciliation Table

**Test Double** is the umbrella term for any object substituted for a real collaborator. There are five kinds, routinely collapsed into the word "mock". The name is documentation: pick the one that matches what the double does.

## The five kinds, precisely

- **Dummy**: fills a parameter slot but is never used by the receiver (an auth token a signature requires yet never reads).
- **Stub**: provides canned answers to the calls made during the test, feeding the SUT **indirect inputs**. No logic, no recording, no assertions.
- **Spy**: a Stub that _also_ records how it was called so the test can assert on it **afterward**. Holds no expectations itself.
- **Mock**: pre-programmed with **expectations** it **verifies** itself, failing on an unexpected, missing, or out-of-spec call. This is behaviour verification.
- **Fake**: a working but shortcut implementation unsuitable for production (in-memory DB, a gateway that always approves under €1000). Has real behaviour; a Stub does not.

Essential contrasts: a **Spy is not a Mock** (a Spy records passively; a Mock owns and enforces expectations); a **Fake is not a Stub** (a Fake runs logic; a Stub returns constants).

## Indirect input vs indirect output

**Indirect inputs** are values the SUT _receives_ from a collaborator; **indirect outputs** are calls the SUT _makes_ with no return worth checking. The double follows the direction:

- Control indirect **inputs** → **Stub** (or Fake) supplies the value.
- Verify indirect **outputs** → **Spy** (assert after) or **Mock** (assert via expectations).
- A **Fake** covers collaborators with both: it stores and returns.

## Responder vs Saboteur stubs

Both are Stubs; the difference is the value injected. **Responder** feeds valid values for the happy path (fare table returns `1250`); **Saboteur** injects errors/exceptions for the unhappy path (fare table throws `"route unavailable"`): use it to test failure handling without making the real collaborator fail.

## The reconciliation table

Reconciles authoritative usage with loose talk so a review can translate one team's vocabulary into another's.

| Kind  | Authoritative meaning                           | Common / loose usage             | Verifies                       |
| ----- | ----------------------------------------------- | -------------------------------- | ------------------------------ |
| Dummy | Filler value, never used by the receiver        | "placeholder", "null object"     | nothing                        |
| Stub  | Canned answers; supplies indirect inputs        | often called a "mock"            | state (indirect input)         |
| Spy   | Stub that records calls for the test to assert  | "mock", "recorder"               | indirect output (test asserts) |
| Mock  | Pre-set expectations the double itself enforces | any double of any kind           | behaviour (indirect output)    |
| Fake  | Real, shortcut implementation (in-memory store) | "stub", "mock", "in-memory mock" | state (real logic)             |

These are points on a continuum, not crisp boxes (a Spy is "a Stub that records"; a Fake can be stubbed). What matters: does the double supply inputs (Stub/Fake) or let you check outputs (Spy/Mock), and does the checking live in the double (Mock) or the test (Spy)?

## Cross-references

- Whether to verify state or behaviour, and what to mock at all: [state-vs-behaviour-and-what-to-mock.md](state-vs-behaviour-and-what-to-mock.md).
- Where a double sits in the Arrange phase: [test-structure-and-first.md](test-structure-and-first.md).
- Classical vs mockist as a _design_ style is owned by **tdd-guide**; a coupling diagnosis by **connascence-guide**; framework double APIs by **vitest-guide**.
