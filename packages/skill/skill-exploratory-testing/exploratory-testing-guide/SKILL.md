---
name: exploratory-testing-guide
description: "Use when designing, running, debriefing, or reviewing focused exploratory software-testing sessions where learning, test design, and execution happen together. Triggers on test charters, session-based testing, risk tours, heuristics, test oracles, exploratory regression, unknown failure modes, session notes, coverage maps, debriefs, or deciding what targeted exploration a release candidate still needs, even when the user doesn't say exploratory testing."
---

# Exploratory Testing

Explore important uncertainty deliberately, capture reproducible evidence, and report
what was and was not covered without replacing automated checks or scripted acceptance
tests.

## Essentials

- **Charter one risk mission** — define target, risk, question, scope, environment,
  data, time box, and useful evidence before starting
- **Learn, design, and execute together** — adapt the next test from observed behavior
  while retaining a traceable session record
- **Vary meaningful dimensions** — explore data, sequence, state, interfaces,
  interruptions, permissions, resources, concurrency, recovery, and environment
- **Use explicit oracles** — compare behavior with requirements, user expectations,
  invariants, comparable products, prior versions, domain rules, or consistency
- **Capture reproducible evidence** — record exact build, setup, actions, observations,
  logs, screenshots, defects, questions, and coverage notes
- **Debrief the session** — separate findings, test coverage, new risks, blocked areas,
  follow-up charters, and confidence limits

## Gotchas

- Unscripted clicking is not exploratory testing; the charter and learning loop make
  the work purposeful.
- A time box bounds effort, not curiosity. Stop safely, preserve state, and create a
  follow-up charter for material unfinished exploration.
- Finding no defect does not prove quality; report tested dimensions and remaining risk.
- A defect count does not measure session value; one clarified high-impact risk can be
  more useful than many cosmetic findings.
- Automation protects known expectations. Exploration searches for important behavior
  and interactions not yet represented by those expectations.

## Example

```text
Charter: Explore checkout recovery after payment-provider interruption.
Target: rc-42 · staging-eu · saved-card and new-card paths.
Risk: A resumed checkout may double-charge or lose the order.
Vary: interruption point, retry count, browser refresh, webhook order, account state.
Oracles: at-most-one charge, order/payment consistency, actionable user state.
Evidence: timeline, request ids, provider events, observed UI, database state.
Debrief: one reproducible duplicate authorization; refund path and offline resume untested.
```

## Progressive Disclosure

- Read [references/charters-and-sessions.md](references/charters-and-sessions.md) - Load when defining a charter, preparing a session, time-boxing work, or capturing notes
- Read [references/heuristics-and-oracles.md](references/heuristics-and-oracles.md) - Load when selecting variations, finding useful test ideas, or deciding how to judge observed behavior
- Read [references/debrief-and-evidence.md](references/debrief-and-evidence.md) - Load when reporting findings, coverage, confidence, blocked exploration, or follow-up work
