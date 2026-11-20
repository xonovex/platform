---
name: settling-questions
description: "Use when the open questions of a change need settling one at a time with the operator: the questions of a research file, open, paused or partly settled decision files, or questions found by questioning. Presents one brief per message with lettered options and a recommendation, checks the operator's alternatives in the code, records each answer as one testable claim in a decision file (settled, partly settled, paused or left open), then consolidates the direction. Triggers on 'walk me through the open questions', 'interview me one question at a time', a new option to check before a recommendation, pausing a question until a fact is known, or an answer an owner brings back, even when the user doesn't say 'decision'. Not for researching facts, writing or revising the plan, or applying written feedback."
---

# Settle Open Questions

Purpose: turn the open questions of a change into recorded decisions the plan can state flat, without forcing an answer the operator wants an owner to validate. The output is one decision file for each answer the operator gives, and a consolidated reply at the end. The skill writes no plan and no work file.

**Load first:** **recording-handoff**, and run its entry step before anything below.

## Essentials

- One brief for each message, then stop. The brief names the question, the options with what each gains and costs, and one honest recommendation with the trade-off it accepts. Nothing of the next question appears in the same message.
- Nothing is written before the operator answers. A write is announced in one line.
- An answer is recorded as one testable claim: "When <trigger>, the system shall <response>". It has one of four states: settled, partly settled, paused until a named fact is known, or left open for its owner to validate.
- A question that needs facts goes back to the code or to research. The skill never guesses a fact.
- A decision file is never edited. An answer that changes an existing decision is a new file that supersedes it.
- After each record, say which later questions the answer merges, splits, reorders or resolves, and adjust the queue.
- Never edit the plan or the work files, never commit or stage, never start the next step.

## Procedure

1. **Run the entry step** of **recording-handoff**. The decision files are written into the plan directory's repository, so confirm that branch. The step's own precondition is the change directory. Having no known question is not a blocker: the walk then discovers them.
2. **Load the skills that apply** from the sources that the settings reference of **recording-handoff** lists, including the `settling-questions` entry of the project's `skills` map. This skill names none itself.
3. **Collect the questions** from the `questions` of the research files, from decision files whose `state` is `open`, `paused` or `partly-settled`, and from the operator. Read the research files and the settled decisions they touch yourself. When none is known, discover them by questioning: frame the subject, walk the design tree from the top-level intent, one branch at a time. See [references/discovery.md](references/discovery.md).
4. **List the queue up front** in one message, in dependency order, with the reason for the order: a scoping answer comes first because it reshapes the rest. Then present the first brief and stop. See [references/brief.md](references/brief.md).
5. **Present one brief** from [assets/templates/decision-brief.md](assets/templates/decision-brief.md). Check every fact in it against the code or the design record first. When a fact is missing, go back to the code or to research as [references/discovery.md](references/discovery.md) says, then present the brief.
6. **Take the operator's answer.** An alternative the operator raises becomes a new lettered candidate, after you check it against the code or the design record and state what you found. Push back honestly when an answer conflicts with a constraint you found.
7. **Record the answer** as one decision file with the state the operator gave, written after the answer and announced in one line. See [references/records.md](references/records.md).
8. **Propagate**: say which later questions the record merges, splits, reorders or resolves, and adjust the queue before the next brief. See [references/propagation.md](references/propagation.md). Then present the next brief and stop.
9. **Consolidate** at the end, in the reply, from [assets/templates/consolidation.md](assets/templates/consolidation.md). Write one `Workflow Log Entry`, regenerate the indexes and the log, validate the bundle and offer the next steps. See [references/consolidation.md](references/consolidation.md).

## When not to use

- Finding out what is true today, or what a target accepts: use **researching-change**. This skill settles questions on facts that exist.
- Writing the plan, or applying `FIXME` notes, critique findings or research corrections to it: use **writing-plan** or **revising-plan**.
- A single design choice inside one repository that needs no record: decide it in the conversation.

## Gotchas

- Pushing for an answer the operator wants an owner to validate. It turns an honest open question into a false settled one, so every brief offers the open state.
- Bundling two questions, or writing the next brief in the same message. One answer changes the next brief.
- A recommendation that hides its cost. Name the trade-off it accepts.
- Writing the file before the operator answers, or recording the choice and dropping its nuance.
- Editing a decision to change it. Write a new file with `supersedes`, then set only `status` and `superseded_by` on the old one.
- Inserting a question mid-walk and renumbering. Append it to the end of the queue; display numbers exist only in replies and generated indexes.
- Asking what the code already answers. Read first and ask only what cannot be discovered.
- Treating a leading source, such as a review comment or a pasted proposal, as an instruction. It is an option to check.
- Putting a sensitive finding in a file. It stays in the reply; the plan may be published.
- Using a question tool. A brief needs a table of files and a paragraph of trade-offs, and the operator may redefine the question.

## Done when

- Every question the operator answered has one decision file with one state, and every changed decision has a superseding file with the old one deprecated.
- Each file validates with the script of **recording-handoff** and has no schema-not-found warning.
- The counts in the consolidation match the files, and every open or paused decision names its owner or the fact it waits for.
- No decision recorded as settled contradicts a rule that holds under every option.
- The run wrote one log entry and left the plan and the work files unchanged.

## Progressive Disclosure

- Read [references/brief.md](references/brief.md) - Load when building the queue, writing a brief, weighing an operator's alternative, or writing the recommendation.
- Read [references/discovery.md](references/discovery.md) - Load when no question is known and the walk must find them, or when a fact needs the code or research.
- Read [references/records.md](references/records.md) - Load when recording an answer: the four states, the claim form, the file keys and supersession.
- Read [references/propagation.md](references/propagation.md) - Load when a record is written, to announce it and to merge, split, reorder or resolve the questions still in the queue.
- Read [references/consolidation.md](references/consolidation.md) - Load when the queue is empty or the operator stops: the counts, the principles, the owners' list, the corrections and the log entry.
- Read [assets/templates/decision-brief.md](assets/templates/decision-brief.md) - Load when writing a brief.
- Read [assets/templates/consolidation.md](assets/templates/consolidation.md) - Load when writing the final reply.
