# Lenses

## Contents

- Choosing lenses
- Red-team
- Pre-mortem
- Falsify
- Steelman

## Choosing lenses

The default is red-team plus pre-mortem. Add falsify when the plan makes quantitative or environmental claims, and steelman when the invoker wants fixes rather than faults. Record the lenses used in the `lenses` key. A lens that was not run is not listed.

Each lens is a question the reviewer asks of every section. A finding names the lens only through its content: a failure mode, the evidence for it, a direction.

## Red-team

Attack the essential assumptions and the unproven dependencies.

- What must be true for this plan to work, and where does the plan show that it is?
- Which dependency is assumed to exist, to be ready or to stay unchanged without evidence?
- Who or what can break the plan on purpose or by accident: a second producer, a stale cache, a team with a different goal?

Example finding: the plan assumes one process writes the queue, while a replay job also writes it, so a limit on the first protects nothing.

## Pre-mortem

Assume delivery failed and trace the likely causes back to planning choices.

- Imagine the change shipped and was rolled back. Which sentence of the plan made that likely?
- What was out of scope that turned out to matter, and what was in scope that nobody owned?
- Which work waits on a decision that has no owner or no date?

Example finding: the rollout has no recovery step, so the first bad batch of records has to be repaired by hand.

## Falsify

Find a concrete input, scale, sequence, environment or recovery case that the plan misses, and show it breaks a success criterion.

- Take each success criterion and construct the input that makes it false while the plan is followed.
- Try the empty case, the largest case, the second run, the interrupted run and the concurrent run.
- Name the environment that differs: another region, an older client, a slower disk.

Example finding: "no dropped rows" holds for one file, and a second import started while the first runs drops rows that arrive twice.

## Steelman

Find the smallest change that would strengthen a weak section. This lens produces directions, not faults.

- For each section that fails another lens, what is the least that would make it hold: one sentence in scope, one measurable criterion, one named owner?
- Prefer the change that keeps the plan's intent and its settled decisions.

Example finding: replace "performance is acceptable" with the p95 latency of the named request path and the load that produces it.
