# Consolidation template

Copy the block into the final reply, replace every angle-bracket placeholder and delete any comment lines. Lead with the counts and the direction. Keep each section short.

```markdown
## Where the decisions stand

Settled: <n>. Partly settled: <n>. Paused: <n>. Open for their owners: <n>. Not reached: <n>.

| Question | State   | Decision file               | Owner or fact needed |
| -------- | ------- | --------------------------- | -------------------- |
| <title>  | <state> | `decisions/<slug>-<suffix>` | <owner or fact>      |

## Principles the answers revealed

- <Rule, in one sentence> (decisions: <titles>)

## For the owners

- <Question, in one sentence> Owner: <name>. Candidates: <letters and names>. Phasing: <the operator's phasing>. Excluded: <letter and reason>. Holds whichever wins: <rule>.

## What can proceed without waiting

- <Work that depends on no open or paused decision>

<!-- Include "Work files affected" when work files exist. -->

## Work files affected

- <Work file title>: waits on <decision>, now <settled or superseded>.

## Order of work

<Sequence showing what blocks what, in one or two lines.>

## Corrections still to make

- <Page or plan>: says <old>, the decisions say <new>.

## Next steps

<Revise the plan and the work files, write the plan, bring the owners' answers to a new run or research a paused fact. None has been started.>
```

## Fill in

- Counts come from the decision files, recounted. A question the operator did not reach is counted under "Not reached", never as settled.
- `For the owners`: one entry for each open, partly settled or paused decision. For a paused one, replace the candidates with the fact needed and who gathers it.
- `Principles`: include the rules that hold under every option. Check that none contradicts a settled decision.
- `Work files affected`: include only when work files exist; **revising-plan** updates them.
- `Corrections still to make`: name each page. Leave the section out when there are none and say so in the next-steps line.
- Say in one line that sensitive findings stayed in the reply and not in a file, when any existed.
