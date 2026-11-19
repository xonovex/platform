# Work item template

Copy the block to `plans/<change>/work/<slug>-<suffix>.md`, replace every angle-bracket placeholder and delete any comment lines. For a work package, set `kind: work-package` and drop `criteria` and `parent`.

```markdown
---
type: Workflow Work Item
schema: workflow/work-item/1
title: <Work item title>
description: <One sentence stating the outcome of the work>
generated: {by: <skill>/<version>, at: <timestamp>}
status: draft
kind: work-item
parent: work/<slug>-<suffix>
team: <team name>
repository: <repository-name>
readiness: ready
depends_on:
  - work/<slug>-<suffix>
decisions:
  - decisions/<slug>-<suffix>
criteria:
  - id: c1
    text: <Observable condition that must hold when the work is done>
    verification: <How the condition is checked>
skills_to_consult:
  - <installed-skill-name>
---

# Objective

<The outcome this work delivers and who benefits, in two or three sentences.>

<!-- Include "Notes" when the implementer needs context that the criteria do not carry. -->

# Notes

- <Constraint, hint or pointer>
```

## Fill in

- `kind: work-package` groups work items through their `parent` key. A work item whose package is unknown omits `parent`.
- `readiness: waiting` requires a non-empty `waits_on` list of decision concept IDs. Add it directly after `readiness`. Use `ready` when no open decision blocks the work.
- `repository` is a name from the project's repository map; omit it when the work touches no repository.
- `depends_on` lists work that must finish first. A cycle is a validation error.
- `decisions` lists the settled decisions that apply to this work. They are referenced, never copied.
- `criteria` ids are `c1`, `c2`, and so on, unique in the file. `verification` says how a verifier checks the criterion (a command, a test, a screenshot, a review question).
- `status: draft` stays until the invoker confirms the work item. Then change it to `stable` (or delete the key) and add a `verified` entry that names the invoker as `human:<id>` (taken from the invocation; ask for it when missing, never invent one): `verified: {by: human:<id>, at: <timestamp>}`, placed directly after `generated`.
- Every `<timestamp>` comes from the system clock: run `date -u +%Y-%m-%dT%H:%M:%SZ` and use its output; never invent, round or copy a time from another file. `<version>` in `generated.by` is the `version` field of the `plugin.json` of the plugin that ships the skill.
- Write a `title`, `description` or other string value in single quotes when it contains a colon followed by a space, a quote or a leading special character. The same holds for `criteria` text and `verification`.
- The file name is `<slug>-<suffix>` with a suffix of four random characters from `a-z` and `2-7` only (never `0`, `1`, `8` or `9`).
- A work package file is optional: a work item with no `parent` belongs to no package, and its package is the set of the plan's work items that are not `deprecated`.
