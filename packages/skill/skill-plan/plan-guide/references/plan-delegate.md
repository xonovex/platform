# plan-delegate: Supervise Roadmap Execution by Delegation

Work a roadmap's normative ordering by handing each item to an implementation agent and independently verifying the result before accepting it. **You are the supervisor: you select, brief, review, and record — you never implement.** Where `plan-continue` implements one plan and stops, this operation keeps walking the ordering until it is exhausted, a stop condition is reached, or something blocks.

Target: a roadmap (`type: roadmap`), a single plan, or an explicit list of subplans.

## Core Workflow

Before item one, mirror the ordering into the environment's task tracker — one entry per item, blocked-by edges matching the order — so progress survives a lost session.

1. **Select** — walk the ordering (the roadmap's `plans:` list, else its phase checkboxes) to the first item not yet done. Honour gates: a gated item does not start until every prerequisite it names is `status: complete`, confirmed by re-reading each dependency plan's frontmatter, never from memory or from the roadmap's prose about itself. A subplan owned by another family is **borrowed**: schedule it, leave its `parent_plan` pointing where it already does, and do not re-own it.
2. **Brief** — read the subplan, its parent plan, and the roadmap's governing decisions, then assemble the brief (below). This is what the supervisor's context is for.
3. **Execute** — spawn one implementation agent per subplan, on the requested implementation model, and wait for it. Under parallel-by-group ordering, spawn a whole `parallel_group` in one go and wait for all of them.
4. **Review** — always, and always yourself (below).
5. **Record** — tick the roadmap's line for the item and close its tracker entry; for a borrowed item whose owning family still has work left in it, record "done for this family" rather than "complete". Then commit with a conventional-commit message scoped to the subplan. Never push unless the user asked.
6. **Advance** — move to the next item. On a blocker, or a criterion that is genuinely unmet, stop and report; never paper over one to keep the loop running.

## The brief

The implementation agent starts with no context, so the brief carries all of it:

- [ ] **Required reading** — the subplan, its parent plan, the completed predecessor subplans, and the concrete existing source files the work builds on, by path
- [ ] **Skills** — every entry in the subplan's `skills_to_consult`, loaded before implementation starts
- [ ] **Hard constraints** — the governing decisions and non-deferrable constraints that bear on this subplan, restated concretely (value currency and units, determinism pins, which package may see which type, real-time-safety rules) rather than cited by number
- [ ] **Deliverables and tests** — the subplan's tasks and success criteria, verbatim
- [ ] **Validation commands** — the exact tasks that must be green, including sanitizer presets, format checks, and the suites of dependent packages
- [ ] **Working agreement** — work on the current branch; do not commit, do not push
- [ ] **Document update** — on completion, set the subplan's `status` and `validation` frontmatter, tick only the criteria actually met, and add a Results section in the factual style of its sibling subplans; leave `status: in_progress` with reasons if any criterion is unmet
- [ ] **Required return** — status; files touched; every success criterion marked met or not-met with the test name that evidences it; validation commands with their results; deviations with rationale; downstream impact on consuming plans

## Review

The agent's report is input to review, never a substitute for it.

- Re-run the key suites yourself, sanitizer presets included — a green claim is a claim.
- Read the diff against the success criteria and against the repo's code and writing conventions.
- Check the document is honest: unmet criteria left unticked, deviations written down, `validation` frontmatter matching what you just ran.
- Distrust editor and stale-index diagnostics, and settle every dispute with a real build.
- On failure, send the findings to the **same** agent for a fix pass; it still holds the context a fresh agent would have to rebuild.

## Defaults

- Sequential ordering. Parallel-by-group only on request, and only for group members whose file sets are genuinely disjoint.
- Plain subagents. A scripted workflow orchestration only when the user asks for one.
- Run until the ordering is exhausted or something blocks, unless given a stop condition (a count of items, or a named plan to finish on).
- One commit per accepted subplan, after review. Batching to the end only on request.
- The implementation model is a parameter with a strong default; the supervisor stays on the stronger model. Never hardcode either.

## Output (per item)

```text
Roadmap: plans/example-family/roadmap.md — item 3 of 7 (sequential)
Item: plans/other-family/subsystem/02-scheduler.md (borrowed, owned by plans/other-family/)
Brief: 5 required reads · 3 skills · 6 criteria · 4 validation commands
Agent: complete — 11 files · 5/6 criteria met · 2 deviations
Review: re-ran test + asan + format-check → green · diff matches criteria · doc honest
Record: roadmap line ticked (done for this family) · commit feat(example): add scheduler
Open: criterion 4 measured -46 dB against the -60 dB asked for — carried to plans/example-family/verification.md
```

## Gotchas

- Implementing "just this small bit" yourself spends the context the next brief needs — delegate every item, however small
- A criterion measured close but short is a miss, not a tick — record the measured number and let the consuming plan decide, because a roadmap that rounds up stops being usable as a schedule
- Reviewing from the agent's report alone certifies its self-assessment, not the build — re-running is the only step that cannot be delegated
- A borrowed subplan marked `complete` can still carry deferred integration checks; following an ordering never reaches them, so a `pending` verification plan has to own them explicitly
- Members of one `parallel_group` that touch the same file leave the supervisor untangling a conflict mid-review — check the file sets, not the group label
- Spawning a fresh agent for a review-fix pass repeats discovery and often reintroduces the finding
- Deviations from the plan as written are expected and worth keeping: enumerated with rationale, with their impact on consuming plans recorded
