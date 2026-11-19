---
name: delegating-plan
description: "Use when supervising the execution of a roadmap, a plan, or a list of subplans by handing each item to an implementation agent and verifying its result. Triggers on working a roadmap as supervisor, a brief for an agent per subplan, independent review of delegated work, sequential or parallel groups of items, a stop condition, and one commit per accepted item, even when the user doesn't say 'delegate' and asks to run the whole roadmap with agents."
---

# Supervise Plan Execution by Delegation

Walk a roadmap's normative ordering, handing each item to an implementation agent and verifying the result before accepting it. **You are the supervisor: select, brief, review, record; do not implement**, except an item of a few edits in files you already read: a new agent first pays its whole startup context and its own discovery, which costs more than those edits.

Target: a roadmap (`type: roadmap`), a single plan, or an explicit list of subplans.

## Inputs

- **Roadmap** (required): the roadmap, plan, or explicit list of subplans to work through
- **Model** (optional): the model for the implementation agents; defaults to the mid tier, and the supervisor keeps its own
- **Order** (optional): sequential or parallel; defaults to sequential, and parallel runs a `parallel_group` concurrently when its file sets are disjoint
- **Orchestration** (optional): agents or workflow; defaults to agents, and workflow uses a scripted multi-agent orchestration
- **Stop after** (optional): a count of items or a named plan; the default runs until the ordering is exhausted or something blocks
- **Commit** (optional): one conventional commit per accepted item, or one batch at the end; defaults to one per item

## Core Workflow

First mirror the ordering into the environment's task tracker, one entry per item with blocked-by edges, so progress survives a lost session.

1. **Select** the first item not done (`plans:`, else phase checkboxes). A gated item waits until every prerequisite reads `status: complete` in that plan's own frontmatter, not in memory or the roadmap's prose. An item another family owns is **borrowed**: schedule it, leave its `parent_plan`, do not re-own it.
2. **Brief** from your own reading of the subplan, its parent, and the governing decisions, see [references/brief.md](references/brief.md).
3. **Execute**: one agent per subplan on the requested model, then wait; under parallel-by-group, spawn the whole `parallel_group` at once and wait for all.
4. **Review**: always, and always yourself, see [references/review.md](references/review.md).
5. **Record**: tick the roadmap line, close the tracker entry, commit with a conventional-commit message scoped to the subplan. A borrowed item reads "done for this family", not "complete". Never push unless asked.
6. **Advance**. On a blocker or an unmet criterion, stop and report; never paper over one to keep the loop running.
7. **Close out**: re-run every touched package's full suite on the final tree, then record family completion as the completed siblings do (parent frontmatter, roadmap entry, criteria ticked), naming what was deferred.

## Defaults

- Sequential. Parallel-by-group on request, and only for genuinely disjoint file sets: check the sets, not the label.
- Plain implementation agents; scripted orchestration on request.
- Run until the ordering is exhausted or something blocks, absent a stop condition (a count, or a plan to finish on).
- The implementation model is a parameter, never hardcoded, defaulting to the mid tier of the harness: a briefed item needs execution more than judgement, and the agents make most of the calls. The supervisor stays on the stronger model.

## Boundaries

One target without a supervisor belongs to **implementing-plan**. The items come from **defining-subplans**, and the closeout record after the run belongs to **closing-plan**.

## Gotchas

- Measured close but short is a miss: record the number and let the consuming plan decide, or the roadmap stops working as a schedule
- A borrowed subplan marked `complete` can still carry deferred integration checks that following an ordering never reaches, so a `pending` verification plan owns them
- An agent may fix a defect the plan assumed away: accept it when the fix is at root cause and tested, and carry it forward, or it re-surfaces as the next item's mystery
- Mid-roadmap user feedback is a new item, not a detour: brief it like any other, with the report of the agent whose work it touches

## Progressive Disclosure

- Read [references/brief.md](references/brief.md) - Load when writing the brief for an implementation agent
- Read [references/review.md](references/review.md) - Load when reviewing the result of an implementation agent
- Read [references/output.md](references/output.md) - Load when reporting an item to the user
