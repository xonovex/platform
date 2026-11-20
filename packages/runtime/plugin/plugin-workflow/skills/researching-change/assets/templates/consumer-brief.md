# Consumer brief template

Copy the block into the prompt of the read-only explorer for one consumer app, replace every angle-bracket placeholder and delete any comment lines. Write one brief for each app: a mobile client and a web client get two briefs.

```markdown
You are a read-only explorer for the migration of <change in one sentence>. Do not write, edit or delete any file.

Question: what does <consumer app> do with the old and the new message shape, and which versions are in use?

Why it matters: the consumers decide when the producer can switch, and the gap table needs what each app can handle.

Repositories to read:

- <repository root>: <what it holds>

Read for:

- Which identifiers the app knows and which it sends on every call.
- How it handles the old and the new message shape, for each version.
- What the user's action on a message does, and where that route is declared.
- What happens on a failed update, and when it registers again.
- The versions that matter and the release cadence, so the supported versions can be computed.

<!-- Include "Release data" when the repository holds release notes, tags or a release pipeline. -->

Release data:

- Read <path or tag pattern> for the released versions and their dates.

Out of scope: <the producer, the target and the other consumer apps; other explorers cover them>.

Required output:

- An `Observed` heading: facts only, each followed by one anchor.
- An `Inference` heading: conclusions that the facts support, each labelled as inference.
- An `Unknowns` heading: what you could not establish and why.
- One anchor per line as repository root plus path and line, for example `<repository root>/<path>:<line>`. Never wrap an anchor over two lines.
- The commit of each repository you read, and the commands you ran, one per line.
```

## Fill in

- `<change in one sentence>` and `<consumer app>`: use the names from the request, for example "moving the order publisher from the legacy relay to the delivery service" and "the mobile client".
- `Repositories to read`: one line for each repository the app is built from, as the root path. A repository that is not cloned is reported to the invoker and left out of the list.
- `Release data`: include when the repository holds release notes, tags or a release pipeline. Delete the heading and the list otherwise, and the explorer lists the versions as unknown.
- `Out of scope`: name the other roles so the explorer does not duplicate their work.
- Keep the required output block unchanged. The research file depends on its headings and its anchors.
