# Explorer brief template

Copy the block into the prompt of one explorer, replace every angle-bracket placeholder and delete any comment lines. Write one brief for each role. The explorer has only this brief, so name every repository, commit and document it needs.

```markdown
# Role: <role name>

## Question

<What to find out, in one or two sentences.>

## Why it matters

<The decision or the plan this answer informs.>

## What to read

- <repository-name> at <commit>: <paths or areas>
- <document or web page>

<!-- Include "Out of bounds" when part of the material must not be read or relied on. -->

## Out of bounds

- <Material not to read, or not to rely on>

## Required output

Return three headings and nothing else.

- `Observed`: facts that a file, a command output or a page shows. End each with one anchor on its own line, as the repository root plus the path and the line, for example `shop-api/src/orders/OrderPublisher.kt:42`. One anchor per line, never wrapped.
- `Inference`: conclusions you draw from the facts. Label each as an inference and name the facts it rests on.
- `Unknowns`: what you could not determine, and why.

## Boundary

Read only. Write no file, run no command that changes state, and propose no fix. Report what you find, not what you would change.
```

## Fill in

- `Role`: the name the research file lists under `roles`.
- `Question` and `Why it matters`: a question the role can answer from its own material, and the reason in one sentence. A brief without a reason gets a broad answer.
- `What to read`: every repository with the commit it was read at, and every document. Give the explorer the sources; do not make it search for them.
- `Out of bounds`: include when some material is stale, owned by another role, or known to be wrong.
- `Required output`: keep it as written, so every report has the same shape and can be merged.
- Model and effort belong to the call that starts the explorer, not to the brief.
