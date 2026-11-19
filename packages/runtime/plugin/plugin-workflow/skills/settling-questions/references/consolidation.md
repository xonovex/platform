# Consolidation

Consolidate in the reply when the queue is empty or the operator stops. The decision files are already written, so there is nothing to save. The consolidation reports; it writes no plan and no work file. Use [../assets/templates/consolidation.md](../assets/templates/consolidation.md).

## Counts

A line with the number of questions in each state, then a table of every question with its state, its decision file and its owner when open, paused or partly settled. Questions the operator did not reach are counted as unanswered and listed with the reason, so nobody mistakes them for settled. The counts come from the files you wrote this run plus the earlier decisions the run touched; recount from the files, not from memory.

## Principles

The rules the answers revealed, each in one sentence and each tied to the decisions it came from. Include the rules that hold under every option; they are what the plan can state flat. A principle that contradicts a settled decision is an error to fix before the reply goes out.

## The list for the owners

For each open, paused or partly settled decision: the owner, the question in one sentence, the candidates, the operator's phasing, what is excluded, and the rule that holds whichever wins. For a paused decision, the fact needed and who gathers it. The operator takes this list to the owners and brings their answers back to a new run of this skill, so it must stand on its own.

## What can proceed without waiting

The work that does not depend on any open or paused decision, named in plain words. When work files exist, name them by title and list which ones wait on a decision from this run. Do not change them; **revising-plan** does.

## The order of work

A sequence of the settled decisions and the open ones, showing what blocks what: for example "the identifier decision first, then the backfill and the gating in either order, the phasing last".

## Corrections still to make

Pages, plans or earlier research the answers make wrong. Name each page and what it says now against what the decisions say. Include the `FIXME` notes that belong in the plan. Offer the correction as a next step; this skill changes none.

## Closing the run

1. Write one `Workflow Log Entry` from the template of **recording-handoff**: `skill: settling-questions`, `action: Creation`, and every created and deprecated decision in `touched`. The body is one sentence for a reader who sees only the generated log line. A run that recorded nothing writes no entry.
2. Regenerate the indexes and the log with the script of **recording-handoff**, then validate the bundle and report every finding.
3. Keep sensitive findings in the reply and out of every file. Say in one line that you did.
4. Offer the next steps: **revising-plan** for the plan and the work files from the new decisions, **writing-plan** when no plan exists, another run of this skill when the owners answer, or **researching-change** for a paused fact. Never start one.

## Checks

- The counts match the files.
- Every open decision names its owner and its options.
- No decision recorded as settled contradicts a principle.
- The reply leads with the counts and the direction, and each section stays short.
