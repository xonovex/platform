# user-stories: Splitting Stories with SPIDR

SPIDR is a compact, five-technique starter set for splitting a story that is too big into smaller stories that each still deliver value. The letters stand for Spike, Path, Interface, Data, Rules. Reach for SPIDR first; reach for the fuller decision tree in [splitting-flowchart.md](splitting-flowchart.md) when SPIDR doesn't yield a clean split.

## The five techniques

**S — Spike** — When the story is too poorly understood to split or build, do a small piece of research first to gain the knowledge. A spike is the **last resort** and is research-only: it produces knowledge, not shippable value, so a spike is not itself a vertically-sliced story. Once the spike resolves the uncertainty, split the real story with one of the other techniques.

**P — Path** — Split by the different paths a user can take through the story. Implement one path end to end first, then add the others as separate stories.

**I — Interface** — Deliver a simpler interface first, then richer variants. Ship the plainest usable version (one input method, a basic screen), then enrich it in later stories. The "I" is **Interface**, singular.

**D — Data** — Support only a subset of the data at first. Handle one currency, one region, or one format initially; add the rest as follow-up stories.

**R — Rules** — Temporarily relax a business rule to reduce the first slice's complexity. Build the simplest rule set that is still valuable, then reintroduce the deferred rules as their own stories.

Rationale: each technique carves off a thin vertical slice that a user can still use, rather than a horizontal layer. Every resulting story should still pass INVEST (see [invest-and-smart.md](invest-and-smart.md)) — especially **Valuable** and **Small**.

## Worked examples (one per technique)

Parent story: "As a rider, I want to pay for my completed trip so that I can leave without handling cash."

| Technique     | First slice                                                                  | Deferred to later stories                |
| ------------- | ---------------------------------------------------------------------------- | ---------------------------------------- |
| **Spike**     | research how the payment gateway handles idempotent retries (knowledge only) | the actual charge story, once understood |
| **Path**      | the saved-card path                                                          | wallet path; cash-tip path               |
| **Interface** | a plain "Paid" confirmation screen                                           | an animated, itemized receipt            |
| **Data**      | charges in one currency                                                      | multi-currency and FX                    |
| **Rules**     | auto-charge the saved card only                                              | split-fare and surge-pricing rules       |

## BAD → GOOD

BAD (horizontal split — components, not slices):

```
Story 1: build the payments API
Story 2: build the payments database
Story 3: build the payments UI
```

None of these delivers value alone; the rider can pay only after all three ship.

GOOD (SPIDR vertical split):

```
Story 1 (Rules + Data): auto-charge the saved card, one currency  ← shippable value
Story 2 (Path):         add the wallet payment path
Story 3 (Rules):        add split-fare
Story 4 (Interface):    add the animated itemized receipt
```

## SPIDR vs the flowchart

SPIDR (five techniques) is **not** the same as the larger splitting-pattern flowchart. They are complementary: SPIDR is a quick starter kit; the flowchart in [splitting-flowchart.md](splitting-flowchart.md) is the fuller decision tree with preconditions and evaluation rules. "Spike" appears in both, and in both it is the last-resort, research-only move.

## Cross-references

- The full pattern flowchart and vertical-slicing / walking-skeleton concept — [splitting-flowchart.md](splitting-flowchart.md).
- The INVEST properties every split story must still pass — [invest-and-smart.md](invest-and-smart.md).
