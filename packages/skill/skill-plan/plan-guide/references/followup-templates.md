# Follow-Up Record Section Skeletons

One skeleton per section of the inline record. Small plans merge sections but keep every part.

## 1. Status

```markdown
# <Plan>: status as of <date>, <ref/commit>

## Met and evidenced

| Criterion   | Evidence                                                             |
| ----------- | -------------------------------------------------------------------- |
| <criterion> | <the grep / test run / measurement that proves it, re-derived today> |

## Gated

| Criterion | Gate | Where the gate is tracked |

## Open

| Criterion | Owner | Where |
```

## 2. Blockers

```markdown
# <Plan> blockers

<N> things stop this being finished. Only <M> are code, and they are not the ones that decide
the schedule.

## The gate chain

<diagram: which gates feed which; which tracks are independent and can run in parallel>

## Blocker 1: <name>

Owner: <who>. Kind: <human gate / design gate / residue>. Gates: <what waits on it>.
<For human gates: the exact steps, who performs each, before which milestone, and which step
fails SILENTLY if skipped.>
<For design gates: the sharpest assessment so far, the sub-blockers none of which tooling
catches, and any recorded workable design - so the next attempt starts from here, not zero.>

## What is no longer a blocker

<Closed risks with the evidence that closed them, so none gets re-litigated.>
```

## 3. Decisions

```markdown
# <Plan>: open decision register

Record the call in the Call line when it is made.

## Status

| # | Decision | Owner | Gates | Status |

## D<N>. <Question>?

<Why it is open; what makes it undecidable mechanically.>
| Option | Costs | Buys |
**Recommendation.** <one option, with the reason>
**Call:** _(unrecorded)_
```

## 4. Open points

```markdown
# <Plan> open points

Deliberately not the blockers list: everything else that was left open.

## Decisions nobody has made <- overflow small items not worth a register entry

## Sweeps with an owner and no pass <- assigned-then-dropped work, called out as such

## Coverage thinner than the headline <- what the headline number hides, row by row

## Defects found, not fixed, not ours <- each with its owning team

## Records already drifting <- disagreeing counts, stale statuses, non-travelling content

## Questions the plan opened and closed <- answered; listed so none is re-asked
```

## 5. Follow-up plan seeds

```markdown
# <Plan>: recommended follow-up plans

## Sequencing

| | Plan | Start | Gated on | Weight |
<which start now, which wait, which are mutually independent>

# A. <plan-name>

**Objective.** <one sentence>
**Why first / why it exists.** <including: was this attempted and not landed? say so - that is
the signal it wants planning rather than improvisation>

## What it unblocks

## Current state, evidenced

## Open questions for planning to settle

## Known traps

## Success criteria sketch

**Validation.** <commands>
```

## 6. Review brief

```markdown
# <Plan>: merge and review brief

## What this is <- one paragraph + a numbers table

## Why this is safer to review than <N> files suggests

## Read order <- tiers: read closely / read the pattern once / self-contained / late fixes

## What to check against <- the rules learned during the work, as review criteria

## Things that look like mistakes and are not <- each with the invisible-in-the-diff reason

## Validation <- what is green, with the caveats a reviewer should have

## Before this merges <- the silent-failure steps, named
```

## 7. Retrospective

```markdown
# <Plan>: retrospective

## The shape of it <- phases, durations, method (delegation? review cadence?)

## The finding <- the one thing that matters most, first

## What the process got right <- keep, with why

## What to do differently <- change, each stated as an actionable rule

## Numbers, for whoever wants them
```
