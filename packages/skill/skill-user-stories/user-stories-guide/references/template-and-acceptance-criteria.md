# user-stories: Story Template and Acceptance Criteria

The three-part template captures who wants something, what they want, and why — and deliberately omits how. Acceptance criteria then make "done" concrete, covering far more than the happy path.

## The three-part template

```
As a [type of user], I want [goal / capability], so that [reason / business value].
```

- **As a** — names the user role and builds empathy for who is asking. It anchors the story in a person, not a system component.
- **I want** — captures **what** the user needs, expressed as a capability, not an implementation. It says nothing about screens, tables, or endpoints.
- **so that** — captures the **why**: the motivation and the business value. This is the part most often dropped, and the part that lets you challenge, reprioritize, or drop the story.

Rationale: the template forces What and Why while keeping How open, which is exactly what makes a story **Negotiable** (the team chooses the How in conversation). A story missing "so that" cannot be evaluated for value; a story whose "I want" describes a mechanism ("I want a Redis cache") has smuggled the How into the What.

The template is a tool for starting a conversation, not a mandatory format. No method prescribes a fixed story format, and a one-line note can be a perfectly good story when the team shares enough context.

## Acceptance criteria

Acceptance criteria are the Confirmation from the 3 Cs (see [three-cs.md](three-cs.md)) written down: the conditions that must hold for the story to be accepted. Good coverage spans four categories, not just the first:

- **Happy path** — the main success scenario the story exists to enable.
- **Boundaries** — edge values: empty, zero, minimum, maximum, first, last.
- **Errors / invalid cases** — what happens when input is rejected, a dependency fails, or a precondition is unmet.
- **UI changes** — the visible state changes: what the user now sees, what becomes enabled or hidden.

Acceptance criteria CAN be expressed as Given-When-Then examples — which makes them straightforward to automate — but the Given-When-Then notation, feature files, and the discovery practice around them are owned by **bdd-guide**. This skill stops at "the criteria exist and cover the four categories"; it does not own that notation or any executable-spec tooling. Designing the tests themselves (levels, doubles, structure) is owned by **testing-guide**.

## BAD → GOOD

BAD (How smuggled into What; happy-path-only criteria):

```
As a backend engineer, I want a charges table with a foreign key to trips
  so that we can store payments.

Acceptance: a row is inserted when a trip ends.
```

GOOD (role + capability + value; four-category criteria):

```
As a rider, I want my completed trip charged to my saved card
  so that I can leave the cab without handling cash.

Acceptance criteria:
  happy     valid saved card is charged the fare; a receipt is shown
  boundary  a zero-distance trip is charged the minimum fare
  error     a declined card blocks ending the trip and prompts a retry
  UI        the trip screen shows "Paid" and the amount once the charge succeeds
```

## Cross-references

- The Card and Confirmation that the template and criteria realize — [three-cs.md](three-cs.md).
- Expressing criteria as Given-When-Then for automation — **bdd-guide**.
- Designing and structuring the tests behind the criteria — **testing-guide**.
