# user-stories: Story Template and Acceptance Criteria

```
As a [type of user], I want [goal / capability], so that [reason / business value].
```

- **As a** — the user role (a person, not a system component).
- **I want** — **what** the user needs, expressed as a capability, not an implementation (no screens/tables/endpoints).
- **so that** — the **why**: motivation and business value. Most often dropped, and the part that lets you challenge, reprioritize, or drop the story.

The template forces What and Why while keeping How open — exactly what makes a story **Negotiable**. A story whose "I want" names a mechanism ("I want a Redis cache") has smuggled the How into the What. The template is a **conversation starter, not a mandatory format**: no method prescribes a fixed story format, and a one-line note can be a fine story when context is shared.

## Acceptance criteria

The Confirmation from the 3 Cs (see [three-cs.md](three-cs.md)) written down. Good coverage spans four categories, not just the first:

- **Happy path** — the main success scenario.
- **Boundaries** — empty, zero, minimum, maximum, first, last.
- **Errors / invalid cases** — rejected input, failed dependency, unmet precondition.
- **UI changes** — visible state changes: what the user now sees, what becomes enabled or hidden.

**Sufficiency:** criteria are enough only when every rule has at least one confirmable (testable) criterion across the four categories; a rule with no confirmable criterion is not ready.

Criteria CAN be expressed as Given-When-Then (owned by **bdd-guide**); designing the tests themselves — levels, doubles, structure — is owned by **testing-guide**. This skill stops at "criteria exist and cover the four categories".

## Example

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

- The Card and Confirmation the template and criteria realize — [three-cs.md](three-cs.md).
- Given-When-Then for automation — **bdd-guide**; test design behind the criteria — **testing-guide**.
