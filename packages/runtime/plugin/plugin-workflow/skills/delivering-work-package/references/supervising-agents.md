# Supervising the step agents

## The model and the effort

Set both in the call that starts the agent, never in the brief. Choose the cheapest that does the step with no loss of quality, and the most capable for a hard fault. A model or an effort that the invoker names wins.

| Step | Usual choice | Why |
| --- | --- | --- |
| **implementing-work-item** | A mid tier, medium effort; higher for an item with several concerns | The work item and its decisions carry the design |
| **validating-work-item** | A mid tier, medium effort | It judges whether the evidence reaches each criterion |
| **preparing-delivery** | The cheapest tier that writes to the project's conventions, low effort | The files already hold the facts |
| A rerun after a hard fault | The most capable tier, high effort, one time | The fault needs more judgement |

Never choose a more capable model to redo a thin result. Brief the same agent again with the missing part named.

## The brief

A step agent starts with no context, so write its brief from [../assets/templates/step-brief.md](../assets/templates/step-brief.md), from your own reading of the files.

- Forward the invoker's branch confirmation in the invoker's words, and name the base of the worktree. Never confirm a branch the invoker did not confirm.
- Say that the run is unattended: an agent that needs an answer stops after the read-only work and reports the question.
- For **implementing-work-item**, carry the facts that earlier items of the package changed and this item builds on, from their implementation notes, so the agent inherits them and does not find them again.
- For **validating-work-item**, give nothing from the implementer: no reply, no notes, no carried facts.

## Verifying the work of each agent

The reply is a lead, never proof: read the files and run the Git commands yourself.

- **After implementing-work-item:** the notes file exists and validates, its worktree and branch exist, `git diff --stat HEAD <hash>` shows a change inside the item, the worktree's copy of `plans/` is unchanged, nothing is committed or staged, and the self-check has a row for each criterion. A self-check is not a validation.
- **After validating-work-item:** the verdict exists and validates, its `source` is `independent`, it has exactly one result for each criterion id of the work file, each `pass` and `fail` names a command or an anchor, and its `worktree_hash` equals the tree hash you compute. Read each `pass` against the criterion's text: a green test proves only what it asserts.
- **After preparing-delivery:** `commit-message.md` exists and validates, its `work` and `branch` match the item, its body starts with a subject line in the project's convention, the agent's scan found nothing, and the tree hash has not moved.

Distrust editor and stale-index diagnostics; settle a doubt with a real command.

## When an agent fails

- **A finding in its work:** send the finding to the same agent. A new agent repeats the discovery and can bring the finding back.
- **It stops on a precondition or a question:** the item stops. Record the question for the invoker. Never answer a design question yourself; **settling-questions** owns it.
- **A harness fault** (no reply, a timeout, a lost connection): start the step again in a new agent, because each step reads its inputs from files. On a second fault of the same kind, start it once on the most capable tier with high effort, then stop the item.
- **It commits, stages, pushes or edits outside its step:** stop the item and report what it did. Reset or delete nothing without the invoker's word.
- **It reports done and the files disagree:** the files win. Send the difference to the same agent.

Never do the step yourself, and never edit the code, the notes, a verdict or a delivery file.
