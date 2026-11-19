# Plan template

Copy the block to `plans/<change>/plan.md`, replace every angle-bracket placeholder and delete any comment lines.

```markdown
---
type: Workflow Plan
schema: workflow/plan/1
title: <Plan title>
description: <One sentence stating what the change delivers>
generated: {by: <skill>/<version>, at: <timestamp>}
status: draft
owners:
  - <owner name or team>
repositories:
  - name: <repository-name>
    resource: <repository-url>
skills_to_consult:
  - <installed-skill-name>
tags:
  - <tag>
---

# Objective

<What the change achieves and why it matters, in two or three sentences.>

# Scope

In scope:

- <Capability or area the change covers>

Out of scope:

- <Capability or area the change leaves alone>

# Success criteria

- <Observable result that shows the change is done>

<!-- generated:begin -->

## Decisions

- [<Decision title>](/decisions/<slug>-<suffix>.md)

## Work

- [<Work title>](/work/<slug>-<suffix>.md)

<!-- generated:end -->
```

## Fill in

- `generated.by` is the skill and its version. `generated.at` is the moment of the last meaningful change. Every `<timestamp>` comes from the system clock: run `date -u +%Y-%m-%dT%H:%M:%SZ` and use its output; never invent, round or copy a time from another file. `<version>` in `generated.by` is the `version` field of the `plugin.json` of the plugin that ships the skill.
- `status: draft` stays until the invoker confirms the plan. Then change it to `stable` (or delete the key) and add a `verified` entry that names the invoker as `human:<id>`, taken from the invocation; ask for it when it is missing and never invent one.
- `owners`: include when named people or teams answer for the change.
- `repositories`: include when the change writes to a repository. `name` is the key in the project's repository map.
- `skills_to_consult`: include when this plan needs skills beyond the project's phase map.
- Scope: keep both lists even when one holds a single line, so a reader sees what is excluded.
- The block between the `generated` markers lists decisions and work by link, one line each. Rewrite it whenever a decision or work file is added or its title changes. Never copy decision text or work item definitions into the plan.
- Write a `title`, `description` or other string value in single quotes when it contains a colon followed by a space, a quote or a leading special character.
