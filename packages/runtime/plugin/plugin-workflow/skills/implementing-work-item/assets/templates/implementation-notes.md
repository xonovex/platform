# Implementation notes template

Copy the block to `plans/<change>/implementation/<work-id>/notes.md`, where `<work-id>` is the work file's name without `.md`. Replace every angle-bracket placeholder and delete any comment lines. A rerun on the same work item edits this file in place; git keeps the history.

```markdown
---
type: Workflow Implementation Notes
schema: workflow/implementation-notes/1
title: <Notes title as a short noun phrase>
description: <One sentence stating what was implemented and how far it got>
generated: {by: implementing-work-item/<version>, at: <output of date -u +%Y-%m-%dT%H:%M:%SZ>}
work: work/<slug>-<suffix>
worktree: <path of the worktree>
branch: <branch name>
tasks:
  - id: <task-id>
    concern: <The concern the task owns>
    summary: <What the task changed, in one sentence>
decision_notes:
  - decision: decisions/<slug>-<suffix>
    location: <path-in-the-work-repository>:<line>
    note: <How the decision shaped the code at this place>
checks:
  - command: <Command as run>
    result: <Outcome in one line, such as "passed, 219 tests" or "failed, 2 tests">
---

# Summary

<What the work item now does, what state the worktree is in, and whether anything is left. Two or three sentences.>

# Self-check against the criteria

<!-- Every row is the implementer's own evidence. It is a self-check and is never shown as validated. -->

| Criterion | Evidence                                 | State                   |
| --------- | ---------------------------------------- | ----------------------- |
| `c1`      | <Command or observation that was re-run> | <met, not met, blocked> |

<!-- Include "Open points" when anything is unmet, deviates from the work item or needs a person. -->

# Open points

- <Unmet criterion, deviation with its reason, or a question for the invoker>
```

## Fill in

- `generated.by`: `implementing-work-item/<version>`, where `<version>` is the `version` field of the `plugin.json` of the plugin that ships the skill (read it, do not guess). `generated.at` is the output of `date -u +%Y-%m-%dT%H:%M:%SZ`, never a time copied from another file. The skill wrote the file, so leave out `verified`.
- Write the file into the main checkout's bundle, never into the worktree's tracked copy of `plans/`.
- `work`: the concept ID of the work file, `work/<slug>-<suffix>`.
- `worktree` and `branch`: where the changes are and the branch they sit on. The branch carries the work file's name, as the worktree procedure of **using-git** forms it.
- `tasks`: one entry for each task, in the order they ran. `id` is kebab-case and unique in the file. `concern` names what the task owns, such as `request-validation` or `app-total`.
- `decision_notes`: one entry only where a decision shaped code that the code itself cannot explain. `decision` is the decision's concept ID. `location` is `path:line` relative to the work repository's root, checked against the final state of the files. Write `decision_notes: []` when no decision shaped the code.
- `checks`: every command the criteria name and every other check you ran, with the outcome as observed. Write the commands you ran yourself, not the ones a worker reported.
- `# Self-check against the criteria`: one row for each criterion of the work file. The state `met` needs evidence you re-ran. Never write `validated` or `passed`; a fresh validation is another step.
