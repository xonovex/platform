# Dispositions template

Use this for the body of the run's `Workflow Log Entry`. Copy the frontmatter from the shared log entry template of **recording-handoff** with `skill: revising-plan` and `action: Update`, or `Creation` when the run wrote a decision, list every created or changed concept ID in `touched`, and replace every angle-bracket placeholder in the block below. The same table goes into the reply.

```markdown
<One sentence stating how many items were processed and what changed.>

| Item                                         | Disposition                | Rationale                                      |
| -------------------------------------------- | -------------------------- | ---------------------------------------------- |
| `<source>`: "<The note, quoted as written.>" | Applied                    | <Why, naming the files that changed.>          |
| `<source>`: "<The note, quoted as written.>" | Partly applied             | <What changed, what did not, and why.>         |
| `<source>`: "<The note, quoted as written.>" | Not applied                | <The evidence or owner that stops it.>         |
| `<source>`: "<The note, quoted as written.>" | Sent to settling-questions | <The question and the fact or owner it needs.> |
```

## Fill in

- The first line of the body is the one-sentence summary. The table follows after a blank line.
- `<source>`: a path and line (`plan.md:41`), or a concept ID and number (`critiques/<slug>-<suffix>` finding 2), or the concept ID of a research file. Quote the original text of a note or the summary of a finding in full, on one line, with each pipe character written `\|`.
- Leave a sensitive finding out of the quote and write "withheld" in its place; describe it in the reply only.
- `Disposition` is exactly one of `Applied`, `Partly applied`, `Not applied` or `Sent to settling-questions`.
- The rationale is one sentence that names a file, a criterion or the evidence.
- Rows follow the order of the enumeration, one row for each item and none for an item that was not enumerated.
