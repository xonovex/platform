# Status of each success criterion

## Contents

- The three states
- Re-deriving evidence
- The result

## The three states

Every success criterion of the plan gets exactly one state.

| State | Meaning | The entry must name |
| --- | --- | --- |
| met | The criterion holds now, shown by evidence gathered in this run | The check, its output and the revision |
| gated | The criterion cannot hold until something no work item can supply happens | The gate and who passes it |
| open | The criterion does not hold and no gate explains it | The owner and the work that is missing |

A flat "done" is not a state. "Done pending the release" is gated, "done except one platform" is two criteria or an open one, and "implemented" is not evidence of anything.

A criterion is judged where the plan says it must hold, which is the confirmed base branch unless the plan names another place. A change that exists only in an uncommitted worktree, on an unmerged branch or in an open pull request does not hold there yet. The criterion is then `gated`, the gate is that the change has not landed on the base branch, and the entry names who lands it. Evidence gathered in the worktree still goes in the entry, labelled with the tree it came from, so the reader sees both what works and where it works.

## Re-deriving evidence

Evidence is re-derived when this run ran the check or read the file, and the entry records what it saw.

- Run the check the plan's criterion names, or the work item's `verification` that covers it. Record the command, the result and the revision checked: the commit of each repository and, for an uncommitted worktree, its tree hash.
- Count again rather than quote a count from a document. When two sources give different numbers, report both and say which you measured.
- A reproducible check beats a reading. Where only reading is possible, cite the file and the line.
- A check that cannot run here, such as one that needs a device, a secret or a remote system, leaves the criterion `open` or `gated`, with the limit stated. Never promote it to met on a verdict alone.
- A verdict or a prepared delivery is a lead with its own rules: see [verdicts-and-deliveries.md](verdicts-and-deliveries.md). The checks across work items are in [integration.md](integration.md).
- A status field in a plan or a work file is neither a criterion nor evidence unless the plan defines it as one.

## The result

- `met`: every entry of `criteria` is met.
- `not-met`: no entry is met.
- `partly-met`: anything else.

A gated criterion is not met. The result never rounds up.
