# Producer brief template

Copy the block into the prompt of the read-only explorer for the current producer, replace every angle-bracket placeholder and delete any comment lines. One brief covers one producer.

```markdown
You are a read-only explorer for the migration of <change in one sentence>. Do not write, edit or delete any file.

Question: what does <producer> send today, through which path, and what does it do with the answer?

Why it matters: every change the migration forces on the producer starts from this behavior, and the gap table compares it with what the target accepts.

Repositories to read:

- <repository root>: <what it holds>

Read for:

- The send path and the one seam every send goes through, and any place that bypasses it.
- The payload and every field the consumers read.
- The stored identifiers and the tables or indexes that hold them.
- How the response is handled, including what triggers a sign-off, a deletion or a state change.
- Retry behavior and what a failure does to the caller of the producer.
- Configuration for each environment and where its credentials are kept.

<!-- Include "Also read" when the request names more of the producer to read. -->

Also read:

- <Extra area, for example a batch job or a scheduled task>

Out of scope: <the target and the consumer apps; other explorers cover them>.

Required output:

- An `Observed` heading: facts only, each followed by one anchor.
- An `Inference` heading: conclusions that the facts support, each labelled as inference.
- An `Unknowns` heading: what you could not establish and why.
- One anchor per line as repository root plus path and line, for example `<repository root>/<path>:<line>`. Never wrap an anchor over two lines.
- The commit of each repository you read, and the commands you ran, one per line.
```

## Fill in

- `<change in one sentence>` and `<producer>`: use the names from the request, for example "moving the order publisher from the legacy relay to the delivery service".
- `Repositories to read`: one line for each repository, as the root path. A repository that is not cloned is reported to the invoker and left out of the list.
- `Also read`: include when the request points at a part of the producer the list does not name. Delete the heading and the list otherwise.
- `Out of scope`: name the other roles so the explorer does not duplicate their work.
- Keep the required output block unchanged. The research file depends on its headings and its anchors.
