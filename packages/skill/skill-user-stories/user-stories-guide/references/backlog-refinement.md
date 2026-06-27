# user-stories: Backlog Refinement

Backlog refinement is the ongoing act of breaking down and further defining backlog items into smaller, more precise items, adding description, order, and size. It is continuous work, not a fixed ceremony, and it is how stories become ready.

## What refinement is

Refinement is the act of breaking down and further defining backlog items into smaller, more precise items. It adds detail, order, and size to items as understanding grows. It is an **ongoing activity**, not a prescribed event on the calendar: there is no required "grooming meeting" and no fixed cap on how much time it takes.

Sizing is the responsibility of the **Developers** who will do the work — the people who build an item judge how big it is. The current wording is "size", a deliberate change from the older "estimate": items are sized, not necessarily estimated with numbers.

Refinement is collaborative. The product owner (PM) brings the value and order, the Developers bring feasibility and sizing, and QA brings the testability and edge cases that become acceptance criteria. The conversation is the 3 Cs' Conversation (see [three-cs.md](three-cs.md)) happening continuously across the backlog.

## When an item is ready

An item is **ready** when it is doable within one sprint. That is the practical upper bound on story size and the link back to INVEST's **Small** (see [invest-and-smart.md](invest-and-smart.md)): if an item cannot be finished inside a single sprint, refine it further — split it (see [splitting-flowchart.md](splitting-flowchart.md)) until each resulting item fits. Items that meet this bar are ready to be pulled into sprint planning.

Rationale: continuous refinement keeps a small buffer of ready, small, valuable items at the top of the backlog, so planning is fast and the team is never blocked on under-defined work. Treating refinement as a one-off weekly meeting with mandatory numeric estimation reintroduces the bottleneck refinement exists to remove.

## BAD → GOOD

BAD (one big estimation ceremony, item too big, no acceptance criteria):

```
Weekly grooming, 90 min, whole team estimates story points:
  "Payments" — 40 points — accepted into the backlog as-is
```

"Payments" is far larger than a sprint, carries no acceptance criteria, and the heavyweight ceremony is the only place refinement happens.

GOOD (continuous, sized by Developers, split until ready):

```
Refined over the week as understanding grew:
  "Auto-charge the saved card (one currency)"  ready: fits a sprint, criteria written
  "Add the wallet payment path"                ready
  "Add split-fare"                             needs more refinement — still too big
```

## Folklore to avoid

- "Refinement is a required weekly grooming meeting" — it is an ongoing activity, not a prescribed event.
- "Every item must carry a numeric estimate" — items are **sized** (the wording moved from "estimate" to "size"); the Developers own that sizing, and it is not strictly required to be a number.

## Cross-references

- The INVEST **Small** property and the sprint-sized bound — [invest-and-smart.md](invest-and-smart.md).
- Splitting an item that is still too big to be ready — [splitting-flowchart.md](splitting-flowchart.md).
- The Conversation that refinement is an instance of — [three-cs.md](three-cs.md).
