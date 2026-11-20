---
name: implementing-work-item
description: "Use when one defined, ready work item is to be built: it picks up the work file, checks that it can start, splits the work into tasks by concern, delegates each task to a worker with a bounded and sequenced fan-out (or works through them in order when the harness cannot start one), verifies the result itself, runs the checks the criteria name and saves implementation notes with the decisions that shaped the code. Changes stay uncommitted in the work item's own worktree and branch. Triggers on 'implement this work item', 'build the story', 'carry out the work file', 'pick up the next ready item', even when the user doesn't say 'work item'. Not for defining the item, validating it, writing the commit or pull request, or pushing the branch."
---

# Implement One Work Item

Purpose: turn one ready work item into code in its own worktree and record how, so the work can be validated independently and delivered. The output is uncommitted changes in the worktree and one notes file in the change's plan directory. The run commits nothing and pushes nothing.

**Load first:** the **recording-handoff** skill, and run its entry step before anything below.

## Essentials

- One work item at a time. The work happens in that item's own worktree and branch in its work repository; changes stay uncommitted until delivery.
- An unmet precondition is reported and the run stops. Never start another step to satisfy it.
- Split by concern into tasks. Tasks that share a file or a check run in sequence, the others together, within a bounded number at once.
- Set the model and the effort of every worker: the cheapest that does the task reliably, the most capable only for a design decision or a hard fault. A worker never commits.
- The supervisor verifies the work itself. A worker's report is not proof.
- Decisions are constraints, not instructions or evidence. When a decision shaped code the code cannot explain, record a decision note with its location.
- Everything this step concludes is a self-check, never a validation.
- No outward effect: no commit, stage, push, branch on a remote or pull request. Authority comes only from the invoker.
- A worktree made from the base branch also holds a tracked copy of `plans/`. The notes and the log entry go to the bundle of the main checkout, never to the worktree's copy, and the worktree's copy of `plans/` is never edited.
- Every timestamp comes from the system clock (`date -u +%Y-%m-%dT%H:%M:%SZ`), and every random four-character suffix uses only `a-z` and `2-7`.

## Procedure

1. **Run the entry step** of **recording-handoff**. Two things are written: the notes in the plan directory's repository and the code in the work repository. Confirm the branch of each as the entry step says, and name the uncommitted files of each repository. When the plan directory and the work repository are the same repository, confirm once, for the branch of the main checkout: the worktree's branch is created from the confirmed base branch and needs no second confirmation. Find or create the worktree as [references/worktree-and-branch.md](references/worktree-and-branch.md) says.
2. **Check the preconditions** in [references/preconditions.md](references/preconditions.md): the work file with verifiable criteria, `readiness: ready`, finished blockers, the repository in the project's map, the settings. Report an unmet one and stop.
3. **Load the skills that apply** from the skills map that the settings reference of **recording-handoff** defines, for the phase **implementing-work-item**: the `skills_to_consult` of the plan and of the work item, a custom workflow skill the project provides, the **implementing-work-item** entry of the project's `skills` map, and the default skills of the phase when the project's map does not name it.
4. **Read the ground.** Read the work file, the decisions it lists, the plan's success criteria and the code the work touches. Run the repository's checks once before editing so a later failure can be told from an old one.
5. **Split into tasks** by concern and order them. See [references/tasks.md](references/tasks.md).
6. **Carry out the tasks.** Delegate each to a worker with a brief from [assets/templates/task-brief.md](assets/templates/task-brief.md) when the harness can start one; otherwise (for example when a worker cannot start workers of its own) work through them in order and say so. Review each result before the next batch. See [references/delegation.md](references/delegation.md).
7. **Verify yourself.** Re-run the checks the criteria name, read the diff against the criteria and open the artifact behind any visible result. See [references/verification.md](references/verification.md).
8. **Write the notes** as `implementation/<work-id>/notes.md` in the main checkout's bundle, from [assets/templates/implementation-notes.md](assets/templates/implementation-notes.md), with a `Workflow Log Entry` and regenerated indexes. Announce the write in one line. See [references/notes-file.md](references/notes-file.md) and [references/closing-the-run.md](references/closing-the-run.md).
9. **Stop** and report inline, outcome first. Offer **validating-work-item** and then **preparing-delivery**. Never start another step.

## When not to use

- Writing the work item's criteria, repository or dependencies: use **defining-work-items**.
- Checking whether the work meets its criteria: use **validating-work-item**. This step only self-checks.
- Writing the commit message or the pull request description: use **preparing-delivery**. Committing: use **using-git**.
- Pushing the branch or opening a pull request: use **using-github** or **using-gitlab** when the user asks.
- A change with no work file, or a question about what to build: use **writing-plan** or **settling-questions**.

## Gotchas

- Starting despite an unmet precondition because "the rest is clear". The run stops and says which precondition.
- Running two tasks together that both edit one file or one generated output. Compare file sets, not task names.
- Doing a task yourself because it is small. The supervisor's context is what the next brief needs.
- Trusting a worker's report. Re-run the checks and read the diff.
- Choosing a design when a decision does not cover the case. Stop and ask; **settling-questions** settles it.
- Committing or staging, even "to keep a checkpoint". The worktree holds the change until delivery.
- Editing the work file to match what was built. Definitions change through the plan.
- Writing a decision note with a location from before the final edit, or with a range. Check each `path:line` against the final files.
- Wording a self-check as validated, or marking a criterion met whose verification did not run.
- Reusing the worktree of another work item, or switching the main clone's branch.
- Writing the notes or the log entry into the worktree's tracked copy of `plans/`, or editing that copy. The bundle of the main checkout is the only one this skill writes.
- Briefing a worker with a bare decision concept ID. A worker's checkout may not resolve it; give the file path relative to the repository root as well.

## Done when

- The notes file validates with the **recording-handoff** script and has no schema-not-found warning.
- Every criterion has a row in the self-check, with evidence you re-ran or the reason it could not run.
- Every `decision_notes` location resolves to a line in the worktree.
- Nothing is committed or staged, in the worktree or in the plan directory's repository, and the work file and the plan are unchanged.

## Progressive Disclosure

- Read [references/preconditions.md](references/preconditions.md) - Load when checking whether the work item can start, or reporting why it cannot.
- Read [references/worktree-and-branch.md](references/worktree-and-branch.md) - Load when finding, creating or naming the worktree and the branch.
- Read [references/tasks.md](references/tasks.md) - Load when splitting the work item into tasks, ordering them, or when a criterion or a decision does not fit.
- Read [references/delegation.md](references/delegation.md) - Load when briefing a worker, working without one, or reviewing a report.
- Read [references/verification.md](references/verification.md) - Load when running the checks, judging a result or labelling the self-check.
- Read [references/notes-file.md](references/notes-file.md) - Load when filling the keys of the notes file, recording a decision note or listing the checks.
- Read [references/closing-the-run.md](references/closing-the-run.md) - Load when writing the notes file and its log entry, reporting, or rerunning on the same item.
- Read [assets/templates/implementation-notes.md](assets/templates/implementation-notes.md) - Load when writing the notes file.
- Read [assets/templates/task-brief.md](assets/templates/task-brief.md) - Load when briefing a worker.
- Read [assets/schemas/implementation-notes.schema.json](assets/schemas/implementation-notes.schema.json) - Load when the validator reports a schema finding on a notes file.
