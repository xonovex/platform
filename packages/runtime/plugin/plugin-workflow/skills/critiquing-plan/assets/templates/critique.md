# Critique template

Copy the block to `plans/<change>/critiques/<slug>-<suffix>.md`, replace every angle-bracket placeholder and delete any comment lines. A critique file is never edited after it is written; a critique of a revised plan is a new file.

```markdown
---
type: Workflow Critique
schema: workflow/critique/1
title: <Critique title as a short noun phrase>
description: <One sentence stating what was reviewed and the most serious finding>
generated: {by: critiquing-plan/<version>, at: <timestamp>}
plan: plan
lenses:
  - red-team
  - pre-mortem
blind: true
sources:
  - id: <source-id>
    resource: <repository-name>@<commit>:<path>
    title: <What the file is>
---

<!-- Use "# First pass, not blind" instead when blind is false. -->

# First pass, blind

1. <Short title of the problem>
   - Severity: <high, medium or low>
   - Confidence: <confirmed, likely or possible>
   - Section: <plan heading, such as `# Success criteria`>
   - Failure mode: <What goes wrong, for whom, and when.>
   - Evidence: <Plan text, decision concept ID or anchor.>[^<source-id>]
   - Direction: <The change that would remove the problem.>

# Second pass

## First-pass findings after context

1. Confirmed: <Why the decisions and references support it.>
2. Changed: <What changed in severity, confidence, section or direction, and why.>
3. Refuted: <Where the context answers it.>

## New findings

4. <Short title of the problem>
   - Severity: <high, medium or low>
   - Confidence: <confirmed, likely or possible>
   - Section: <plan heading>
   - Failure mode: <What goes wrong, for whom, and when.>
   - Evidence: <Decision concept ID, plan text or anchor.>
   - Direction: <The change that would remove the problem.>

# Could not check

- <What the reviewer could not verify, and why.>

[^<source-id>]: <Source title>, lines <from> to <to>
```

## Fill in

- `generated.by`: `critiquing-plan/<version>`. The skill wrote the file, so leave out `verified`.
- `plan`: the concept ID of the plan, `plan`.
- `lenses`: the lenses the reviewer ran, from `red-team`, `pre-mortem`, `falsify` and `steelman`. The default is the first two.
- `blind`: `true` only when the first-pass reviewer received nothing but the plan, the review criteria and the lenses. Otherwise `false`, and the first heading reads `# First pass, not blind`.
- `sources`: one entry for each cited file at the commit it was read at and one for each web page. Delete the key and the footnote lines when nothing is cited. A decision cited by concept ID needs no entry.
- `# First pass`: the reviewer's findings, verbatim and numbered from 1. Every finding has the six fields.
- `# Second pass`: one verdict line for every first-pass number, then the new findings numbered on from the last first-pass number. Delete `## New findings` when the context added none and say so in the section.
- `# Could not check`: always include it. Write "Nothing was left unchecked." when that is true.
