# user-stories: INVEST Stories and SMART Tasks

INVEST checks one story; SMART checks the tasks a ready story is broken into. A story that fails an INVEST letter is a signal to negotiate, split, or clarify — not to ship.

## INVEST (the story)

- **I — Independent** — schedulable and buildable in any order. This is scheduling independence, not coupling theory (owned by **connascence-guide**).
- **N — Negotiable** — not a fixed contract; details co-created in conversation.
- **V — Valuable** — value the customer can see (not "add a payments table").
- **E — Estimable** — understood well enough to size and rank.
- **S — Small** — a few person-weeks at most, finishable within one sprint. The canonical expansion is **Small**, not "Sized appropriately".
- **T — Testable** — a test can confirm it is done.

Fails **Small** or **Estimable** → split (see [splitting-spidr.md](splitting-spidr.md), [splitting-flowchart.md](splitting-flowchart.md)).

## SMART (the tasks beneath a story)

**Specific · Measurable · Achievable · Relevant · Time-boxed** (hours, not weeks). The canonical "A" is **Achievable** and "R" is **Relevant** — not "Assignable / Agreed-upon / Realistic", which circulate but are non-canonical.

INVEST governs the story (customer value); SMART governs its tasks (engineering work). Do not cross-apply them.

## Example

```
BAD  (fails Independent, Valuable, Small — a horizontal component):
  As a developer, I want to build the payments database schema
    so that we can store charges later.
GOOD (passes INVEST):
  As a rider, I want my completed trip auto-charged to my saved card
    so that I can step out of the cab without paying cash.
```

## Cross-references

- The 3 Cs view of the same story — [three-cs.md](three-cs.md).
- Coupling theory behind **Independent** — **connascence-guide**.
