# discovery: Three Amigos and Example Mapping

Scenarios are the output of a conversation, not its substitute. Skip discovery and the Gherkin you author alone just re-encodes one person's assumptions.

## Three perspectives, not three people

The "three amigos" are three PERSPECTIVES; invite more people, but never lose one of these:

- **Business / customer** — what problem, what is valuable?
- **Development** — how might we build it; what is feasible?
- **Testing** — what could go wrong; which cases break the rule?

## The workshop

Pop one story (its INVEST qualities owned by **user-stories-guide**) and talk through concrete examples in plain language for ~20-25 minutes. Close with a thumb vote: up = understood well enough to build; down/sideways = split it or answer questions first. A story that can't be talked through in the time-box is too big or too uncertain to start.

## Example Mapping — four card colours

- **Yellow** — the user story (one per session, at the top).
- **Blue** — a business rule / acceptance criterion.
- **Green** — a concrete example illustrating a blue rule; becomes a scenario.
- **Red** — an unanswered question / assumption blocking agreement.

Read the map as a sizing signal: FULL of blue = too many rules, split it; MANY red = too uncertain, not ready; a SINGLE rule with MANY greens = it hides several rules, tease them apart; healthy = a few rules, each with a couple of greens, few reds.

```text
[YELLOW] Assign the best nearby taxi
  [BLUE] A closer taxi with a higher rating wins
    [GREEN] 0.8-rated at 1400m vs 0.9-rated at 1500m -> the 0.9 taxi
  [BLUE] A taxi out of range is never assigned
    [GREEN] Nearest taxi 9km away, max range 5km -> no assignment
  [RED] No taxi in range — queue, or reject?
```

Each green card becomes a candidate scenario; each blue rule becomes a Gherkin `Rule` (see gherkin-reference). Turn each rule's greens into scenarios (often a couple, sometimes none) — not one per code path; exhaustive cases go to **tdd-guide**, step glue and stubbing to **testing-guide**. The framing of WHY discovery tests understanding rather than software is in appendix-discovery.
