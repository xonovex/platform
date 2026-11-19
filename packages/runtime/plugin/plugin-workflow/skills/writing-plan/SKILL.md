---
name: writing-plan
description: "Use when the research is saved and the decisions are recorded, and the plan must now be written: a plan.md that states the objective as the settled picture, the scope, the success criteria and, where variants differ, a table of the differences, with decisions and work linked and never copied, plus the first outline of the work files split into what can start now and what waits on which open decision. Triggers on 'write the plan', 'turn the research and decisions into a plan', 'what can start now and what waits', even when the user doesn't say 'plan'. Not for research, settling a decision, changing a plan that exists, or completing the work items."
---

# Write the Plan

Purpose: turn the research and the recorded decisions into one plan that tells a reader what the change delivers, what is in and out, how success is observed, and into a first outline of the work that tells each team what it can start now and what waits on which decision. The plan states settled things flat and links everything else.

**Load first:** the **recording-handoff** skill, and run its entry step before anything below.

## Essentials

- The plan holds only what no other file holds: objective, scope, optional current state, optional differences between variants, success criteria and a generated block of links. It never copies decision text or work item definitions.
- The objective is a picture of the settled state only. A rejected option, the reasoning against it and a mechanism nobody agreed to do not appear. An open decision is linked, never summarized.
- Every work row has an owner team and a readiness: `ready`, or `waiting` with the open decisions it waits on named.
- The generated block is rewritten from the files that exist, never typed by hand.
- The plan and the work files use the shared plan and work item templates of **recording-handoff**. This skill has no template or schema of its own.
- Sensitive findings stay out of every file and are named in the reply.
- Never commit or stage. Never start another step.

## Procedure

1. **Run the entry step** from **recording-handoff**. The plan and the work files go into the plan directory's repository, so confirm that branch. The step's own preconditions: the bundle holds at least one research file or decision, and no `plan.md` yet. When a plan exists, report it and stop: changing it belongs to **revising-plan**.
2. **Load the skills that apply** from the skills map that the settings reference of **recording-handoff** defines, for the phase `writing-plan`: the `skills_to_consult` of the research files, a custom workflow skill the project provides, the `writing-plan` entry of the project's `skills` map, and the default skills of the phase when the project's map does not name it.
3. **Read the inputs yourself.** Read every research file (a file named in another's `corrects` is read through the file that corrects it) and every decision that is not deprecated, grouped by `state` from `decisions/index.md`. Note each owner, the repositories and the skills the research found. Collect the findings that must stay out of the files. A research question with no decision file is unrecorded: list it for the reply, and do not outline work that depends on it. See [references/plan-inputs.md](references/plan-inputs.md).
4. **Draft the plan** with the front matter from [references/plan-inputs.md](references/plan-inputs.md), then section by section: objective, scope, current state when the reader needs it, differences between variants when they differ, success criteria. See [references/plan-sections.md](references/plan-sections.md) and [references/writing-rules.md](references/writing-rules.md).
5. **Outline the work.** Choose the file name of every work row first, then draft each file from the shared work item template, so `depends_on` can name rows that do not exist yet. See [references/work-outline.md](references/work-outline.md).
6. **Write the files.** Write each work file, then `plan.md` from the shared plan template with the generated block rewritten from the work and decision files that exist. Announce every write in one line: the path and what it holds.
7. **Check.** Run the project's prose check and fix what it reports. Write one `Workflow Log Entry` from the **recording-handoff** template with `skill: writing-plan`, `action: Creation` and every created concept ID in `touched`. Run `index --write` and `log --write`, then `validate`, and fix every finding in your files.
8. **Report** the outcome first: the plan written, the work that can start now, the work that waits and on which decision, the questions with no decision file, and the sensitive findings kept out, by topic. Offer the next steps: **settling-questions** for the open decisions, **critiquing-plan** for a critique, and **defining-work-items** to complete the work items. Stop.

## When not to use

- Facts are missing: use **researching-change** first.
- A question is open or an answer needs recording: use **settling-questions**.
- A plan exists and must change from `FIXME` notes, critique findings or research corrections: use **revising-plan**.
- The work files need full criteria and work packages: use **defining-work-items**. This skill writes only the outline.

## Gotchas

- Summarizing an open decision in the plan. Link it; the decision file owns the question and the options.
- A rejected option, or the reason it lost, in the objective. State the settled picture and nothing else.
- A mechanism the invoker never agreed to, added because it seems the natural design. If no decision states it, it is absent or an open decision.
- Copying a decision sentence or a work objective into the plan. The reader follows the link.
- A criterion that names an implementation detail, such as a class or a library. Name what is observed and by which means.
- A waiting work file with no decision named in `waits_on`, or a ready file that depends on an open decision.
- A work row with no team, or a team invented to fill the field. Ask once.
- A criterion in a work file that restates the objective. State a condition that a check can fail.
- A hand-written generated block. Rewrite it whole from the files.
- Running the prose check after reporting, or reporting what it found as done.
- Writing a `verified` entry for a plan nobody confirmed. The plan stays `status: draft`.

## Done when

- `plan.md` has the type, schema, `status: draft`, `owners`, `repositories`, `skills_to_consult` and `sources` that the inputs support, and the headings of the plan template in order.
- No sentence of a decision file or a work objective is repeated in the plan, and the settled text carries no rejected reasoning.
- Every `waiting` work file lists the open decisions in `waits_on`, every file names a team and has at least one criterion with its verification.
- The generated block links exactly the decisions that are not deprecated and the work files that exist.
- The bundle validates with no error and no schema-not-found warning, the prose check passes, and the log entry exists.

## Progressive Disclosure

- Read [references/plan-inputs.md](references/plan-inputs.md) - Load when reading the research and the decisions, or choosing the front matter of the plan.
- Read [references/plan-sections.md](references/plan-sections.md) - Load when drafting a section of the plan or rewriting the generated block.
- Read [references/work-outline.md](references/work-outline.md) - Load when deciding the work rows, their teams, readiness and dependencies, or writing a work file.
- Read [references/writing-rules.md](references/writing-rules.md) - Load when stating settled things, handling open decisions, unrecorded questions or sensitive findings, or running the prose check.
