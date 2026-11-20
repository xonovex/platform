# Editing the work files and the plan

## Work files

A work file is edited in place when a decision it waits on or applies changes.

- A decision the work waits on settled: remove it from `waits_on`, add the new decision to `decisions`, and when `waits_on` is empty set `readiness: ready` and delete the key.
- A decision reopened, paused or only partly settled: replace the old ID in `waits_on`, or in `decisions`, with the new one in `waits_on`, and set `readiness: waiting`. A partly settled decision keeps the work waiting unless the settled part is all the work needs, which the rationale then states.
- The settled decision changes what the work must deliver: edit the objective, the criteria and their verification. Keep criterion ids stable; add `c<next>` for a new one.
- `depends_on` changes when the revision moves work between files or drops a dependency.
- A revision that removes a work file's scope: set `status: deprecated`, add a line to its `# Notes` that names the run's log entry, and take it out of other files' `depends_on`. The file is not deleted, because other files and the log refer to its concept ID.

## The plan

Edit `plan.md` only outside the block between `<!-- generated:begin -->` and `<!-- generated:end -->`.

- `# Objective` is a picture of the settled state only. A settled approach reads as a flat statement of the future, without the rejected reasoning, which lives in the decision file.
- A mechanism nobody agreed to does not appear. A rejected sub-design is removed together with every success criterion and scope line that mentions it.
- An open decision is linked, not summarized. Point the sentence at the decision that owns it and remove the assumption it used to make.
- `# Scope` keeps both lists. `# Success criteria` stay observable by a stated means, never an implementation detail.

## Every edit

- Follow the write rules of **recording-handoff**, which keep unknown frontmatter keys.
- Update `generated.at` and set `generated.by` to `revising-plan/<version>` on a plan or work file whose content changed. The deprecation of an old decision leaves `generated` as it was.
- A plan or work file that was confirmed by a person (`status: stable` or a `verified` entry) and changed by this run returns to `status: draft` with the `verified` entry removed. The invoker confirms it again.
