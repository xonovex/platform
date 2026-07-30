# user-stories: The 3 Cs: Card, Conversation, Confirmation

A user story has three components, not one. Treating the card as the whole story is the most common way stories fail: value lives in the conversation, "done" lives in the confirmation. Each C reinforces an INVEST letter: Card↔Small, Conversation↔Negotiable, Confirmation↔Testable (see [invest-and-smart.md](invest-and-smart.md)).

- **Card**: a short written token summarizing intent, deliberately small: if the story does not fit on the card, it is too big for the sprint and must be split. It carries just the who/what/why sentence: a promise to have a conversation, not the specification.
- **Conversation**: the verbal discussion over time where the real value is co-created between the people who want the feature and those who build it, supplemented by sketches and tests. This exercises **Negotiable**.
- **Confirmation**: the acceptance test that confirms the story is done. The Confirmation **is** the acceptance test: the same idea as INVEST **Testable**. Concrete examples from the conversation become the acceptance criteria.

| C            | Produced artifact                             | Owner reference                                                            |
| ------------ | --------------------------------------------- | -------------------------------------------------------------------------- |
| Card         | the "As a / I want / so that" sentence        | [template-and-acceptance-criteria.md](template-and-acceptance-criteria.md) |
| Conversation | shared understanding, split decisions         | [splitting-flowchart.md](splitting-flowchart.md)                           |
| Confirmation | acceptance criteria (happy/boundary/error/UI) | [template-and-acceptance-criteria.md](template-and-acceptance-criteria.md) |

Acceptance criteria from the Confirmation can be expressed as Given-When-Then for automation, but that notation is owned by **bdd-guide**. This skill only names the handoff.

## Cross-references

- The template and acceptance-criteria the Card and Confirmation produce: [template-and-acceptance-criteria.md](template-and-acceptance-criteria.md).
- Given-When-Then to express the Confirmation for automation: **bdd-guide**.
