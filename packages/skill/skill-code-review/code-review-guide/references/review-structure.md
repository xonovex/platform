# Structuring a review — summary, inline, links, verification

How to assemble a whole review so it is easy to act on, independent of host.

## Summary plus inline

Post two kinds of comment:

- **One top-level summary.** State the verdict in the first line, then the headline issues in priority order (most important first). Keep it short — it is the map, not the territory. Acknowledge what is good.
- **Line-anchored inline comments.** Each carries one labelled point, the reasoning, and a suggested fix anchored to the exact line it concerns. Detail lives here, not in the summary.

This split lets a busy author read the summary for the verdict and dive into only the inline threads that matter.

## Order by priority, not by file

In the summary, lead with what would block or most improve the change. A reader should be able to stop after the first two bullets and know the important things. Do not order by file position or by the sequence you happened to find them.

## Cross-link, do not renumber

When the summary references a detailed point, link to the inline comment rather than writing "see comment 3":

- The reader sees no ordinal numbering, so "comment 3" is noise.
- Links survive reordering; ordinals do not.
- If two comments depend on each other, link them both ways.
- If the host has no comment-linking, keep each comment fully self-contained instead — never leave a dangling "see above".

When you edit a comment later, re-check any comment that linked to it.

## Verify before you assert

Never review from memory. For every factual claim:

- Read the code **on the branch under review**, not your recollection or the base branch.
- Check API and component signatures against their source before suggesting a call — a confidently wrong snippet costs the author more than no snippet. This is where framework skills compose in: load the design-system / library skill to confirm the real component name, parameters, and accessibility contract before you write the fix.
- Cite `file:line` so the author can jump straight to it.
- Distinguish what you verified from what you suspect. If unsure, use `question`, not `issue`.

## Tone and house style

- Lead with what works. A review that is only negatives reads as hostile and hides the signal.
- Be specific and actionable: show the fix, not just the problem.
- Frame as suggestions, not commands ("consider", "suggest"), and let the label carry the urgency.
- Match the repo's writing conventions (commit/PR prose rules often apply to review comments too — check the project instructions for the repo's prose rules).
- Right-size the review to the change. A two-line fix does not need a six-section summary.

## Self-contained comments

Each comment should make sense alone. Avoid "as I said above" or "same as the other one" — restate or link. This matters most when comments may be edited or resolved independently, where surrounding context disappears.
