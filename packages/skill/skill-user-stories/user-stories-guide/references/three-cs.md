# user-stories: The 3 Cs — Card, Conversation, Confirmation

A user story has three components, not one. The written sentence is only the first. Treating the card as the whole story is the most common way stories fail: the value lives in the conversation, and "done" lives in the confirmation.

## The three components

**Card** — a short written token, deliberately small, that summarizes the intent and invites a discussion. It is a promise to have a conversation, not the specification itself. The card's small size is a deliberate constraint: if the story does not fit on the card, it is too big for the sprint and must be split. The card carries just enough — usually the who/what/why sentence — to remember the conversation and to prioritize.

**Conversation** — the verbal discussion, over time, where the real value is co-created. Details that don't fit (and shouldn't fit) on the card are worked out here between the people who want the feature and the people who build it, supplemented by documents, sketches, and tests as needed. This is where **Negotiable** from INVEST is exercised; see [invest-and-smart.md](invest-and-smart.md).

**Confirmation** — the acceptance test that confirms the story is done to the customer's satisfaction. The conversation produces concrete examples; those examples become the acceptance criteria and ultimately the pass/fail check. The Confirmation **is** the acceptance test — it is the same idea as INVEST's **Testable** property, viewed from the card-and-conversation angle rather than the checklist angle.

Rationale: the three Cs keep teams from treating a story as a frozen requirements document. A card too detailed to negotiate has thrown away the conversation; a card with no confirmation has no definition of done. Each C reinforces an INVEST letter — Card↔Small, Conversation↔Negotiable, Confirmation↔Testable.

## How the 3 Cs map onto the rest of the craft

| C            | Produced artifact                             | Owner reference                                                            |
| ------------ | --------------------------------------------- | -------------------------------------------------------------------------- |
| Card         | the "As a / I want / so that" sentence        | [template-and-acceptance-criteria.md](template-and-acceptance-criteria.md) |
| Conversation | shared understanding, split decisions         | [splitting-flowchart.md](splitting-flowchart.md)                           |
| Confirmation | acceptance criteria (happy/boundary/error/UI) | [template-and-acceptance-criteria.md](template-and-acceptance-criteria.md) |

The acceptance criteria that come out of the Confirmation can be expressed as Given-When-Then examples for automation, but that notation and the collaborative discovery practice around it are owned by **bdd-guide** — this skill only names the handoff.

## BAD → GOOD

BAD (card swallows the conversation, no confirmation):

```
Card: As a rider I want to pay. The pay button must be #2E7D32 green, 48px tall,
      bottom-right, call POST /v3/charge with idempotency key, retry 3x with
      exponential backoff, show spinner... [continues for 2 pages]
```

The detail pre-empts the conversation and still names no acceptance check.

GOOD (small card, confirmation captured separately):

```
Card:         As a rider, I want to pay for my completed trip
              so that I can leave without handling cash.
Conversation: card first; wallet and cash-tip later; one currency to start.
Confirmation: valid card charged + receipt shown; declined card prompts retry;
              zero-distance trip charges the minimum fare.
```

## Cross-references

- The template and acceptance-criteria coverage the Card and Confirmation produce — [template-and-acceptance-criteria.md](template-and-acceptance-criteria.md).
- Given-When-Then as a way to express the Confirmation for automation — **bdd-guide**.
