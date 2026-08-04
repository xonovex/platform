# delegate: Supervise Roadmap Execution by Delegation

Walk a roadmap's normative ordering, handing each item to an implementation agent and verifying the result yourself before accepting it. **You are the supervisor: select, brief, review, record; never implement**, however small the item, because doing it yourself spends the context the next brief needs. Where `continue` implements one plan and stops, this keeps walking the ordering.

Target: a roadmap (`type: roadmap`), a single plan, or an explicit list of subplans.

## Core Workflow

Before item one, mirror the ordering into the environment's task tracker, one entry per item, blocked-by edges matching the order, so progress survives a lost session.

1. **Select**: walk the ordering (the roadmap's `plans:` list, else its phase checkboxes) to the first item not done. A gated item waits until every prerequisite it names is `status: complete`, confirmed by re-reading that plan's frontmatter, not from memory or the roadmap's own prose. A subplan owned by another family is **borrowed**: schedule it, leave its `parent_plan` where it points, and do not re-own it.
2. **Brief**: assemble the brief (below) from your own reading of the subplan, its parent plan, and the roadmap's governing decisions. This is what the supervisor's context is for.
3. **Execute**: spawn one implementation agent per subplan, on the requested implementation model, and wait for it; under parallel-by-group ordering, spawn a whole `parallel_group` at once and wait for all of them.
4. **Review**: always, and always yourself (below).
5. **Record**: tick the roadmap's line, close the tracker entry, and commit with a conventional-commit message scoped to the subplan, batching commits to the end only on request; for a borrowed item whose owning family still has work left in it, record "done for this family" rather than "complete". Never push unless the user asked.
6. **Advance**: move to the next item. On a blocker, or a criterion genuinely unmet, stop and report; never paper over one to keep the loop running.
7. **Close out**: with the ordering exhausted, re-run every touched package's full suite on the final tree, then record family completion as the roadmap's completed siblings do (parent plan frontmatter, roadmap entry, criteria ticked), naming anything deliberately deferred.

## The brief

The implementation agent starts with no context, so the brief carries all of it:

- [ ] **Required reading**: the subplan, its parent plan, the completed predecessor subplans, the `AGENTS.md` of every package the item touches, and the existing source files the work builds on, by path
- [ ] **Carried facts**: the downstream-impact notes from every predecessor's report (renamed types, new APIs, changed capacities, moved conventions), restated so the agent inherits them instead of rediscovering them
- [ ] **Skills**: every entry in the subplan's `skills_to_consult`, loaded before implementation starts
- [ ] **Hard constraints**: the governing decisions bearing on this subplan, restated concretely (value currency and units, determinism pins, which package may see which type, real-time-safety rules) rather than cited by number
- [ ] **Deliverables and tests**: the subplan's tasks and success criteria, verbatim
- [ ] **Validation commands**: every task that must be green, sanitizer presets, format checks, and dependent packages' suites included
- [ ] **Working agreement**: work on the current branch; do not commit, do not push
- [ ] **Document update**: on completion set `status` and `validation` frontmatter, tick only the criteria actually met, add a Results section in the factual style of its sibling subplans, and leave `status: in_progress` with reasons if one is unmet
- [ ] **Required return**: status; files touched; every criterion met or not-met with the test that evidences it; validation commands with their results; deviations with rationale; downstream impact on consuming plans

## Review

The agent's report is input to review, never a substitute for it: re-running is the one step that cannot be delegated.

- Re-run the key suites yourself, sanitizer presets included: a green claim is a claim.
- Where a criterion is visual, open the captures the agent named: an assertion can pass while its screenshot shows an empty panel.
- Read the diff against the success criteria and the repo's code and writing conventions.
- Check the document is honest: unmet criteria left unticked, deviations enumerated with rationale and their impact on consuming plans, `validation` frontmatter matching what you just ran.
- Distrust editor and stale-index diagnostics, and settle every dispute with a real build.
- On failure, send the findings to the **same** agent for a fix pass; a fresh one repeats discovery and often reintroduces the finding.

## Defaults

- Sequential ordering. Parallel-by-group only on request, and only for members whose file sets are genuinely disjoint: check the sets, not the group label, or the conflict lands mid-review.
- Plain subagents; scripted workflow orchestration only on request.
- Run until the ordering is exhausted or something blocks, unless given a stop condition (a count of items, or a named plan to finish on).
- The implementation model is a parameter with a strong default, never hardcoded; the supervisor stays on the stronger model.

## Output (per item)

```text
Roadmap: plans/example-family/roadmap.md, item 3 of 7 (sequential)
Item: plans/other-family/subsystem/02-scheduler.md (borrowed, owned by plans/other-family/)
Brief: 5 required reads · 3 skills · 6 criteria · 4 validation commands
Agent: complete, 11 files · 5/6 criteria met · 2 deviations
Review: re-ran test + asan + format-check → green · diff matches criteria · doc honest
Record: roadmap line ticked (done for this family) · commit feat(example): add scheduler
Open: criterion 4 measured -46 dB against the -60 dB asked for, carried to plans/example-family/verification.md
```

## Gotchas

- A criterion measured close but short is a miss: record the number and let the consuming plan decide, because a roadmap that rounds up stops working as a schedule
- A borrowed subplan marked `complete` can still carry deferred integration checks; following an ordering never reaches them, so a `pending` verification plan has to own them explicitly
- An agent may fix a defect the plan assumed away (a missing pin, a dropped parameter): accept it when the fix is at root cause and tested, and carry it forward, because a hidden fix re-surfaces as the next item's mystery
- User feedback that arrives mid-roadmap is a new item, not a detour: brief it like any other, with the report of the agent whose work it touches
