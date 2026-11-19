# The queue and the brief

## Collecting the questions

Take them from three places and merge duplicates:

- The `questions` of the research files, each `{ id, text }`. The `# Open decisions` table of the file holds the options the research found.
- Decision files whose `state` is `open`, `paused` or `partly-settled`. A paused one returns to the queue when its named fact is known.
- Questions the operator raises, in words or in a pasted proposal, including the answer an owner gave to an open decision.

Skip a question that a settled decision already answers, and say so in one line. A research file that corrects an earlier one (`corrects`) replaces the earlier file's questions. When no question is known, discover them as [discovery.md](discovery.md) says.

## Ordering the queue

Order by dependency: a scoping question comes first, because its answer prunes or reshapes the rest. Within a level, put the question the most others wait on first. State the order and the reason in one message before the first brief, as a numbered list of titles. The numbers are display numbers for the conversation; they never go into a file name or frontmatter.

A question that appears mid-walk goes to the end of the queue. Renumbering mid-walk makes earlier answers hard to cite.

## One brief for each message

Use [../assets/templates/decision-brief.md](../assets/templates/decision-brief.md). A brief has four parts, and the message ends after the recommendation:

1. **Files involved**: a table of the files, documents or systems the question touches, with what each does today. Read them first.
2. **Why it is not obvious**: the constraint that makes the question hard, and what the source assumes and whether the code agrees.
3. **Options**: two to four lettered options in a table with what each gains and costs. Include the literal reading of the leading source as one of them.
4. **Recommendation**: one pick and the trade-off it accepts.

Close by naming the answers the operator may give: answer with an option, leave it open for its owner, or pause until a named fact is known. This keeps the honest open state on offer.

Write the brief as prose and tables, not as a question tool with option chips. The operator may answer in part, add a constraint, or redefine the question, and chips would lose that.

## Options

- Each option is a design the operator could pick, stated in one sentence, not a slogan.
- Gains and costs are concrete: what changes for a user, a team, a release or a repository.
- Do not pad to reach four options. Two honest options beat four with two invented.
- An option the code rules out is not offered. Say in the "why it is not obvious" part that it was excluded and why.

## The operator's own alternatives

When the operator proposes something the brief did not list:

1. Add it as the next letter. Keep their wording.
2. Check it against the code or the design record before you comment on it. A proposal to remap identifiers at send time is checked by reading what the registry can look up.
3. State what you found in a few lines: what it needs, what it costs, or why it is not possible.
4. If it is possible, keep it as a candidate. If it is not possible, say why and keep it out of the record.

A recommendation may change after the check. Say so.

## The recommendation

- Name one option, never "it depends".
- Name the trade-off it accepts, in one sentence: "This costs one extra release, and avoids a second code path".
- Label it as your recommendation. When the facts do not support one option, say what fact would, and offer the paused state.
