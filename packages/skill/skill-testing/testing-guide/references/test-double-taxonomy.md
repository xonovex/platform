# test-double-taxonomy: The Five Doubles and the Reconciliation Table

A **Test Double** is the umbrella term for any object substituted for a real collaborator in a test. There are five kinds — Dummy, Stub, Spy, Mock, and Fake — and they are routinely confused, most often by collapsing all five into the word "mock". Naming the right one makes a test say what it actually checks.

## Contents

- The umbrella term and why the word "mock" misleads
- The five kinds, precisely
- Indirect input vs indirect output
- Responder vs Saboteur stubs
- The reconciliation table (authoritative vs loose usage)
- The continuum caveat
- Cross-references

## The umbrella term and why the word "mock" misleads

"Test Double" — by analogy with a stunt double — covers every stand-in. Folklore says "a mock is any test double"; that is wrong. A Mock is one specific kind: a double pre-programmed with **expectations** about the calls it should receive, which it **verifies** at the end of the test. Reserve "mock" for that. Most objects people casually call mocks are Stubs that just answer queries — they make no assertions and verify nothing.

The cost of the loose word is real: a "mock" that only returns canned data tells a reader the test checks interactions when it checks state, and vice versa. The name is documentation; pick the one that matches what the double does.

## The five kinds, precisely

- **Dummy** — passed around to fill a parameter list but never actually used. An auth token a code path requires in its signature yet never reads. It is an alternative to the value patterns: the receiver never exercises it.
- **Stub** — provides canned answers to the calls made during the test, and only those. It feeds the system under test (SUT) **indirect inputs**. No logic, no recording, no assertions.
- **Spy** — a Stub that _also_ records information about how it was called, so the test can assert on it **afterward**. The Spy holds no expectations; the test does the checking.
- **Mock** — pre-programmed with expectations that form a specification of the calls it is expected to receive. It **verifies** those calls itself and fails on an unexpected, missing, or out-of-spec call. This is behaviour verification.
- **Fake** — a working but shortcut implementation unsuitable for production: an in-memory database, a fake payment gateway that always approves under €1000. A Fake has real behaviour; a Stub does not.

The load-bearing contrasts: a **Spy is not a Mock** (a Spy records passively; a Mock owns and enforces the expectations), and a **Fake is not a Stub** (a Fake really runs logic; a Stub returns constants).

## Indirect input vs indirect output

The SUT talks to collaborators in two directions. **Indirect inputs** are values the SUT _receives_ from a collaborator (a price returned by a fare table). **Indirect outputs** are calls the SUT _makes_ to a collaborator that have no return value worth checking (charging a card, sending a notification). The kind of double follows the direction:

- To control indirect **inputs**, use a **Stub** (or a Fake) — it supplies the value.
- To verify indirect **outputs**, use a **Spy** (assert after) or a **Mock** (assert via expectations).
- A **Fake** provides an alternative implementation for collaborators that have both — it stores and returns.

## Responder vs Saboteur stubs

A Stub injects indirect input, and there are two flavours of what it injects:

- **Responder** — injects valid values that drive the happy path: the fare table returns `1250`.
- **Saboteur** — injects errors or exceptions to drive the unhappy path: the fare table throws `"route unavailable"`.

Both are Stubs; the distinction is the value they feed in. Use a Saboteur to test how the SUT handles a collaborator's failure without making the real collaborator fail.

## The reconciliation table

The same five words are used differently by the authoritative taxonomy, by widely-cited usage, and by everyday loose talk. This table reconciles them so a review can translate one team's vocabulary into another's.

| Kind  | Authoritative meaning                           | Common / loose usage             | Verifies                       |
| ----- | ----------------------------------------------- | -------------------------------- | ------------------------------ |
| Dummy | Filler value, never used by the receiver        | "placeholder", "null object"     | nothing                        |
| Stub  | Canned answers; supplies indirect inputs        | often called a "mock"            | state (indirect input)         |
| Spy   | Stub that records calls for the test to assert  | "mock", "recorder"               | indirect output (test asserts) |
| Mock  | Pre-set expectations the double itself enforces | any double of any kind           | behaviour (indirect output)    |
| Fake  | Real, shortcut implementation (in-memory store) | "stub", "mock", "in-memory mock" | state (real logic)             |

When a colleague says "mock", check which column they mean before trusting that the test verifies interactions.

## The continuum caveat

These are points on a continuum, not always crisply separated boxes. A Spy is "a Stub that records"; a Mock library can be configured to behave like a Stub; a Fake can be stubbed. Use the names to communicate intent, not to win arguments about edge cases. What matters is whether the double supplies inputs (Stub/Fake) or lets you check outputs (Spy/Mock), and whether the checking lives in the double (Mock) or in the test (Spy).

## Cross-references

- Whether to verify state or behaviour, and what to mock at all, is in [state-vs-behaviour-and-what-to-mock.md](state-vs-behaviour-and-what-to-mock.md).
- Where a double sits in the Arrange phase is in [test-structure-and-first.md](test-structure-and-first.md).
- Choosing real collaborators vs doubles as a _design_ style (classical vs mockist) is owned by **tdd-guide**; this file owns the doubles themselves.
- A test that needs an unwieldy pile of doubles is a coupling signal — see **connascence-guide** for the diagnosis.
- Framework-specific double APIs (spies, auto-mocking, module mocks) are owned by **vitest-guide**.
