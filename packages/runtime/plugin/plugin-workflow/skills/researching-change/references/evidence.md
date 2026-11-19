# Evidence

## Contents

- Checking a repository
- Observed and inferred
- Anchors
- Sources and footnotes
- Pointers outside the default branch
- Re-verifying the claims that matter

## Checking a repository

Before an explorer reads a repository, confirm it is cloned and note what it is at:

```bash
for r in <repo-dir>...; do git -C "$r" log -1 --format='%h %cs' 2>/dev/null; git -C "$r" status -sb 2>/dev/null | head -1; done
```

- A directory that is missing is a repository that is not cloned. Report it under `# Unknowns` as a gap and plan no claim about it.
- A short commit and a clean branch line go into the file as the commit the facts were read at.
- Uncommitted edits in a repository that is read are named in the file, since an explorer would read them as facts.
- Keep the commands you ran. A command collapsed into "checked freshness" cannot be reconstructed later.

## Observed and inferred

- **Observed** is what a file, a command output or a page says. It carries an anchor.
- **Inference** is a conclusion drawn from facts. It keeps its label in every place it is repeated, in the bottom line and in the tables too. "Probably" is not a label; the word `Inference` or a marked column is.
- **Unknowns** are the questions no read answered. Never fill one with a guess.
- A claim from a design document that the code does not show is not a fact. Report what the code does, and list the document under `# Corrections`.

## Anchors

One anchor per line as repository root plus path and line: `shop-api/src/orders/OrderPublisher.kt:42`. A wrapped path cannot be spot-checked, so a line holds exactly one anchor and nothing else wraps it.

## Sources and footnotes

The research file cites evidence through its `sources` list:

- One entry for each cited file, at the commit it was read at. `resource` is `<repository>@<commit>:<path>`, `id` is a kebab-case key, and `title` says what the file is.
- One entry for each web page. `resource` is the URL.
- The body cites a source with a footnote whose label is the `id`. One footnote definition for each source: its text starts with the source title and lists the line ranges cited, for example `Order publisher, lines 40 to 72 and 118 to 130`.

Add an entry only for a source the body cites, and give every cited source a footnote definition.

## Pointers outside the default branch

Explorers read the default branch. When a report points at material elsewhere, such as design notes on another branch, an unmerged change or a ticket, read it yourself, and label the branch or the page in the source entry. Material outside the default branch is evidence of intent, not of shipped behavior; say which in the file.

## Re-verifying the claims that matter

Pick the few claims that would change a decision: a surprising one, one that two explorers disagree on, or one a document contradicts. Check each in the code or the page yourself and say which side you checked. Put them under a `## Re-checked` heading in `# Observed`, with the command or the anchor. Do not delegate this step, and do not extend it to every claim.
