# Inline comment template

Copy the block to `plans/<change>/deliveries/<work-id>/comments/<suffix>.md`, where `<suffix>` is four random characters that use only `a-z` and `2-7` (never `0`, `1`, `8` or `9`). Replace every angle-bracket placeholder and delete any comment lines. Write a comment only when the delivery mode is `pull-request` and a decision shaped code that a reviewer would not understand from the code alone.

```markdown
---
type: Workflow Inline Comment
schema: workflow/inline-comment/1
title: <Comment label for the location>
description: <One sentence stating the reason the comment gives>
generated: {by: preparing-delivery/<version>, at: <timestamp>}
work: work/<slug>-<suffix>
path: <path/in/the/work/repository>
line: <line number>
side: right
marker: <stable-kebab-case-key>
label: decision
---

<The reason, in one to three sentences, with the decision ID when the label is decision.>
```

## Fill in

- `path` and `line`: the location in the current tree of the worktree. `side` is `right` for added or changed code and `left` for removed code.
- `marker`: a stable kebab-case key built from the topic, such as `retry-limit-source`. A rerun finds the file with the same marker and updates it, so never change a marker and never write two files with one marker.
- `label`: `decision` when the comment cites a decision, `note` for any other reason worth a reviewer's time.
- Body: state why the code is the way it is. Never restate what the code does and never say that the code is correct or safe.
- `generated.by` is `preparing-delivery/<version>`, with `<version>` read from the `version` field of the `plugin.json` of the plugin that ships this skill. `generated.at` is the output of `date -u +%Y-%m-%dT%H:%M:%SZ`; never invent or copy a time. The skill wrote the file, so leave out `verified`.
- A frontmatter value with a colon followed by a space, a quote or a leading special character is written in single quotes.
