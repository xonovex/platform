# Delegating tasks

Each task from [tasks.md](tasks.md) goes to one worker, or to the supervisor when the harness cannot start one.

## Contents

- Delegating a task
- When the harness cannot start a worker
- Reviewing what returns

## Delegating a task

When the harness can start a worker (whatever delegated agent it provides), give each task to one. The supervisor stays the supervisor: it splits, briefs, reviews and records, and does not implement a task itself, because the context the next brief needs is spent by doing it.

- Write the brief from [../assets/templates/task-brief.md](../assets/templates/task-brief.md). It names the files the task reads and may change, the rules that apply, the criteria, the checks and the return. The worker starts with nothing else.
- Give every decision as the path of its file relative to the repository root (for example `plans/<change>/decisions/<name>.md`) as well as its concept ID. A worker's checkout may not resolve a bare concept ID, and the path lets it read the decision. The path is that of the main checkout's bundle: a worktree made from the base branch holds a tracked copy of `plans/`, and the notes and the log entry are written to the main checkout's bundle, never to the worktree's copy. The worktree's copy of `plans/` must not be edited, by the supervisor or by a worker.
- Set the model and the effort of every worker explicitly. Choose the cheapest that does the task reliably. Choose the most capable only for a design decision or a hard fault, never to redo a thin result: brief again with the missing part named.
- Run at most as many at once as the project's instructions allow, and at most four when they name no limit. When there are more tasks, start the next batch after reading the reports of the first.
- A worker never commits, stages, pushes, creates a branch or changes state outside its task.
- A worker that finds the task cannot be done as written stops and says why. It does not pick another design.

## When the harness cannot start a worker

A harness limit counts too: a harness whose workers cannot start workers of their own, or that provides no delegated agent, falls back to working through the tasks yourself, in the order of [tasks.md](tasks.md). Say so in the first line of the report. Keep the task boundaries: finish and check one task before the next, and write its `tasks` entry as you go. The checks and the notes are the same.

## Reviewing what returns

A report is a lead, not proof. For every task, before it counts:

- Read the diff of the files it changed against the task and the criteria.
- Re-run the checks it names yourself. See [verification.md](verification.md).
- Send a finding back to the same worker, because a fresh one repeats the discovery and can reintroduce the finding.
- Accept a fix to a defect the work item did not foresee only when it is at the root cause and tested, and record it as a deviation.
