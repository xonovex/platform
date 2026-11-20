# The entry step

Every workflow step runs these four moves before its own work and reports the outcome in a few lines. The step then continues, stops or waits for the invoker, as set out below.

## 1. Confirm the branch

For each repository the step writes to, run the detect commands and report what they show:

```bash
git branch --show-current
git status --short | head -20
```

Name the branch and list the uncommitted files. Uncommitted edits in the plan directory are the work the step builds on, so say so. Proceed on the branch the invoker confirms, even when it is not the default branch.

The branch is confirmed when the invocation names it or says to work on the current one. Authorization to write files is not a branch confirmation. Otherwise ask in the first reply, and until the answer arrives do only read-only work (reading files, running the validator without `--write`, read-only commands); every write waits for the answer. An unattended run, which nobody can answer, stops after the read-only work and reports the question; it never assumes.

A repository the step only reads needs no confirmation. The plan directory's repository counts as a repository the step writes to whenever the step creates or edits a file there. When the plan directory and the work repository are the same repository, confirm once, for the branch of the main checkout; the branch of a work item's worktree is created by the skill from the confirmed base branch and needs no second confirmation.

## 2. Validate what the step reads

Validate the bundle, or the files the step reads, and report every finding with its path and line:

```bash
uv run scripts/plan-bundle.py validate plans/<change>
```

Run the command from the skill directory or give the script's full path. A finding in a file the step does not read is still worth one line, but it does not block the step. A finding in a file the step reads is reported before any work, and the step asks whether to continue when the finding could change its result. [validator.md](validator.md) describes the findings and what to do when the script is not available.

## 3. Read the indexes

Read the root `index.md`, then the index of each directory the step needs, such as `decisions/index.md` for a step that settles a question. The indexes show what exists and, for decisions and work, the grouping that matters: settled or open, ready or waiting.

An index that is stale or missing is regenerated in the working tree:

```bash
uv run scripts/plan-bundle.py index plans/<change> --write
uv run scripts/plan-bundle.py log plans/<change> --write
```

Say in the report that the step regenerated it. Never edit a generated file by hand.

## 4. Check the step's own preconditions

Each step names the files it needs, for example a plan with at least one decision. Check that they exist and that the document the step needs has a status it can work with. If a file is missing:

- Report which file is missing and which kind of step normally writes it, in words.
- Stop. Do not write a stand-in, do not assume an earlier step ran, and do not start another step.

Settings are preconditions too. A step that applies something outward, such as pushing a branch or opening a pull request, needs the explicit settings in [settings.md](settings.md). A missing explicit setting is a blocker: report which one and stop.

When no source (plan or work item `skills_to_consult`, a custom workflow skill, the project's `skills` map) names a skill for the phase, the step loads nothing extra and says so in one line.

## What the entry step does not do

- It does not compare files with repository heads. There is no drift detection, no revision counter and no stamp on copies.
- It does not commit or stage anything.
- It does not write a log entry by itself. The log entry belongs to the run that changes the bundle, at its end.

## Example report

- Branch: `feature/fare-capping` in the shop-app repository, with two uncommitted files: `plans/fare-capping/plan.md` and `plans/fare-capping/decisions/index.md`. Is this the branch to work on? Until you answer I only read; every write waits.
- Validation of `plans/fare-capping`: 0 errors, 1 warning (`work/pricing-api-7k2f.md:1: schema file workflow/work-item/2 not found`). The decisions index was stale, so I regenerated it.
- Precondition: the step needs at least one decision with state `open`. There is none, so I stop here.
