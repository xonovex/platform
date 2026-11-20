---
name: reviewing-code
description: "Use when reviewing the code of a pull request, merge request or diff for defects, or reviewing an existing module or file that is not a change, and writing the review feedback. Triggers on 'review PR 418', 'review this module', code review, review comments, Conventional Comments labels, nit, blocking vs non-blocking, request changes vs approve, or a review summary with inline comments, even when the user doesn't say 'code review'."
---

# Code-review feedback: quick reference

Host-independent code review: find the defects in a change (a pull request, a merge request or a diff) or in existing code that is not a change, then write the feedback. **Every comment states its type and its urgency: a reader must never guess whether it blocks the merge.** Review operations return findings inline and never post, edit, resolve, or persist them. A separate Publish operation may select a matching host skill for delivery. Fetching a pull request, reading its threads and posting comments belong to the host skill; preparing your own pull request belongs to **authoring-pull-requests**.

## Essentials

- **Label every comment** - Conventional Comments format `label (decoration): subject`; the label categorizes (issue / suggestion / nitpick / question / praise / ...), see [references/conventional-comments.md](references/conventional-comments.md)
- **Set urgency with a decoration** - `(blocking)` must be fixed before merge, `(non-blocking)` is the author's discretion, `(if-minor)` resolve only if the change is small; absence = ambiguity, see [references/conventional-comments.md](references/conventional-comments.md)
- **One summary plus anchored inline comments** - a top-level summary carries the verdict and the headline issues; the detail and suggested code go inline on the exact line, see [references/review-structure.md](references/review-structure.md)
- **Cross-link, never cross-reference by number** - link the summary to its inline threads; a reader does not see your "comment 3" numbering, see [references/review-structure.md](references/review-structure.md)
- **Verify before you assert** - check each claim against the branch code / API / design-system source and cite `file:line`; never review from memory, see [references/review-structure.md](references/review-structure.md)
- **Lead with what works, then be specific** - praise is a label too; for every issue show the fix, not just the problem, see [references/review-structure.md](references/review-structure.md)
- **Review existing code like a change** - with no diff, analyze the named files or module the same way and anchor each finding to its current line, see [references/review-analyze.md](references/review-analyze.md)
- **Carry findings in a shared contract** - analyze and refine return one canonical findings shape, anchored to new-file lines; later Publish or Execute operations may pass it to a host skill, see [references/findings-schema.md](references/findings-schema.md), [references/review-analyze.md](references/review-analyze.md), [references/review-refine.md](references/review-refine.md)

## Gotchas

- Referencing another comment by ordinal ("see comment 3") is meaningless to the reader. They see no numbering. Link to it, or restate the point. If your host cannot link, keep each comment self-contained.
- Marking everything `(blocking)` dilutes the signal until nothing reads as urgent. Reserve blocking for genuine merge gates; most feedback is non-blocking.
- `question` and `thought` are inherently non-blocking and rarely need a decoration; adding `(blocking)` to a question is a contradiction.
- A "before merge" verdict belongs once in the summary, not repeated in every inline comment.
- Keep each comment self-contained, especially if you may edit it later. If comment A says "see comment B" and B is reworded or removed, A is left dangling.
- Suggested-code blocks must compile against the real APIs. A confidently wrong snippet is worse than prose; verify component / function signatures first (this is where a framework skill, e.g. a design-system guide, composes in).
- AI-generated findings and AI-written code are not self-verifying. The human reviewer must actively evaluate each, not rubber-stamp; adversarial AI review complements, never replaces, human judgement on the merge decision.
- Loading a host skill during Review to post or resolve findings silently changes the operation and is out of scope.

## Example: a labelled inline comment, and a summary that links to it

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
> 1. The reload overlay fires on background refresh and covers content. ([details](https://example.com/reviews/123#comment-1))
>
> Remaining notes are inline and non-blocking.

## Progressive Disclosure

Each reference is a trigger: read it only when the user's intent matches; do not preload everything.

- Read [references/conventional-comments.md](references/conventional-comments.md) - Load when choosing a label or decoration, mapping severity to whether it gates merge, or wanting the full label list and the machine-parseable shape.
- Read [references/review-structure.md](references/review-structure.md) - Load when structuring a whole review (summary vs inline), ordering by priority, cross-linking comments, verifying claims before posting, or setting tone and matching a repo's writing style.
- Read [references/findings-schema.md](references/findings-schema.md) - Load when producing or refining structured findings: the canonical findings JSON contract and new-file line anchoring.
- Read [references/review-analyze.md](references/review-analyze.md) - Load when turning a branch diff into structured findings (correctness then quality, anchoring, the effort dial, comparing against prior findings).
- Read [references/review-refine.md](references/review-refine.md) - Load when refining findings before posting (keep / reword / relabel / re-anchor / merge / split / drop, label discipline, stop-after-each-pass).
