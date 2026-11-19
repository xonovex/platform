# Step brief template

Copy the block into the prompt of one step agent, replace every angle-bracket placeholder and delete any comment lines. Write one brief for each step of each item. The agent has only this brief, so name every path it needs. The model and the effort belong to the call that starts the agent, not to the brief.

```markdown
# Step: <step skill> for <work-id>

Load the skill **<step skill>** and follow it. It runs the entry step of **recording-handoff** first.

## Change

- Change directory: <plans/change>, in the main checkout at <path> of repository <name>
- Work item: work/<slug>-<suffix> (file: <path relative to the repository root>)
- Package run: item <n> of <m>

## Branch

- The invoker confirmed the branch <branch> of the main checkout, in these words: '<words of the invoker>'.
- Base of the worktree: <the confirmed branch, or the branch of the last committed item, work/<slug>-<suffix>, which this package run builds on>.
- Worktree: <path and branch, when it exists>

<!-- Include "Carried facts" only for implementing-work-item. Give the validation agent nothing from the implementer. -->

## Carried facts

- <What an earlier item of this package changed that this item builds on, such as a renamed type, a new interface or a moved convention, with the path>

## Rules

- This run is unattended. When you need an answer, stop after the read-only work and report the question.
- Do not commit, stage or push, and do not start another workflow step.
- Leave the worktree's copy of `plans/` unedited.
- Filter long command output to failures and summary lines, and read a file by range after a search.

## Required return

- Status: done, stopped or blocked, with the reason.
- The path of each file you wrote.
- <For implementing-work-item: the worktree path, the branch and each criterion's self-check state.>
- <For validating-work-item: the counts by result and the first 12 characters of the tree hash.>
- <For preparing-delivery: the subject line and the result of the scan.>
- Each question that needs the invoker.
```

## Fill in

- `Step` and the skill line: one of **implementing-work-item**, **validating-work-item** or **preparing-delivery**.
- `Work item`: the concept ID and the path of the file in the main checkout, because the agent's checkout may not resolve a bare concept ID.
- `Branch`: forward the confirmation in the invoker's words, in single quotes. Never write a confirmation the invoker did not give.
- `Carried facts`: take them from the implementation notes of the items this run committed. Leave the section out for the validation and the delivery agents.
- `Required return`: keep only the line of the step this brief is for.
