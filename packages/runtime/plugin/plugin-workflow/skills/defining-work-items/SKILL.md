---
name: defining-work-items
description: "Use when the draft work outline of a change must become complete work item definitions: each ready row gets observable criteria with their verification, a repository, real dependencies, the decisions that apply and the skills to consult, and a work package groups the items when the plan calls for one. Proposes the cut in the reply first (package, items, teams, repositories, criteria, dependencies) and waits for approval before writing anything. Triggers on 'define the work items', 'complete the outline', 'cut the work', 'group the work in a package', 'add acceptance criteria to the work files', even when the user doesn't say 'work item'. Not for writing the plan or its first outline, settling a decision, or implementing."
---

# Define Work Items

Purpose: turn the draft rows of the work outline into work items that a team can start, finish and verify without asking what was meant, and, when the plan calls for one, a work package that groups them. The output is edited work files, a proposed or existing work package, and a reply that shows the cut before anything is written.

**Load first:** **recording-handoff**, and run its entry step before anything below.

## Essentials

- Propose the cut in the reply, then stop. Nothing is written before the operator approves it.
- Work only on the rows of the operator's team. Rows of other teams stay as they are unless the operator asks. The `owners` of the plan and the `team` of a work file use one vocabulary, a team name, so the operator's team is written the same way in both.
- A row that waits on an open decision stays `waiting` with `waits_on`, and is defined only as far as its outline. Do not invent the criteria of work whose shape an open decision still decides.
- Every criterion of a work item is a condition a check can fail, and says how it is checked.
- `depends_on` names real blockers only. A decision is never a dependency: it goes in `waits_on`.
- Edit work files in place. A work file is never renamed or deleted. Keep unknown frontmatter keys.
- Definitions stay `status: draft` until the operator confirms the written files. Only then they become `stable` with a `verified` entry that names the operator as `human:<id>`. Approving the cut and confirming the definitions are two separate steps: a confirmation given before the files were shown, such as in the same message as the approval, does not count.
- A work package is optional. A plan whose items need no shared parent has no package, and that is a valid result: the proposal says so and why.
- Every timestamp comes from the system clock (`date -u +%Y-%m-%dT%H:%M:%SZ`), and every random four-character suffix uses only `a-z` and `2-7`.
- Never edit the plan, never commit or stage, never start another step.

## Procedure

1. **Run the entry step** of **recording-handoff**. The work files go into the plan directory's repository, so confirm that branch: until the answer arrives, do only read-only work, and an unattended run stops after it and reports the question. Preconditions: `plan.md` exists, and the `work/` directory holds at least one outline row. If either is missing, report which step normally writes it and stop.
2. **Load the skills that apply** from the sources the settings reference of **recording-handoff** lists: the `skills_to_consult` of the plan and of each work file, a custom workflow skill the project provides, the **defining-work-items** entry of the project's `skills` map, and the defaults for this phase. When no source names a skill, load nothing extra and say so in one line.
3. **Read the inputs yourself.** Read `plan.md`, every decision that is not deprecated (grouped by `state` from `decisions/index.md`), and every work file. Find the operator's team; when the invocation does not say, ask once. Sort the rows of that team into the ones to define (`ready`) and the ones that wait. Read the work package the operator names, if any; without one, the plan may have no package at all. See [references/outline-rows.md](references/outline-rows.md).
4. **Draft the cut** and send it as the reply from [assets/templates/cut-proposal.md](assets/templates/cut-proposal.md): the work package (or the line that says there is none, and why), the work items, their teams, repositories, criteria with verification, dependencies, decisions and skills. See [references/cut-rules.md](references/cut-rules.md), [references/writing-criteria.md](references/writing-criteria.md) and [references/dependencies-and-package.md](references/dependencies-and-package.md). Stop and wait.
5. **Take the operator's answer.** Apply requested changes to the proposal and send it again. Write only after a clear approval of the cut. A silence, a question or an approval of something else is not an approval. Do not take a confirmation of the definitions here: the files do not exist yet.
6. **Write the files**, announcing each write in one line (the path and what changed). Attach the items to the named package, or write a new package file, or write the items with no `parent` when the cut has no package. See [references/existing-package.md](references/existing-package.md). Name a new file `<slug>-<suffix>`. Set `generated.by` to `defining-work-items/<version>`, with the version read as the frontmatter reference of **recording-handoff** says (never guessed), and set `generated.at` from the system clock on every file you change.
7. **Check.** Run `index --write` and then `validate` for the bundle. Fix every finding in your files. Confirm that the dependency graph is acyclic, that every ready row of the team has criteria with verification, a repository and its dependencies, and that no waiting row gained criteria beyond its outline.
8. **Ask for confirmation** of the written definitions, listing the files, and wait. A confirmation counts only when the operator gives it after the files were shown or listed; one given in advance, in the same message as the approval of the cut, does not count, and the files stay `draft`. The `human:<id>` comes from the invocation; ask for it when it is missing, never invent one. The confirmation is a separate write with its own log entry (`action: Update`), and the earlier entry is not edited. Set `status: stable` and `verified: {by: human:<id>, at: <timestamp>}` on the confirmed files and update `generated.at`, both timestamps from the system clock.
9. **Record.** Write one `Workflow Log Entry` from the template of **recording-handoff**, but only when the run changed the bundle (a run that stopped after the proposal writes no entry). Set `skill: defining-work-items`, `action: Creation` when the run created a concept file other than the log entry itself and `Update` otherwise, and `touched` to the concept files the run created or changed (never the log entry itself; omit the key when empty). Run `index --write` and `log --write`, then `validate`.
10. **Report** the outcome first: the items defined, the package, the rows left as outline and the decision each waits on, the rows of other teams left alone, and anything that needs **revising-plan** (a waiting row whose decisions have since settled, the plan's generated block that lists a new package). Offer the next steps: **critiquing-plan** for a critique, or **implementing-work-item** for a ready item. Stop.

## When not to use

- The plan or its first outline does not exist: use **writing-plan**.
- A decision is open and the operator wants it settled: use **settling-questions**.
- The plan or the readiness of a row must change because a decision settled: use **revising-plan**.
- Carrying out a work item: use **implementing-work-item**.
- A user story outside a change directory: use **writing-user-stories**.

## Gotchas

- Writing before the approval, or treating "looks fine" about the plan as an approval of the cut.
- Counting a confirmation of the definitions given in advance, in the message that approves the cut. The files stay `draft` until the operator confirms them after seeing them.
- Forcing a package onto a cut that needs none. No package is a valid result; say so in the proposal.
- Taking a branch confirmation from permission to write files. Only a named branch, or "the current one", confirms it.
- Cutting by platform or layer. Group by outcome and discipline; see [references/cut-rules.md](references/cut-rules.md).
- Defining a waiting row in full. The open decision may change its shape; leave it at its outline and name what it waits on.
- A criterion that restates the objective, or that nobody can fail ("works as expected"). State the condition and the check.
- A criterion with no verification, or a verification that is only "test it". Name the command, the request, the screenshot or the review question.
- A `depends_on` that expresses ordering preference, a shared reader or an open decision. List only blockers whose work must be done first.
- Redefining a criterion id that another file may cite. Keep `c1`, `c2` stable and add new ids at the end.
- Setting `parent` to a work item or to a missing file. The parent is the concept ID of a file with `kind: work-package`.
- Giving a package criteria, or listing its items in it. The items point to the package through `parent`.
- Touching another team's rows to make the graph tidy. Report the gap instead.
- Writing `verified` for a definition nobody confirmed, or inventing a person id.
- Editing the plan to make the outcome look tidier. Report what needs the other steps instead.

## Done when

- Every ready row of the operator's team is a complete work item: criteria with verification, a repository from the project's map, dependencies, applicable decisions.
- Every waiting row still names its open decisions in `waits_on` and has no more than its outline.
- The dependency graph is acyclic and the script's `validate` reports no error and no schema-not-found warning for the work files.
- Every attached item names the package in `parent` (or the cut has no package and the items have no `parent`), and a new package has `kind: work-package`, a team and no criteria.
- The files are `draft` unless the operator confirmed them after seeing them, in which case they are `stable` with a `verified` entry.
- The plan and the decisions are unchanged, and a run that changed the bundle wrote one log entry.

## Progressive Disclosure

- Read [references/cut-rules.md](references/cut-rules.md) - Load when choosing the items, the teams and the repositories, or when sizing an item.
- Read [references/outline-rows.md](references/outline-rows.md) - Load when deciding whose rows belong to the cut, which rows stay as outline, or when splitting or merging rows.
- Read [references/writing-criteria.md](references/writing-criteria.md) - Load when writing or reviewing a criterion and its verification.
- Read [references/dependencies-and-package.md](references/dependencies-and-package.md) - Load when setting `depends_on`, `decisions` and `skills_to_consult`, checking the graph, or proposing a new work package.
- Read [references/existing-package.md](references/existing-package.md) - Load when the invoker names an existing work package to attach the items to.
- Read [assets/templates/cut-proposal.md](assets/templates/cut-proposal.md) - Load when writing the proposal reply. The work files use the shared work item template of **recording-handoff**; this skill has no schema of its own.
