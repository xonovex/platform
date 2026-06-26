---
name: code-review-guide
description: Use when writing, structuring, or labelling code-review feedback on a pull or merge request — Conventional Comments labels (praise / nitpick / suggestion / issue / question / thought / chore / todo), the blocking vs non-blocking vs if-minor decorations, pairing one top-level summary with line-anchored inline comments, cross-linking the summary to its details instead of "see comment 3", and verifying each claim against an authoritative source before asserting it. Platform-independent review craft. Triggers on PR / MR review, leaving review comments, blocking / non-blocking, nit, severity, request-changes vs approve, or a review summary — even when the user doesn't say "code review".
---

# Code-review feedback — quick reference

How to write review comments that are unambiguous and actionable, independent of the platform they are posted on. This is the generic tier: for delivering these comments on a specific host (auth, line anchors, the posting API) load the matching host skill — `github-guide` (plugin `xonovex-skill-github`), `gitlab-guide` (plugin `xonovex-skill-gitlab`), or another `skill-<host>`.

The one rule to internalize: **every comment states its type and its urgency. A reader must never have to guess whether a comment blocks the merge.**

When this skill fires:

1. Decide the comment's type (is this a blocking issue, a suggestion, a nitpick, a question?) and its urgency, then lead with that label.
2. Verify the claim against the real source (the code on the branch, the API signature, the design-system definition) before asserting it.
3. Load the `references/*.md` file matching the question, not everything upfront.

## Essentials

- **Label every comment** - Conventional Comments format `label (decoration): subject`; the label categorizes (issue / suggestion / nitpick / question / praise / ...), see [references/conventional-comments.md](references/conventional-comments.md)
- **Set urgency with a decoration** - `(blocking)` must be fixed before merge, `(non-blocking)` is the author's discretion, `(if-minor)` resolve only if the change is small; absence = ambiguity, see [references/conventional-comments.md](references/conventional-comments.md)
- **One summary plus anchored inline comments** - a top-level summary carries the verdict and priority-ordered headlines; the detail and suggested code go inline on the exact line, see [references/review-structure.md](references/review-structure.md)
- **Cross-link, never cross-reference by number** - link the summary to its inline threads; a reader does not see your "comment 3" numbering, see [references/review-structure.md](references/review-structure.md)
- **Verify before you assert** - check each claim against the branch code / API / design-system source and cite `file:line`; never review from memory, see [references/review-structure.md](references/review-structure.md)
- **Lead with what works, then be specific** - praise is a label too; for every issue show the fix, not just the problem, see [references/review-structure.md](references/review-structure.md)
- **Carry findings in a shared contract** - the analyze → refine → post → resolve pipeline moves one canonical findings shape, anchored to new-file lines; host delivery (post / resolve) is the host skill's (`github-guide` / `gitlab-guide`), see [references/findings-schema.md](references/findings-schema.md), [references/review-analyze.md](references/review-analyze.md), [references/review-refine.md](references/review-refine.md)

## Gotchas

- Referencing another comment by ordinal ("see comment 3") is meaningless to the reader — they see no numbering. Link to it, or restate the point. If your host cannot link, keep each comment self-contained.
- Marking everything `(blocking)` dilutes the signal until nothing reads as urgent. Reserve blocking for genuine merge gates; most feedback is non-blocking.
- `question` and `thought` are inherently non-blocking and rarely need a decoration; adding `(blocking)` to a question is a contradiction.
- A "before merge" verdict belongs once in the summary, not repeated in every inline comment.
- Keep each comment self-contained, especially if you may edit it later. If comment A says "see comment B" and B is reworded or removed, A is left dangling.
- Suggested-code blocks must compile against the real APIs. A confidently wrong snippet is worse than prose; verify component / function signatures first (this is where a framework skill, e.g. a design-system guide, composes in).

## Example — a labelled inline comment, and a summary that links to it

Inline, on the exact line:

> **issue (blocking)**
>
> This overlay also fires on background auto-refresh, so it covers content the user is reading. The existing indicator already signals refresh, so drop the overlay:
>
> ```diff
> - if (isRefreshing && state is Success) { Box(...) { Loader() } }
> ```

Top-level summary, pointing at the detail rather than renumbering it:

> **Review summary**
>
> Direction is good. One thing before merge:
>
> 1. The reload overlay fires on background refresh and covers content. ([details](link-to-the-inline-comment))
>
> Remaining notes are inline and non-blocking.

## Progressive Disclosure

Each reference is a trigger — read it only when the user's intent matches; do not preload everything.

- Read [references/conventional-comments.md](references/conventional-comments.md) - Load when choosing a label or decoration, mapping severity to whether it gates merge, or wanting the full label list and the machine-parseable shape.
- Read [references/review-structure.md](references/review-structure.md) - Load when structuring a whole review (summary vs inline), ordering by priority, cross-linking comments, verifying claims before posting, or setting tone and matching a repo's writing style.
- Read [references/findings-schema.md](references/findings-schema.md) - Load when producing, refining, or publishing structured findings — the canonical findings JSON contract and new-file line anchoring.
- Read [references/review-analyze.md](references/review-analyze.md) - Load when turning a branch diff into structured findings (correctness then quality, anchoring, the effort dial, comparing against prior findings).
- Read [references/review-refine.md](references/review-refine.md) - Load when refining findings before posting (keep / reword / relabel / re-anchor / merge / split / drop, label discipline, stop-after-each-pass).
