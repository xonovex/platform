---
name: revising-plan
description: "Use when what is waiting must be folded into the plan and its work files: FIXME notes the operator left in the plan or other files, findings of a critique, or the corrections a later research file lists. Gives every item a disposition with a rationale, records each changed decision as a new decision file that supersedes the old one, updates the work files that wait on a changed decision, edits the plan outside its generated block, and saves the source of every note in the run's log entry. Triggers on 'apply my notes', 'process the critique', 'the new research corrects the plan', 'what changed for each note', even when the user doesn't say 'revise'. Not for writing the first plan, settling a question or implementing the plan."
---

# Fold Feedback into the Plan

Purpose: apply what is waiting to the plan and the work files so the operator sees what became of each item, the plan stays consistent with the decisions behind it, and nothing is left for a later pass. The run revises and never implements.

**Load first:** **recording-handoff**, and run its entry step before anything below.

## Essentials

- Three kinds of item wait: `FIXME` notes in the plan or other files, critique findings and research corrections. [references/inputs.md](references/inputs.md) says how to find each.
- Every item gets one disposition with a rationale: applied, partly applied, not applied with a reason, or sent to **settling-questions**. Nothing is dropped silently, and the original text of each note is kept in the run's log entry.
- An item that changes a decision becomes a new decision file that supersedes the old one, and the old file gets `status: deprecated` and `superseded_by` and nothing else. The operator's answer settles a decision; evidence only reopens one. A question that is new or still open goes to **settling-questions** and is never decided here.
- Work files are edited in place when the decisions they wait on change. `plan.md` is edited outside its generated block. Only this skill and **writing-plan** change `plan.md`.
- Never implement, commit or stage. Never start the next step.

## Procedure

1. **Run the entry step** of **recording-handoff**. The run writes to the plan directory's repository, so confirm that branch. Preconditions: `plan.md` exists and at least one item is waiting. When nothing waits, say so and stop.
2. **Enumerate every item** from the three inputs before editing anything, with its source: the file and line, or the concept ID and the finding number. See [references/inputs.md](references/inputs.md).
3. **Give each item a disposition** and a one-sentence rationale. A note that is a question gets its answer from the code or the design record, not a guess. A conflict the evidence cannot resolve is sent to **settling-questions**, never averaged away. See [references/dispositions.md](references/dispositions.md).
4. **Apply** in the order of [references/applying.md](references/applying.md): new decision files and the deprecation of the old ones, then the work files and `plan.md` as [references/work-and-plan.md](references/work-and-plan.md) says. Announce each write in one line.
5. **Record the run.** Write one `Workflow Log Entry` (`action: Update`, or `Creation` when the run wrote a decision) whose body is a one-sentence summary and the table from [assets/templates/dispositions.md](assets/templates/dispositions.md), then remove the resolved `FIXME` notes from the files they were in.
6. **Propagate and verify** with [references/propagation.md](references/propagation.md): search for old labels, regenerate the plan summary block and the indexes, validate the bundle, run the project's prose check on the files you edited.
7. **Report inline**, result first: the table of dispositions, what was not verified, what stays for other owners, and the next steps (**settling-questions** for the items sent on, **critiquing-plan** on the revised plan, **illustrating-plan** for a diagram the change made stale). Stop.

## When not to use

- A plan that does not exist yet: use **writing-plan**.
- A question with no recorded answer, or an answer the operator wants to walk through: use **settling-questions**.
- Cutting the work package or implementing work: use **defining-work-items** or **implementing-work-item**.

## Gotchas

- Editing a decision to change it. Write a new file with `supersedes`, then deprecate the old one.
- Settling a decision from a critique finding or a research file. Evidence can reopen a decision; only the operator's answer settles one.
- Dispositions that all say applied. A reworded, answered or sent-on note is not applied; use the full set.
- Removing a `FIXME` note before the log entry holds its text. The log entry is the only place the source survives.
- Deciding in the plan. A mechanism nobody agreed to does not appear, and an open decision is linked, not summarized.
- A work file left waiting on a decision that settled, or ready while it waits on one that reopened.
- A stale label left in the plan, a work file or a diagram after a decision changed. Search before reporting.
- Quoting a sensitive finding into the log entry. Leave it out of every file and describe it in the reply.
- Treating a prior approval as still valid for the changed text. A revised plan or work file returns to a draft.

## Done when

- Every enumerated item has a row in the dispositions table with a rationale, and the quoted text of each note is in it.
- Each changed decision is a new file, and each old one has only `status` and `superseded_by` added.
- No work file waits on a settled decision or is ready while it waits on an open one.
- `rg FIXME` finds no note this run resolved, the plan summary block and the indexes are current, and the bundle validates with no error.

## Progressive Disclosure

- Read [references/inputs.md](references/inputs.md) - Load when finding what is waiting, reading a note, a critique finding or a research correction, or deciding whether an item was already processed.
- Read [references/dispositions.md](references/dispositions.md) - Load when assigning a disposition, answering a question note, handling a conflict, or writing the table.
- Read [references/applying.md](references/applying.md) - Load when choosing the order of work or writing a superseding decision.
- Read [references/work-and-plan.md](references/work-and-plan.md) - Load when editing a work file's readiness or criteria, or the plan outside its generated block.
- Read [references/propagation.md](references/propagation.md) - Load when carrying a change to the artifacts that restate it, regenerating the summary block or verifying the run.
- Read [assets/templates/dispositions.md](assets/templates/dispositions.md) - Load when writing the log entry body with the dispositions table.
