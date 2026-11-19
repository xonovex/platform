# Workflow Plugin

Use `xonovex-workflow` to take a change from research to validated delivery as a series of standalone skills that keep everything they learn and decide in one directory of files. Record the project settings in the `Workflow` section of your project instructions (see [Record the project settings](#record-the-project-settings)), invoke `recording-handoff` once for the change, and continue with `researching-change` or `settling-questions`. Use the [Core Plugin](../plugin-core/README.md) for Git operations and the [Agentic Plugin](../plugin-agentic/README.md) to write the skills that `distilling-change` produces.

## How the steps work

Each step is a skill. Invoke it by name or describe the task in plain language. No step starts another: a step reads the files it needs, checks its preconditions, does its work, writes its files and stops. You, `delivering-work-package` or a script decide which step runs next.

Before its own work, every step runs the entry step of `recording-handoff`: it confirms the branch of each repository it writes to and names the uncommitted files, validates the files it reads, reads the generated indexes and checks that the files it needs exist. If a precondition is missing, the step reports it, names the step that normally satisfies it and stops. It never assumes that an earlier step ran, and it never runs that step itself. No step commits or stages anything, except `delivering-work-package`, which commits each validated work item locally and pushes nothing.

| Step | What it does and writes |
| --- | --- |
| `recording-handoff` | Owns the layout, the schemas, the entry step, the settings and the validator. Writes the generated files and the first log entry of a change. |
| `researching-change` | Researches the change with read-only explorers and writes one research file with the open questions. Holds the research of a system migration as a reference. |
| `settling-questions` | Settles open questions with you one at a time and writes each answer as a decision file. |
| `writing-plan` | Writes `plan.md` from the research and the decisions, and outlines draft work files. |
| `critiquing-plan` | Reviews the plan in a fresh read-only context and writes one critique file of graded findings. |
| `illustrating-plan` | Draws and renders the diagrams of a change into `diagrams/`, as SVG by default. |
| `revising-plan` | Folds `FIXME` notes, critique findings and research corrections into the plan and the work files, and records each changed decision as a new decision file. |
| `defining-work-items` | Proposes the cut and, after your approval, completes the work package and the work items. |
| `implementing-work-item` | Builds one ready work item in its own worktree and writes implementation notes. |
| `validating-work-item` | Checks one implemented work item against its criteria in a fresh read-only context and writes a verdict tied to the tree hash of the worktree. |
| `preparing-delivery` | Writes the commit message, the pull request description and the inline comments of one work item under `deliveries/`. It applies nothing. |
| `validating-work-package` | Checks the success criteria of the plan and the integration of the work package in a fresh read-only context, and writes the close-out. |
| `delivering-work-package` | Runs each ready work item through `implementing-work-item`, `validating-work-item` and `preparing-delivery` in dependency order, each in its own agent, and commits the item locally with `using-git` when the verdict covers the tree. It pushes nothing. |
| `distilling-change` | Turns a finished change into reusable skills from its directory, after `validating-work-package`. |

A push, a pull request and an inline comment go through `using-git`, `using-github` or `using-gitlab` when you ask for them. No step reaches a tracker or a host on its own.

## Typical order

Planning runs once for a change. Execution then repeats for each work item in dependency order, and the work package is validated after the last item.

![Planning from intent to work items, then implement, validate and deliver each work item, then validate the work package and distill the change](diagrams/change-flow.svg)

[Open the PNG](diagrams/change-flow.png) or [edit its Graphviz source](diagrams/change-flow.dot).

The order is a loop, not a line. Each step lists what must exist before it runs:

1. `researching-change` and `settling-questions` need only the change directory. `settling-questions` discovers questions when none are known.
2. `writing-plan` needs at least one research file or decision and no plan yet.
3. `critiquing-plan` needs `plan.md`. `illustrating-plan` needs `plan.md` and at least one decision.
4. `revising-plan` needs `plan.md` and something waiting: `FIXME` notes, critique findings or research corrections. Settle new questions and revise again until nothing is open.
5. `defining-work-items` needs `plan.md` and the draft work outline. Rows that wait on an open decision stay at their outline.
6. For each work item: `implementing-work-item` needs a work file whose criteria say how each is verified, `readiness: ready`, finished `depends_on` items and a repository in the settings. `validating-work-item` needs the worktree to hold a change. `preparing-delivery` needs the implementation notes and the `delivery` setting. `delivering-work-package` runs the three steps for each ready item.
7. `validating-work-package` needs `plan.md` with success criteria and the work items. A missing verdict does not stop it, but the criteria that the verdict would cover cannot be marked met.
8. `distilling-change` needs `close-out.md`.

Return to an earlier step whenever a later one shows a gap. Every step can run again, because each reads its inputs from files.

## Where the files live

Each change is one directory, `plans/<change>/`, that is also an Open Knowledge Format v0.2 bundle. The directory is the source of truth. Related changes of one initiative can share a family directory, `plans/<family>/<change>/`, whose `roadmap.md` orders them and is not a bundle.

[`examples/saved-searches/`](examples/saved-searches/) holds a finished bundle of an invented change, and the check of this plugin validates it.

```text
plans/<change>/
  index.md                     generated
  log.md                       generated from log/
  plan.md
  research/<slug>-<suffix>.md
  decisions/<slug>-<suffix>.md
  critiques/<slug>-<suffix>.md
  diagrams/                    sources and images
  work/<slug>-<suffix>.md
  implementation/<work-id>/notes.md
  verdicts/<work-id>.md
  deliveries/<work-id>/        commit-message.md, pull-request.md, comments/<suffix>.md
  log/<date>-<skill>-<suffix>.md
  close-out.md
```

- `plan.md` holds the objective, scope and success criteria. Only `writing-plan` and `revising-plan` change it, outside the generated block of links to the decisions and the work.
- `research/`, `critiques/` and `log/` only grow: a change is a new file.
- A decision is never edited. A new decision supersedes it, and the old file gets `status: deprecated` and `superseded_by`.
- `work/` holds the work package and the work items. `defining-work-items` and `revising-plan` edit them in place, and Git keeps the history.
- `implementation/`, `verdicts/` and `deliveries/` hold what the steps of each work item produce. `close-out.md` is the record of `validating-work-package`.
- `index.md` in each directory except `diagrams/`, and `log.md`, are generated. Never edit them by hand.

A file name is `<slug>-<suffix>`, with a random four-character suffix. Files are never moved or renamed, so links stay valid. Only the skill that owns a kind of document creates it. The [layout reference](skills/recording-handoff/references/layout.md) lists the type, owner and schema of each document.

Check a bundle or a family with the validator, `uv run skills/recording-handoff/scripts/plan-bundle.py validate <directory>`. It checks Open Knowledge Format conformance, the schemas and the cross-file rules, and exits non-zero on an error. The `index` and `log` commands, each with `--check` or `--write`, check or regenerate the generated files.

## Record the project settings

Add a `Workflow` section to your project instructions. The [settings reference](skills/recording-handoff/references/settings.md) holds a complete example and the default skills of each phase.

| Setting | Default | Meaning |
| --- | --- | --- |
| `plans` | `plans/` | The directory that holds the bundles, directly or one level down in a family directory. |
| `repositories` | none | Explicit when more than one repository is involved. A map from a name to `{ resource, host }`. |
| `delivery` | none | Explicit. `mode` is `branch`, `trunk` or `pull-request`. |
| `skills` | none | A map from a phase name to the installed skills to load in that phase. The phase name is the name of the workflow step. |

A missing `repositories` map blocks any step that writes to a repository once more than one is involved. A missing `delivery` blocks a step that applies something outward, such as pushing a branch or opening a pull request. The step names the missing setting and stops, and it never guesses a value from a remote. Steps that only read and write plan files carry on.

A step loads knowledge skills from four sources, most specific first: the `skills_to_consult` list of the plan or the work item, a custom workflow skill of your project, the project's `skills` map, and the defaults of the settings reference. The sources add up, and the more specific one wins on conflict. A skill named in a `skills_to_consult` list that is not installed blocks the step; any other missing skill is reported and skipped.

## Validate a change

Run the checks of this plugin from the repository root. They validate the example bundle, check its generated files and run the unit tests of the validator:

```bash
npx moon run plugin-workflow:ci-check
```

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-workflow` in the selected harness. The bundle version is `5.3.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-workflow@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-workflow@xonovex-marketplace
```

Optional companions: `xonovex-agentic`, `xonovex-core`, `xonovex-integrations`, `xonovex-native`.

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [critiquing-plan](skills/critiquing-plan/SKILL.md) | A written plan needs an independent adversarial review before anyone approves it: stress-testing its assumptions, tracing how delivery could fail, finding inputs and sequences it misses, or finding the smallest change that would strengthen a weak section. |
| [distilling-change](skills/distilling-change/SKILL.md) | A change is finished and its work should become reusable skills: an umbrella map skill and granular skills that replay the change on a clean base branch. |
| [illustrating-plan](skills/illustrating-plan/SKILL.md) | The diagrams of a change are to be drawn or redrawn: a lead diagram of what is settled, one diagram for each set of options whose choice changes who calls whom, and an optional system model, saved as PlantUML sources with rendered SVG images, or PNG on request, in the bundle's diagrams directory. |
| [recording-handoff](skills/recording-handoff/SKILL.md) | A workflow step reads or writes the directory of a change: the entry step, plan, decision, work item and log files, names, frontmatter, generated indexes, validation, a new change, or the project's `Workflow` settings. |
| [researching-change](skills/researching-change/SKILL.md) | A change needs checked facts before anyone decides or plans: a migration, a new feature in existing software, a code-quality improvement, an investigation of a problem or a new project. |
| [revising-plan](skills/revising-plan/SKILL.md) | What is waiting must be folded into the plan and its work files: FIXME notes the operator left in the plan or other files, findings of a critique, or the corrections a later research file lists. |
| [settling-questions](skills/settling-questions/SKILL.md) | The open questions of a change need settling one at a time with the operator: the questions a research file lists, open, paused or partly settled decision files, or questions found by questioning. |
| [writing-plan](skills/writing-plan/SKILL.md) | The research is saved and the decisions are recorded, and the plan must now be written: a plan.md that states the objective as the settled picture, the scope, the success criteria and, where variants differ, a table of the differences, with decisions and work linked and never copied, plus the first outline of the work files split into what can start now and what waits on which open decision. |
