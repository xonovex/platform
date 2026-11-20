# Writing rules

## Contents

- State settled things flat
- Mechanisms nobody agreed to
- Open decisions
- Links, not copies
- Unrecorded questions
- Sensitive findings
- The generated block
- The prose check

## State settled things flat

A settled decision becomes a plain statement of the target, in the objective or the scope, with no trace of the argument.

- Flat: "Invoices reach finance through the feed."
- Not flat: "After weighing a file drop against a feed, and because the file drop cannot be retried, invoices reach finance through the feed."

The reasoning and the rejected options live in the decision file's `# Reason` and `# Options considered`. The plan links the file in its generated block and, for a variant, in the `Decided by` column. Hedges ("probably", "we expect") do not belong in settled text; if you need one, the decision is not settled.

## Mechanisms nobody agreed to

The plan names only mechanisms that a settled decision states or that the request states. A natural-looking design the research suggested is an option, and an option lives in a decision file. If no decision covers it and the plan needs it, report it as an unrecorded question; do not write it into the objective, the scope or a criterion.

## Open decisions

An open, paused or partly-settled decision is not summarized, paraphrased or weighed in the plan. It appears as a link in the generated block, as a link in a `Decided by` cell, and as a name in the `waits_on` of the work that depends on it. A reader who wants the question follows the link.

## Links, not copies

- Do not copy a `# Decision` sentence, a table of options or a work objective into the plan. State the effect in the plan's own words and link the file.
- Do not copy a criterion of a work file into the success criteria. Success criteria describe the observed result of the whole change; a work criterion describes one piece.
- Do not copy research prose. Cite the research file as a source and summarize in a few lines.

A quick test: delete the link and ask whether the sentence still says something the linked file does not. If it only repeats the file, remove it.

## Unrecorded questions

A research question that no decision file covers has no concept ID, so nothing can wait on it. Name each in the reply, write no work file that depends on it, and offer **settling-questions** to record it. Never create the decision file in this skill.

## Sensitive findings

A finding about a credential, a security weakness, a personal data exposure or an internal address stays out of `plan.md`, the work files and the log entry. The plan can say that a security question exists, without the details, when a reader needs to know. In the reply, name the topic of each finding kept out and say where the research file holds it, so the invoker decides who sees it.

## The generated block

Rewrite the whole block from the files that exist, with the layout in [plan-sections.md](plan-sections.md). A block that someone edited by hand is replaced as a whole, not merged.

## The prose check

Find the project's prose check in its instructions, such as a text validator or a style linter, and run it on `plan.md` and the work files before you report. Fix each reported finding and run it again. If the project has none, say so in the reply. Whichever applies:

- Plain sentences; a comma, colon or full stop instead of an em dash; straight quotes.
- One paragraph or list item on one line, never hard wrapped.
- Active voice, one term for each concept, no invented labels or shorthand for names.
