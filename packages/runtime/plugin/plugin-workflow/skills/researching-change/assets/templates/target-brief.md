# Target brief template

Copy the block into the prompt of the read-only explorer for the target, replace every angle-bracket placeholder and delete any comment lines. One brief covers one target.

```markdown
You are a read-only explorer for the migration of <change in one sentence>. Do not write, edit or delete any file.

Question: what does <target> accept, and what does it do when the input is wrong?

Why it matters: the gap table compares this contract with what the producer sends today.

Repositories to read:

- <repository root>: <what it holds>

Read for:

- The accepted recipient types and what each one needs.
- Authentication, and whether anything restricts a caller to a channel.
- What it does with a rejected recipient, an unknown recipient and a switched-off channel, and whether any of it reaches the caller.
- Whether documented behavior exists in the code, for example persistence or deduplication on an idempotency key. Report what the code does and name any document that claims more.
- Capabilities removed in history and capabilities on unmerged branches, with the branch of each.
- Neighbor repositories that call it, as reference callers.

<!-- Include "Design record" when the target does not exist yet. -->

Design record:

- The target is not built. Read <contract or design record path> instead and say in the answer that no running target was read.

Out of scope: <the producer and the consumer apps; other explorers cover them>.

Required output:

- An `Observed` heading: facts only, each followed by one anchor.
- An `Inference` heading: conclusions that the facts support, each labelled as inference.
- An `Unknowns` heading: what you could not establish and why.
- One anchor per line as repository root plus path and line, for example `<repository root>/<path>:<line>`. Never wrap an anchor over two lines.
- The commit of each repository you read, and the commands you ran, one per line.
```

## Fill in

- `<change in one sentence>` and `<target>`: use the names from the request, for example "moving the order publisher from the legacy relay to the delivery service".
- `Repositories to read`: one line for each repository, including the neighbor repositories that call the target, as the root path. A repository that is not cloned is reported to the invoker and left out of the list.
- `Design record`: include when the target does not exist yet. Delete the heading and the list otherwise.
- `Out of scope`: name the other roles so the explorer does not duplicate their work.
- Keep the required output block unchanged. The research file depends on its headings and its anchors.
