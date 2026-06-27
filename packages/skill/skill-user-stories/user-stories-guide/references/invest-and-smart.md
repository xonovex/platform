# user-stories: INVEST Stories and SMART Tasks

INVEST is a six-property checklist for the quality of a single story; SMART is its task-level companion for when a ready story is broken into engineering tasks. A story that fails an INVEST letter is a signal to negotiate, split, or clarify — not to ship as-is.

## The six INVEST properties

| Letter | Property        | Means                                                                | Neutral example                                                                        |
| ------ | --------------- | -------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| I      | **Independent** | Stories don't overlap and can be scheduled and built in any order    | "Apply a discount code at checkout" doesn't depend on "Save a card for later"          |
| N      | **Negotiable**  | Not a fixed contract; the details are co-created in conversation     | The card says "pay for a trip"; whether wallet or card comes first is decided together |
| V      | **Valuable**    | Delivers value the customer can see                                  | A rider paying without cash is valuable; "add a payments table" is not                 |
| E      | **Estimable**   | Understood well enough to estimate and so to rank and schedule       | The team can size "charge the saved card"; "integrate payments somehow" they cannot    |
| S      | **Small**       | At most a few person-weeks, small enough to finish within one sprint | "Charge the saved card for a completed trip" fits a sprint                             |
| T      | **Testable**    | Understood well enough to write a test confirming it is done         | "Declined card prompts a retry" can be confirmed; "make payments delightful" cannot    |

Rationale: INVEST gives a fast, shared vocabulary for why a story isn't ready. Most letters push toward the same outcome — a small, valuable, end-to-end slice the team can estimate and confirm. When a story is too big (fails **Small**) or you cannot estimate it (fails **Estimable**), split it; see [splitting-spidr.md](splitting-spidr.md) and [splitting-flowchart.md](splitting-flowchart.md). The coupling theory behind **Independent** — what makes two stories truly entangled — is owned by **connascence-guide**; INVEST only asks for schedulable independence.

The canonical "S" is **Small**, not "Sized appropriately"; the canonical bound is "a few person-weeks at most", which in practice means small enough to complete inside a single sprint.

## SMART tasks

Once a story is ready and the Developers decompose it into tasks, the tasks should be SMART:

- **Specific** — one clear thing to do, not a vague area
- **Measurable** — you can tell when it is done
- **Achievable** — doable as scoped
- **Relevant** — contributes to the story's value
- **Time-boxed** — bounded to a short, predictable span (hours, not weeks)

INVEST governs the story (the unit of customer value); SMART governs the tasks beneath it (the units of engineering work). They are different acronyms for different altitudes — do not apply INVEST to tasks or SMART to stories.

## BAD → GOOD

BAD (fails Independent, Valuable, Small):

```
As a developer, I want to build the payments database schema
  so that we can store charges later.
```

This is a horizontal component with no customer value, entangled with every later payment story, and unbounded in scope.

GOOD (passes INVEST):

```
As a rider, I want my completed trip auto-charged to my saved card
  so that I can step out of the cab without paying cash.
```

Valuable to the rider, negotiable on method, estimable, small enough for a sprint, and testable via acceptance criteria.

## Folklore to avoid

- "S = Sized appropriately" — narrowing; the canonical expansion is **Small**.
- SMART as "Assignable / Agreed-upon / Realistic" — those circulate but the canonical task-level expansion is **Achievable** ("A") and **Relevant** ("R").

## Cross-references

- The 3 Cs view of the same story (Card / Conversation / Confirmation) — [three-cs.md](three-cs.md).
- Why **Independent** is about scheduling, not coupling theory — **connascence-guide**.
