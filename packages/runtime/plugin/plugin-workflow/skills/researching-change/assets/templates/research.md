# Research template

Copy the block to `plans/<change>/research/<slug>-<suffix>.md`, replace every angle-bracket placeholder and delete any comment lines. A research file is never edited after it is written; a correction is a new file.

```markdown
---
type: Workflow Research
schema: workflow/research/1
title: <Research title as a short noun phrase>
description: <One sentence stating what was researched and the main finding>
generated: {by: researching-change/<version>, at: <timestamp>}
subject: <One sentence naming what was researched>
kind: <migration, feature, quality, investigation or project>
roles:
  - <role-name>
questions:
  - id: <question-id>
    text: <One open decision as a question>
corrects:
  - research/<slug>-<suffix>
skills_to_consult:
  - <installed-skill-name>
sources:
  - id: <source-id>
    resource: <repository-name>@<commit>:<path>
    title: <What the file is>
  - id: <page-id>
    resource: <https://example.org/page>
    title: <Page title>
---

# Bottom line

<The answer in a few sentences, with the largest gap or risk.>

# Observed

## Repositories read

- <repository-name> at <commit>: `<command used to check it>`

## <Role name>

- <Fact.>[^<source-id>]

<!-- Include "Re-checked" when you re-verified claims yourself. -->

## Re-checked

- <Claim>: checked on <which side> with `<command>` or at <anchor>.

# Inference

- Inference: <Conclusion>, from <the facts it rests on>.

# Unknowns

- <What no read answered, or a repository that is not cloned.>

# Open decisions

| Id              | Question                  | Options                  | Recommendation                |
| --------------- | ------------------------- | ------------------------ | ----------------------------- |
| `<question-id>` | <The question to settle.> | <Option one; option two> | <Inference: option, and why.> |

<!-- Include "Corrections" when an existing page is wrong or stale. -->

# Corrections

- <Page name>: <what it says, and what the code or another page shows instead.>

# Not checked

- <What was not verified at runtime, or only read in code.>

[^<source-id>]: <Source title>, lines <from> to <to>

[^<page-id>]: <Page title>
```

## Fill in

- `generated.by`: `researching/<version>`. The skill wrote the file, so leave out `verified`.
- `subject`: one sentence. `kind` is a short label and may be left out.
- `roles`: the names of the explorer roles used. Leave it out for a research that used no explorer.
- `questions`: one entry for each open decision, with a kebab-case `id`. The list may be empty, but then leave out the rows of `# Open decisions` and say so there.
- `corrects`: include only when this file corrects an earlier research file. Values are concept IDs without `.md`.
- `skills_to_consult`: include when the research found work in the area of an installed skill that the later steps should load.
- `sources`: one entry for each cited file at the commit it was read at, and one for each web page. Each `id` is a kebab-case key that a footnote label repeats. Delete the key and the footnote lines when nothing is cited.
- `# Observed`: one subheading for each role. Every fact carries a footnote. The footnote text starts with the source title and lists the line ranges cited.
- `# Inference`: keep the label `Inference` on every conclusion, in every place it is repeated.
- `# Open decisions`: one row for each `questions` id.
- `# Corrections`: include when any existing page needs one, naming each page.
- `# Not checked`: always include it. Write "Nothing was run" when no command ran beyond reading.
