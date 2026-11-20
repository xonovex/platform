# Commit message template

Copy the block to `plans/<change>/deliveries/<work-id>/commit-message.md`, replace every angle-bracket placeholder and delete any comment lines. The body of the file is the commit message itself: the first line is the subject, a blank line follows, then the explanation. Nothing outside the message goes in the body.

```markdown
---
type: Workflow Commit Message
schema: workflow/commit-message/1
title: <Commit message for the work item>
description: <One sentence stating what the commit delivers>
generated: {by: preparing-delivery/<version>, at: <timestamp>}
work: work/<slug>-<suffix>
branch: <branch-name>
---

<Subject line in the project's commit convention, at most 72 characters>

<Why the change exists and what it changes, in plain sentences.>

<Durable reasoning: the choices that shaped the change, each with its decision ID.>

- decisions/<slug>-<suffix>: <the claim in one line>

Plan: <plan link>

<!-- Include "Suggested code comment" only where the code cannot show the reason. -->

Suggested code comment (<path>:<line>): <the reason, as one comment line>
```

## Fill in

- `work`: the concept ID of the work item. `branch`: the branch of the work item's worktree, as the implementation notes record it.
- Subject: follow the conventional commit rules of **using-git**, or the convention the project's instructions set.
- Body: state the reasoning that stays true after the change merges. Leave out progress, task lists, test output and the criteria, which belong to the pull request description.
- Decision lines: one for each decision the change carries out, from the work item's `decisions` and the implementation notes. Cite the concept ID and state the claim in one line.
- Plan link: a URL only when the repository's `resource` is a browsable URL for files (an `https://` URL of a hosted repository); otherwise the repository name and the path `plans/<change>/plan.md`, for example `project: plans/greeting-locale/plan.md`. Never a file URL or a bare repository URL.
- Suggested code comment: a proposal for the implementer, never applied by the skill. Include it only where no name, type or test can show the reason.
- `generated.by` is `preparing-delivery/<version>`, with `<version>` read from the `version` field of the `plugin.json` of the plugin that ships this skill. `generated.at` is the output of `date -u +%Y-%m-%dT%H:%M:%SZ`; never invent or copy a time. The skill wrote the file, so leave out `verified`.
- A frontmatter value with a colon followed by a space, a quote or a leading special character, such as a `title` or `description`, is written in single quotes.
