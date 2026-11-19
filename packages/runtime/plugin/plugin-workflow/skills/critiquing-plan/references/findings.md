# Findings and the critique file

## Contents

- A finding
- Severity and confidence
- Where the file lives
- Frontmatter
- Body
- Writing and closing the run
- Critiquing a revised plan

## A finding

A finding is a numbered item with six fields. Numbering runs through both passes, so a number never changes meaning.

1. Short title of the problem
   - Severity: high, medium or low
   - Confidence: confirmed, likely or possible
   - Section: the plan heading it affects, such as `# Success criteria`
   - Failure mode: what goes wrong, for whom, and when
   - Evidence: the plan text, a decision concept ID or a repository anchor, one anchor per line, cited with a footnote
   - Direction: the change that would remove the problem, not the rewritten text

A finding with no failure mode or no evidence is vague. Send it back to the reviewer for the missing part, or drop it. Do not fill the gap yourself.

## Severity and confidence

| Severity | Meaning |
| --- | --- |
| high | A success criterion fails, work is unsafe, or the plan cannot be carried out |
| medium | A criterion is unproven, or a risk lacks an owner or a recovery |
| low | A gap in wording, ordering or traceability that a reader can work around |

| Confidence | Meaning                                                         |
| ---------- | --------------------------------------------------------------- |
| confirmed  | The evidence shows the failure mode in the plan or the code     |
| likely     | The evidence supports it and one step is inferred               |
| possible   | The reviewer could not reproduce it, or only a lens suggests it |

Label an inference as an inference wherever it is repeated.

## Where the file lives

`critiques/<slug>-<suffix>.md` in the change's plan directory, with `type: Workflow Critique` and `schema: workflow/critique/1`. The slug names the plan and the round, such as `fare-capping-round-one`. The slug and suffix follow the naming rules of **recording-handoff**. The file is never edited after it is written.

## Frontmatter

The common header is described by **recording-handoff**. The keys this type adds, validated by [../assets/schemas/critique.schema.json](../assets/schemas/critique.schema.json):

| Key | Required | Value |
| --- | --- | --- |
| `plan` | yes | The concept ID of the plan, `plan` |
| `lenses` | yes | A non-empty list of `red-team`, `pre-mortem`, `falsify` and `steelman` |
| `blind` | yes | `true` only when the first-pass reviewer received nothing but the plan and criteria |
| `sources` | no | One entry for each cited decision, file at a commit or web page, each with an `id` |

`generated.by` is `critiquing-plan/<version>`. The skill wrote the file, so it carries no `verified` entry and no `human:` actor.

## Body

In this order:

1. `# First pass, blind` (`# First pass, not blind` when `blind` is false): the first-pass findings, verbatim from the reviewer. When the review used more than one lens, the finding text says which lens produced it.
2. `# Second pass`: `## First-pass findings after context` with one verdict line for each first-pass number (confirmed, changed or refuted, with the reason), then `## New findings` with the findings the context added, numbered on from the first pass.
3. `# Could not check`: what the reviewer could not verify and why, or the sentence "Nothing was left unchecked."

Cite evidence with footnotes keyed to the `sources` ids, as research files do. Line ranges go in the footnote text. A decision cited by concept ID needs no source entry.

## Writing and closing the run

1. Report the findings inline, bottom line first: how many by severity, and the one that matters most.
2. Pick a slug and a random suffix and write the file.
3. Announce the write in one line: the path and the findings by severity.
4. Write one `Workflow Log Entry` from the template in **recording-handoff** with `skill: critiquing-plan`, `action: Creation` and the critique concept ID in `touched`.
5. Regenerate the indexes and the log with the **recording-handoff** script, validate the bundle and report any finding.
6. Do not commit or stage. Offer **revising-plan** and **settling-questions** and stop.

## Critiquing a revised plan

A critique that has been answered is not edited and does not count as the next review. Run the skill again for the revised plan: a new file, a new blind first pass that is not given the earlier critique, and a second pass that may cite the earlier file as a source.
