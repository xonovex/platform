# Gap table template

Paste the block into the research file: `## Gap table` and `## Producer changes` go under `# Inference`. Replace every angle-bracket placeholder and delete any comment lines. Each fact in a row carries a footnote anchor that the file's `sources` entries define.

```markdown
## Gap table

<One sentence stating how many gaps need a decision and which one matters most.>

| Gap                     | Producer today                         | Target contract                         | Options                            | Recommendation                                 | Question        |
| ----------------------- | -------------------------------------- | --------------------------------------- | ---------------------------------- | ---------------------------------------------- | --------------- |
| <Gap in a short phrase> | <What the producer does>[^<source-id>] | <What the target accepts>[^<source-id>] | <Option 1>; <Option 2>; <Option 3> | Inference: <option and the reason in one line> | `<question-id>` |

<!-- Include the sentence below when a row has no question. -->

A row without a question needs no decision: <reason, for example both sides already agree>.

## Producer changes

<One sentence stating what the producer must change and what stays as it is.>

- <Change>: closes `<gap>`; touches <area of the code>.[^<source-id>]
- <Change>: closes `<gap>`; touches <area of the code>.[^<source-id>]

[^<source-id>]: <Source title>: <path>, lines <from> to <to>.
```

## Fill in

- One row for each gap between the producer's behavior and the target's contract. Order rows by how much a wrong choice costs.
- `Producer today` and `Target contract`: facts only, each with a footnote. Keep an explorer's inference out of these two columns.
- `Options`: at least two, separated by semicolons. Include the option of changing nothing when it is real.
- `Recommendation`: start with `Inference:` and keep the word wherever the recommendation is repeated.
- `Question`: the kebab-case `id` of the matching entry in the `questions` frontmatter list, and a matching row in `# Open decisions`. Leave the cell empty only for a row without a decision.
- `Producer changes`: a list of changes, each tied to a gap and to the code it touches. Do not order it into work or assign it to a team.
- Footnote labels (`<source-id>`) are the `id` values of the file's `sources` entries. Line ranges go in the footnote text.
