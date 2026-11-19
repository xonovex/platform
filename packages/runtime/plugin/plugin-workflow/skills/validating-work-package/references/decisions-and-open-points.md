# Unsettled decisions and open points

## Contents

- The decision register
- Calls made and never written
- Open points and residue
- Every item ends somewhere

## The decision register

The register covers every decision the package still depends on and that is not settled: decisions whose `state` is `open`, `paused` or `partly-settled`, and questions the work raised that have no decision file. Settled decisions are not listed. For each entry:

- The question, as a question.
- Why it is open and what makes it undecidable mechanically.
- The options, each with what it costs and what it buys.
- A recommendation: one option and the reason.
- The owner.
- What waits on it: the work files, seeds or criteria.
- A literal line `Call: (unrecorded)`. The decision is recorded through **settling-questions**, never here.

The register describes. It does not decide, and it does not write a decision file.

## Calls made and never written

When the work shows a decision was made in practice and never recorded, for example the code settles a choice or a note says "we agreed", list it as a call to record. Do not reopen it, and do not record it here. Name the evidence and hand it to **settling-questions**.

## Open points and residue

Everything left open that is not a blocker and not a register entry:

- Sweeps assigned and dropped: work someone owned and no one finished.
- Coverage thinner than its headline: what a green count hides, row by row.
- Defects found and deliberately not fixed, each with the team that owns it.
- Behavior changes shipped without sign-off.
- Deliveries prepared and not applied, each with its directory and the commit or publication it still needs.
- Records already drifting: disagreeing counts, stale statuses, content that does not travel with the change.
- Questions the plan opened and closed, so none is asked again.

## Every item ends somewhere

An open item ends in one of three places: an owner, a register entry or an explicit acceptance by a named person. "Noted", "should", "someone" and "later" are none of them. Before writing, scan the record for these orphan words and give each its end.

Corrections to an earlier claim sit next to the claim as an addendum, not as a silent rewrite.
