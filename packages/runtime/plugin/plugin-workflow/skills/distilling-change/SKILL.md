---
name: distilling-change
description: "Use when a change is finished and its work should become reusable skills: an umbrella map skill and granular skills that replay the change on a clean base branch. Reads the change directory, git history and session logs. Triggers on a replay playbook, or making a finished upgrade repeatable, even when the user doesn't say 'distill'."
---

# Distill a Finished Change into Skills

Purpose: turn a finished change into an umbrella map skill plus granular skills that replay the work, as a controlled and incremental replay onto a clean base branch. The original branch is an oracle, never something to merge whole. The output is the suite of skill files; the change directory stays unchanged except for the log entry and the indexes.

**Load first:** **recording-handoff**, and run its entry step before anything below.

## Essentials

- The change directory is the record: `plan.md`, `decisions/`, `work/`, `implementation/`, `verdicts/` and `close-out.md`. Read it fully yourself before any delegation.
- Git wins on what changed, the session logs on why and on the mechanism of a failure, and the record on intent and decisions. When two sources disagree, investigate.
- Each fact of the suite carries what, why, order, temporary or final, failure modes with mechanisms, and validation commands.
- The suite works with the branch deleted and the change directory archived. Branch names, commit ids and the run narrative live only in provenance files.
- **authoring-skills** owns the form of each skill the suite holds: names, descriptions, references and validation.
- Never commit or stage. Never start another step.

## Procedure

1. **Run the entry step** of **recording-handoff**. The suite is written to a skills directory, so confirm the branch of its repository. Preconditions: `plan.md` and `close-out.md`. A missing `close-out.md` means the change is not finished: report it and stop, and offer `validating-work-package`.
2. **Scope the sources, cheap first.** Establish the branch, the merge base and the drift of the base. Read the change directory in this order: `close-out.md`, `plan.md`, the decisions, the work items, the implementation notes and the verdicts. Then find the session logs of each harness that ran the work, the CI runs and the review threads. For each document the work produced, `git cat-file -e <base>:<path>` decides whether it is a worked example or an output. Read [references/sources.md](references/sources.md) when you mine a source.
3. **Reconstruct by parallel fan-out.** Start one agent for each tier of the work (foundation, the repeated pattern, cleanup and late fixes) and one log miner for what the record omits: failed and superseded approaches, reverts, changes of plan, steering by the operator. Require each agent to report each piece with the facts of the Essentials.
4. **Write the map before the suite.** As references of the umbrella: the phase map with the dependency graph and the team split; the invariants and the inventory of temporary scaffolding with removal criteria; the safety model that classes each change by what protects code that is not migrated yet; pitfalls as symptom, mechanism and fix; the register of open decisions; the process lessons; and a replay order that improves on the original where the evidence supports it.
5. **Write the suite.** The umbrella is `<topic>-map`, a router and not a runbook. Each granular skill has a gerund name, a boundary taken from the reconstructed work, prerequisites, checks that detect the current state, the recipe with its essential code, failure modes, temporary versus final, and deterministic validation. Read [references/templates.md](references/templates.md) when you write the files, and follow **authoring-skills** for their form.
6. **Verify mechanically.** Frontmatter names match their directories; commit ids and branch names appear only in provenance files; each cross-reference resolves; each validation command runs; a reader of the umbrella alone can tell which phase the tree is in and what to do first. Plant one violation to prove that each sweep can fail before you trust its green result.
7. **Record the run.** Write a log entry with the path of the suite and regenerate the indexes with the script of **recording-handoff**, then stop.

## When not to use

- Checking whether the change met its criteria: use `validating-work-package`.
- Folding the lessons of one session into a rule or a skill: use **authoring-instructions** or **authoring-skills**.

## Gotchas

- Content that exists only at a commit, such as an example folded away later, is not in the final tree: capture it in the suite, or a squash merge loses it.
- A document that the work created is an output with one owning skill ("create if absent"), never an input.
- A "missing infrastructure" blocker in the record is a hypothesis until someone checks what the library already ships.
- A close-out marked `partial` leaves open criteria: the suite replays only what was met and carries the rest as open decisions.

## Done when

- The suite passes the validation of **authoring-skills**, and each sweep of step 6 was proved able to fail.
- The change directory is unchanged except for one log entry and the regenerated indexes.

## Progressive Disclosure

- Read [references/sources.md](references/sources.md) - Load when mining a source: git history, harness session logs, or CI and review records.
- Read [references/templates.md](references/templates.md) - Load when writing the files of the suite: layout, granular skill skeleton, provenance skeleton.
