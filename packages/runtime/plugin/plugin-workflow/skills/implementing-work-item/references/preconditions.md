# Preconditions

Check these after the entry step and before reading any code. One unmet precondition is reported and the run stops there. Never write a stand-in, never start the step that would satisfy it, and never carry on "just to prepare".

## Contents

- The checks
- Reporting an unmet precondition
- Dependencies that count as done

## The checks

1. **The work file exists** at `work/<slug>-<suffix>.md`, is `kind: work-item` and validates. A work package is not implemented; it groups items.
2. **The criteria say how they are verified.** Every entry of `criteria` has a non-empty `verification` that names a command, a test, a screenshot or a review question. A criterion without one cannot be checked at the end, so the work cannot be called done.
3. **`readiness` is `ready`.** A `waiting` item waits on open decisions listed in `waits_on`. Name them and stop; **settling-questions** settles them.
4. **The blockers are done.** Every concept ID in `depends_on` counts as done by one of two proofs, see below.
5. **The repository is in the project's map.** The work file's `repository` names an entry of the `repositories` map in the project's `Workflow` settings, and that entry's resource is cloned. A work item with no `repository` has nothing to implement in: report it.
6. **The settings it needs are present.** The `repositories` map is needed. The `delivery` setting is not, because nothing in this step goes outward.
7. **The skills to consult are installed.** A skill named in a `skills_to_consult` list that is not installed is a blocker. Skills named only by the project's `skills` map or by a default of the phase are reported and skipped. See the settings reference of **recording-handoff**.

## Reporting an unmet precondition

Report in a few lines: which precondition, what you found, and the step or the person that normally satisfies it, in words. Leave every file as it was: no notes file, no worktree, no log entry. Offer the next step and wait.

> Stopped before any edit. The work item `work/capped-total-api-7k2f` depends on `work/fare-rules-model-3m4q`, and there is no verdict for it. Run **validating-work-item** on that item, or tell me it is done, and I continue.

## Dependencies that count as done

A blocker in `depends_on` is done when either holds:

- `verdicts/<work-id>.md` exists for it and every entry of its `results` is `pass`. A `blocked` or `fail` result, or a verdict file with a `source` of `self-check`, does not count.
- The invoker says so in the request. Record the words in the log entry's `authorization`, in single quotes when they contain a colon followed by a space.

A blocker whose work is merged but has no verdict needs the invoker's word. Do not infer done from a status on a host or a green build.
