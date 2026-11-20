# Splitting and ordering tasks

The supervisor splits the work item into tasks and orders them before any worker starts. [delegation.md](delegation.md) says how each task is carried out.

## Splitting the work item into tasks

Read the objective, the criteria, the decisions the work item lists and the code they touch. Split by concern: one task owns one coherent change that a person could review on its own, such as a validation rule, a data mapping or a screen. Tasks are not split by file count or by layer for its own sake.

Give every task an `id` (kebab-case), a `concern` and a one-sentence `summary`. Every criterion is served by at least one task, and every task serves at least one criterion. A task that serves none is scope creep: drop it or ask.

## Ordering the tasks

- Tasks that share a file, or share a check whose result one of them changes, run in sequence, in the order the later one needs the earlier one's result.
- Tasks that touch disjoint files and checks run together, within a bounded number at once.
- Check the file sets, not the task names. Two tasks called "frontend" and "backend" may still both edit a generated file.
- Run a task that other tasks build on first, and pass its result into their briefs.

## When something does not fit

- A criterion that cannot be met, or a decision that looks wrong once the code is open: stop that task, say what you found, and leave the choice to the invoker. A settled decision is a constraint, not a suggestion; a question the decisions do not answer goes to **settling-questions**, never decided here.
- A change the work item does not call for, even a good one: report it and leave it out.
