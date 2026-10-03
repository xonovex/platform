# Followup: Close Out a Plan

Assemble the follow-up record for one explicit completed, paused, or handed-over plan and return
it inline. The operation changes nothing: no files, no plan-state mutation. Persistence belongs to
the caller; plan seeds persist through Create, decision entries settle through Decide.

Run it while context is fresh: at the last landed commit, at an external gate, or at handover.
The record outlives its plan as the densest input for whoever later mines the work (an audit, a
successor plan, or Distill). Section skeletons: read
[followup-templates.md](followup-templates.md) when assembling any section.

## Core Workflow

1. Report where the record should live. Usually the plan's own storage travels and there is
   nothing to flag; when it does not, recommend a durable destination per section as part of the
   record. Recommend only.
2. Status against the success criteria, three buckets per criterion: met AND evidenced (name the
   evidence, re-derived now), gated (name the gate), or open (name the owner). Never a flat
   "done".
3. Blockers and gates, classified by kind: human or external gates no code can satisfy (with the
   exact steps, who performs each, and which step fails silently if skipped), design gates (with
   the sharpest known assessment and any workable design, so the next attempt does not start from
   zero), and residue needing a person or a device. Draw the gate chain; state which tracks are
   independent. Record what is no longer a blocker, with the closing evidence.
4. The open-decision register: per unmade decision, the question, why it is open, options with
   costs, a recommendation, the owner, what waits on it, and a literal `Call: (unrecorded)` line.
   Capture decisions already made but never written as recorded calls instead of reopening them.
5. Open points and residue: sweeps assigned and dropped, coverage thinner than its headline
   number, defects found and deliberately not fixed (with the owning team), behaviour changes
   shipped without sign-off, records already drifting, and the questions the plan opened and
   closed so none is re-asked.
6. Follow-up plan seeds, one per coherent chunk of remaining work: objective, evidenced current
   state, what is genuinely open, known traps, a success-criteria sketch, validation, what gates
   it, and sequencing (start now, wait, independent). Flag work previously attempted and not
   landed: that is the signal it needs planning rather than improvisation. Name the
   `skills_to_consult` the future plan should carry; hand seeds to Create for persistence.
7. The review brief, when the work still faces review: what this is, a tiered read order, the
   rules learned during the work restated as review criteria, and "things that look like mistakes
   and are not", each with the reason invisible in the diff.
8. The retrospective, about the method rather than the code: the most important finding first,
   then keep and change, each stated as an actionable rule.

## Gotchas

- Closing out is descriptive, not authorizing: the record neither approves a merge nor gates the
  follow-up plans; owners and gates do.
- Evidence over assertion: re-derive counts; report disagreeing numbers instead of picking one.
- Mechanism over symptom: a symptom-level blocker record misdirects the next attempt; restate
  object-shaped blockers as what the code actually reads.
- Every open item ends with an owner, a register entry, or an explicit acceptance; "noted" is
  none of those. Sweep the record for orphan phrases ("should", "someone", "later").
- Corrections sit next to the original claim as addenda, not rewrites.
- A cold reader must be able to answer from the record alone: what is done, what is blocked and
  by what, who decides what, and what to do first.
