# discovery: Three Amigos and Example Mapping

Scenarios are the output of a conversation, not its substitute. Discovery is the first of BDD's three practices: a structured conversation over real, concrete examples that surfaces the rules and the gaps BEFORE any code is written. Skip it and the Gherkin you author alone just re-encodes one person's assumptions.

## The three amigos are three perspectives

The "three amigos" are three PERSPECTIVES on the work, not a head-count of three people:

- **Business / customer** — what problem are we solving, and what is valuable?
- **Development** — how might we build it; what is technically feasible?
- **Testing** — what could go wrong; which cases break the rule?

More people may join the conversation; fewer than these three perspectives and you lose either value, feasibility, or risk. The deliverable is shared understanding, captured as examples — the meeting itself is the mechanism.

## The discovery workshop

Pop one story from the backlog and talk through concrete examples in plain language for roughly 20-25 minutes. The story container and its INVEST qualities are owned by **user-stories-guide**; discovery takes that story as input. Close each story with a thumb vote: thumbs-up means the group understands it well enough to build; thumbs-down or sideways means it needs splitting or more questions answered first. A story that cannot be talked through in the time-box is too big or too uncertain to start.

## Example Mapping: four card colours

Example Mapping makes the conversation visible with four colours of card:

- **Yellow** — the user story under discussion (one per session, at the top).
- **Blue** — a business rule / acceptance criterion the story must satisfy.
- **Green** — a concrete example that illustrates a blue rule. Examples become scenarios.
- **Red** — an unanswered question or assumption that blocks agreement.

A well-sized story maps in about 25 minutes. Read the table as a sizing signal:

- A table FULL of blue cards means the story has too many rules — split it.
- MANY red cards mean too much uncertainty — the story is not ready; resolve the questions first.
- A healthy map has a few rules, each with a couple of green examples, and few reds.

```text
[YELLOW] Assign the best nearby taxi
  [BLUE] A closer taxi with a higher rating wins
    [GREEN] A 0.8-rated taxi at 1400m vs a 0.9-rated taxi at 1500m -> the 0.9 taxi
    [GREEN] Two taxis at the same distance -> the higher-rated one
  [BLUE] A taxi out of range is never assigned
    [GREEN] Nearest taxi is 9km away, max range 5km -> no assignment
  [RED] What happens when NO taxi is in range — queue, or reject?
```

## From map to scenarios

Each green card becomes a candidate scenario; each blue rule becomes a Gherkin `Rule` (see gherkin-reference). Turn 2-5 green examples per rule into scenarios — not one scenario per code path. The exhaustive cases are driven down into unit tests by **tdd-guide**; the step glue and any external stubbing is **testing-guide**. Why the practice frames discovery as testing understanding rather than testing software is in appendix-discovery.

```text
# BAD — skipped discovery, author writes Gherkin solo
One developer writes 30 scenarios covering every branch they imagined.
No tester challenged the edge cases; no business owner confirmed the rule.
The "what if no taxi is in range" question was never asked.

# GOOD — discovery first, scenarios second
A 20-minute three-amigos session maps the rules and one red question,
resolves the red, and yields 2-3 agreed green examples per rule that
the whole group thumbs-up before any are written as scenarios.
```
