# Task brief template

Copy the block into the prompt of one worker, replace every angle-bracket placeholder and delete any comment lines. Write one brief for each task. The worker has only this brief, so name every file, rule and command it needs.

```markdown
# Task: <task id>

## Concern

<The one concern this task owns, and what done looks like, in two or three sentences.>

## Work item

<Title and objective of the work item, so the task is read in context.>

## Read first

- <Path in the worktree>: <why>
- <Instruction file of each package touched>
- <Path of a finished task's file this one builds on, and what it changed>

## You may change

- <Paths or areas this task owns>

<!-- Include "Do not change" when a neighbouring task or a shared file is off limits. -->

## Do not change

- <Path or area owned by another task, and who owns it>

## Rules that apply

- <Decision concept ID> (file: `<path relative to the repository root, such as plans/<change>/decisions/<name>.md>`): <the decision restated as a concrete constraint, such as the unit a value carries, a version, which module may see which type>
- <Skill to load before writing code>
- <Convention of the repository that bears on this task>

## Criteria this task serves

- `<c1>`: <criterion text, verbatim>. Verification: <verification, verbatim>

## Checks to run

- <Command, run from the worktree>

## Working agreement

- Work only in <worktree path>, on branch <branch>.
- Do not commit, stage, push or create a branch.
- Leave changes that are not yours as they are, and leave the worktree's copy of `plans/` unedited.
- If the task cannot be done as written, stop and say why. Do not choose a different design.

## Required return

- Status: done, partly done or blocked.
- Files changed, one path per line.
- For each criterion served: met or not met, and the command or observation behind it.
- Each check run, with its result.
- Each place a decision shaped the code, as the decision concept ID and `path:line`, and one sentence on how.
- Deviations from this brief, with the reason.
```

## Fill in

- `Concern`: one concern for one task. Two concerns in one brief make the report impossible to check.
- `Read first`, `You may change`, `Do not change`: name the files. A worker that has to search for them wastes effort and drifts outside its task.
- `Rules that apply`: restate each decision concretely and keep its concept ID, so the return can cite it, and give the path of its file relative to the repository root, because a worker's checkout may not resolve a bare concept ID. Never cite a decision by its ID alone. The path is the main checkout's bundle; the worktree's tracked copy of `plans/` is read-only for the worker.
- `Criteria this task serves`: copy the text and the verification word for word.
- `Checks to run`: the commands the verification names, plus the repository's format, lint and type checks for the files changed.
- Model and effort belong to the call that starts the worker, not to the brief.
