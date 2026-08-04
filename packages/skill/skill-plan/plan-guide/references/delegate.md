# delegate: Supervise Roadmap Execution by Delegation

Walk a roadmap's normative ordering, handing each item to an implementation agent and verifying the result before accepting it. **You are the supervisor: select, brief, review, record; never implement**, however small the item: doing it yourself spends the context the next brief needs.

Target: a roadmap (`type: roadmap`), a single plan, or an explicit list of subplans.

## Core Workflow

First mirror the ordering into the environment's task tracker, one entry per item with blocked-by edges, so progress survives a lost session.

1. **Select** the first item not done (`plans:`, else phase checkboxes). A gated item waits until every prerequisite reads `status: complete` in that plan's own frontmatter, not in memory or the roadmap's prose. An item another family owns is **borrowed**: schedule it, leave its `parent_plan`, do not re-own it.
2. **Brief** from your own reading of the subplan, its parent, and the governing decisions (below).
3. **Execute**: one agent per subplan on the requested model, then wait; under parallel-by-group, spawn the whole `parallel_group` at once and wait for all.
4. **Review**: always, and always yourself (below).
5. **Record**: tick the roadmap line, close the tracker entry, commit with a conventional-commit message scoped to the subplan. A borrowed item reads "done for this family", not "complete". Never push unless asked.
6. **Advance**. On a blocker or an unmet criterion, stop and report; never paper over one to keep the loop running.
7. **Close out**: re-run every touched package's full suite on the final tree, then record family completion as the completed siblings do (parent frontmatter, roadmap entry, criteria ticked), naming what was deferred.

## The brief

The agent starts with no context, so the brief carries all of it:

- [ ] **Required reading**, by path: the subplan, its parent, completed predecessors, the `AGENTS.md` of each package touched, the source files built on
- [ ] **Carried facts**: predecessors' downstream-impact notes (renamed types, new APIs, changed capacities, moved conventions), so the agent inherits rather than rediscovers them
- [ ] **Skills**: every entry in `skills_to_consult`, loaded before implementation
- [ ] **Hard constraints**: governing decisions restated concretely (units a value carries, versions pinned, which module sees which type, what a hot path may not do), never cited by number
- [ ] **Deliverables and tests**: the subplan's tasks and success criteria, verbatim
- [ ] **Validation commands**: every task that must be green, strictest presets, format checks, dependent packages' suites
- [ ] **Working agreement**: the current branch; do not commit, do not push
- [ ] **Document update**: set `status` and `validation` frontmatter, tick only criteria met, add a Results section in the siblings' factual style, leave `in_progress` with reasons if one is unmet
- [ ] **Required return**: status; files touched; each criterion met or not-met with its evidencing test; validation results; deviations with rationale; downstream impact

## Review

The report is input to review, never a substitute: re-running is the one step that cannot be delegated.

- Re-run the key suites yourself, strictest presets included.
- Open the captures behind a visual criterion: an assertion passes while its screenshot shows an empty panel.
- Read the diff against the criteria and the repo's code and writing conventions.
- Check the document is honest: unmet criteria unticked, deviations and their downstream impact written down, `validation` matching what you ran.
- Distrust editor and stale-index diagnostics; settle disputes with a real build.
- Send findings to the **same** agent: a fresh one repeats discovery and reintroduces the finding.

## Defaults

- Sequential. Parallel-by-group on request, and only for genuinely disjoint file sets: check the sets, not the label.
- Plain implementation agents; scripted orchestration on request.
- Run until the ordering is exhausted or something blocks, absent a stop condition (a count, or a plan to finish on).
- The implementation model is a parameter with a strong default, never hardcoded; the supervisor stays on the stronger model.

## Output (per item)

```text
Roadmap: plans/example-family/roadmap.md, item 3 of 7 (sequential)
Item: plans/other-family/subsystem/02-scheduler.md (borrowed, owned by plans/other-family/)
Brief: 5 required reads · 3 skills · 6 criteria · 4 validation commands
Agent: complete, 11 files · 5/6 criteria met · 2 deviations
Review: re-ran test + lint + format-check → green · diff matches criteria · doc honest
Record: roadmap line ticked (done for this family) · commit feat(example): add scheduler
Open: criterion 4 measured 140 ms against the 100 ms asked for, carried to plans/example-family/verification.md
```

## Gotchas

- Measured close but short is a miss: record the number and let the consuming plan decide, or the roadmap stops working as a schedule
- A borrowed subplan marked `complete` can still carry deferred integration checks that following an ordering never reaches, so a `pending` verification plan owns them
- An agent may fix a defect the plan assumed away: accept it when the fix is at root cause and tested, and carry it forward, or it re-surfaces as the next item's mystery
- Mid-roadmap user feedback is a new item, not a detour: brief it like any other, with the report of the agent whose work it touches
