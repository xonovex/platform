# The research file

## Contents

- Where it lives
- Frontmatter
- Body headings
- Open decisions
- Corrections
- Writing and closing the run
- Correcting an earlier file

## Where it lives

`research/<slug>-<suffix>.md` in the change's plan directory, with `type: Workflow Research` and `schema: workflow/research/1`. The slug and the suffix follow the naming rules of **recording-handoff**. The file is written once and never edited.

Write the report inline first, bottom line first. Save the file with the same content once the report is complete.

## Frontmatter

The common header is described by **recording-handoff**. The keys this type adds, validated by [../assets/schemas/research.schema.json](../assets/schemas/research.schema.json):

| Key | Required | Value |
| --- | --- | --- |
| `subject` | yes | One sentence naming what was researched |
| `kind` | no | A short label such as `migration`, `feature`, `quality`, `investigation` or `project` |
| `roles` | no | The names of the explorer roles used |
| `questions` | yes | A list, possibly empty, of `{ id, text }`; the `id` is kebab-case and each question is an open decision |
| `corrects` | no | Concept IDs of earlier research files this file corrects, such as `research/order-flow-4kq7` |
| `skills_to_consult` | no | Installed skill names that the later steps should load, because the research found work in their area |
| `sources` | no | One entry for each cited file at a commit and one for each web page; see [evidence.md](evidence.md) |

`generated.by` is `researching/<version>`. The skill wrote the file, so it carries no `verified` entry and no `human:` actor.

## Body headings

In this order:

1. `# Bottom line`: the answer in a few sentences, with the largest risk or gap.
2. `# Observed`: facts with a footnote anchor, grouped under a subheading for each role. Starts with a `## Repositories read` list of repository, commit and the command used. Ends with `## Re-checked` when claims were re-verified.
3. `# Inference`: conclusions, each labelled and tied to the facts it rests on.
4. `# Unknowns`: what no read answered, and the gaps such as a repository that is not cloned.
5. `# Open decisions`: see below.
6. `# Corrections`: include when any existing page is wrong or stale.
7. `# Not checked`: what was not verified at runtime, and what was read only in code.

## Open decisions

A table whose rows repeat the `questions` ids, with the options and, when the facts support one, a recommendation labelled as an inference. These ids are what **settling-questions** picks up, so keep each question to one decision. The table and the `questions` list always agree: every id appears in both.

| Id | Question | Options | Recommendation |
| --- | --- | --- | --- |
| `cap-scope` | Does the daily cap include night fares? | Include them; exclude them | Inference: include them, as the pricing rules do |

## Corrections

Name each existing page and say what is wrong and what the code or the other page says instead. A correction here changes no page. Offer it as a next step, and the owner of that page makes the change.

## Writing and closing the run

1. Pick a slug from the subject and a random suffix, and write the file.
2. Announce the write in one line: the path and the number of open questions.
3. Write one `Workflow Log Entry` from the template in **recording-handoff** with `skill: researching`, `action: Creation` and the research concept ID in `touched`.
4. Regenerate the indexes and the log with the script of **recording-handoff**, then validate the bundle and report any finding.
5. Do not commit or stage. Offer the next steps and stop.

## Correcting an earlier file

A research file that turns out wrong is not edited. Write a new file whose `corrects` lists the concept ID of the old one, restate only what changed in its `# Corrections`, and cite the evidence. The old file stays as written; readers follow `corrects` from the new file.
