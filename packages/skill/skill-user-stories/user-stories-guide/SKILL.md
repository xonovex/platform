---
name: user-stories-guide
description: "Use when writing, evaluating, splitting, or refining user stories — applying INVEST, the 3 Cs (Card / Conversation / Confirmation), the 'As a / I want / so that' template, writing acceptance criteria, slicing vertically into a walking skeleton, splitting with SPIDR or the splitting-pattern flowchart, and refining the backlog into ready items. Triggers on user story, INVEST, story splitting, SPIDR, vertical slice, walking skeleton, acceptance criteria, 'As a... I want... so that...', 3 Cs, backlog refinement / grooming, SMART tasks — even when the user doesn't say 'user story'. Skip the Given-When-Then notation and executable specs (see **bdd-guide**), test design and levels (see **testing-guide**), domain modelling and ubiquitous language (see **ddd-guide**), the FDD feature list (see **fdd-guide**), and coupling theory behind 'Independent' (see **connascence-guide**)."
---

# User Stories

A user story is a placeholder for a conversation about user value, not a specification. Write it small, talk it through, confirm it with acceptance criteria, and slice it vertically so every story ships usable value end to end.

## Essentials

- **Check every story against INVEST** - Independent, Negotiable, Valuable, Estimable, Small, Testable, see [references/invest-and-smart.md](references/invest-and-smart.md)
- **Treat the story as the 3 Cs** - a Card invites a Conversation, confirmed by a Confirmation test, see [references/three-cs.md](references/three-cs.md)
- **Write who / what / why, then acceptance criteria** - "As a / I want / so that" plus happy-path, boundary, error and UI checks, see [references/template-and-acceptance-criteria.md](references/template-and-acceptance-criteria.md)
- **Split a too-big story vertically** - SPIDR for a quick set, the splitting flowchart for the full decision tree, see [references/splitting-spidr.md](references/splitting-spidr.md) and [references/splitting-flowchart.md](references/splitting-flowchart.md)
- **Refine the backlog continuously** - break items down until each is doable within one sprint and ready, see [references/backlog-refinement.md](references/backlog-refinement.md)

## Gotchas

- INVEST "S" is **Small** (a few person-weeks at most, finishable within a sprint), not "Sized appropriately".
- A walking skeleton is **production code with tests** linking the main components end to end — not a throwaway prototype or a research spike.
- Never split horizontally ("build the DB", "build the API", "build the UI"); those are tasks/components. Every story is a thin **vertical** slice.
- A Spike yields **knowledge, not shippable value**, so it is a last-resort move, not itself a vertically-sliced story.
- The template is a conversation starter, not a mandatory format — no method prescribes a fixed story format.
- Acceptance criteria are not just the happy path; cover boundaries, error/invalid cases, and UI changes. They can later be written as Given-When-Then for automation — that notation is owned by **bdd-guide**.
- Backlog refinement is an **ongoing** activity, not a fixed weekly "grooming meeting"; sizing is the Developers' responsibility.

## Example

```
As a rider, I want to pay for my completed trip
  so that I can leave the cab without handling cash.

INVEST: Valuable (rider), Negotiable (method TBD), Testable. Too big → split.
Split (SPIDR): Rules (auto-charge saved card; defer split-fare/surge)
               Data (one currency first) · Path (card, then wallet, then cash-tip)
               Interface (plain "paid" screen first; animated receipt later)

First thin vertical slice (walking skeleton):
  request trip → match driver → charge saved card → show "paid"
  (touches dispatch + payment + UI end to end)

Acceptance criteria (the Confirmation):
  happy   valid card charged, receipt shown
  boundary zero-distance trip = minimum fare
  error   declined card blocks trip end, prompts retry
```

## Progressive Disclosure

- Read [references/invest-and-smart.md](references/invest-and-smart.md) - Load when evaluating a story against INVEST or breaking it into SMART tasks
- Read [references/three-cs.md](references/three-cs.md) - Load when explaining the Card / Conversation / Confirmation model of a story
- Read [references/template-and-acceptance-criteria.md](references/template-and-acceptance-criteria.md) - Load when writing the story sentence or its acceptance criteria
- Read [references/splitting-spidr.md](references/splitting-spidr.md) - Load when splitting a story with the five SPIDR techniques
- Read [references/splitting-flowchart.md](references/splitting-flowchart.md) - Load when splitting with the full pattern flowchart or slicing vertically
- Read [references/backlog-refinement.md](references/backlog-refinement.md) - Load when refining the backlog and deciding when an item is ready
- Read [references/best-practices-appendix.md](references/best-practices-appendix.md) - Load when you want the whole user-story checklist on one page
