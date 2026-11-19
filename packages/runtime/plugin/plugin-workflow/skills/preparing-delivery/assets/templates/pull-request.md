# Pull request description template

Copy the block to `plans/<change>/deliveries/<work-id>/pull-request.md`, replace every angle-bracket placeholder and delete any comment lines. Write this file only when the delivery mode is `pull-request`. The body is the description as the host will show it and stays under about 3,500 characters.

```markdown
---
type: Workflow Pull Request Description
schema: workflow/pull-request/1
title: <Pull request title>
description: <One sentence stating what the pull request delivers>
generated: {by: preparing-delivery/<version>, at: <timestamp>}
work: work/<slug>-<suffix>
base: <target-branch>
head: <work-item-branch>
verdict: verdicts/<work-id>
verdict_matches: true
---

## Work item

<The work item's concept ID.>

## Acceptance criteria

<!-- Use the marked form only when verdict_matches is true. -->

- c1 (met): <criterion text>
- c2 (missing): <criterion text>

<!-- Use the unmarked form when verdict_matches is false. -->

- c1: <criterion text>
- c2: <criterion text>

No verdict matches this tree, so no criterion is marked.

## Plan

<Plan link: a URL only for a browsable hosted repository, else `<repository name>: plans/<change>/plan.md`.>

## Decisions

- `decisions/<slug>-<suffix>`: <the claim in one line>

## How it was tested

- <Check or verdict, and what it showed. Label the implementer's own checks as self-check.>
```

## Fill in

- `title`: short, in the conventional commit form of **using-git** or the project's convention. The body follows **authoring-pull-requests**. `base` is the branch the pull request targets. `head` is the work item's branch.
- `verdict`: the concept ID of the verdict file, or `none` when no verdict exists. `verdict_matches` is `true` only when the verdict's `worktree_hash` equals the current tree hash of the worktree; with `verdict: none` it is `false`.
- Acceptance criteria: take the text from the work file and never reword it. Marked form: `met` for a criterion whose result is `pass`, `missing` for `fail`, `blocked` or no result. Unmarked form: no mark at all, and the closing line stays. When the verdict is a `self-check`, write `(met, self-check)` and never the word validated.
- Decisions: one line each for the decisions the work item cites, with the claim in one line from the decision's `# Decision` sentence.
- How it was tested: with a matching independent verdict, name the verdict and the first 12 characters of its tree hash. Add the checks from the implementation notes, each labelled self-check.
- Keep one of the two criteria forms and delete the other with its comments. Delete the `Decisions` section when the work item cites none.
- Plan link: a URL only when the repository's `resource` is a browsable URL for files (an `https://` URL of a hosted repository); otherwise the repository name and the path `plans/<change>/plan.md`, for example `project: plans/greeting-locale/plan.md`. Never a file URL or a bare repository URL. Work item: the work item's concept ID.
- `generated.by` is `preparing-delivery/<version>`, with `<version>` read from the `version` field of the `plugin.json` of the plugin that ships this skill. `generated.at` is the output of `date -u +%Y-%m-%dT%H:%M:%SZ`; never invent or copy a time. The skill wrote the file, so leave out `verified`.
- A frontmatter value with a colon followed by a space, a quote or a leading special character is written in single quotes; a `title` like `feat: add retry limit` needs them.
