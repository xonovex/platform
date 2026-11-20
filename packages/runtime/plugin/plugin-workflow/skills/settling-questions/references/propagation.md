# Propagation

After each record, say what the answer does to the questions still in the queue and adjust the queue before the next brief. Skipping this produces briefs that contradict earlier answers.

## The four effects

| Effect | Meaning | What changes in the queue |
| --- | --- | --- |
| Merge | Two questions are now one, because the answer ties them together. | One brief replaces two. Its decision file lists both research question ids. |
| Split | The answer shows a question has two independent parts. | Two entries replace one; the second goes to the end of the queue. |
| Reorder | The answer makes another question a prerequisite, or removes the need to ask it first. | The queue changes order; give the reason in one line. |
| Resolve | The answer settles a later question as a consequence. | The later question leaves the queue after the operator confirms the consequence. |

A question the answer makes moot is a resolve, not a silent deletion. A question that becomes unanswerable until a fact is known is paused, with the fact named.

## What to say

After the operator's answer and after the file is written, announce the write in one line: the path and the state, such as "Wrote decisions/target-identifier-7k2f.md as settled." Never write before the answer. Then list only the effects that exist, one line each, and name the question by its title:

- "Resolves the retention question: with device ids only, there is no account record to expire."
- "Reorders: the rollout question now comes before the phasing question, because it decides who owns the flag."
- "Nothing else in the queue changes."

Then present the next brief. If the queue is empty, consolidate.

## Effects on files

- A resolved question still gets its decision file, because a later reader cannot tell a resolved question from a forgotten one. Ask the operator to confirm the consequence in one line, then write a `settled` decision whose `depends_on` names the decision it follows from. The confirmation makes it the operator's, so it carries `verified`.
- A merged question gets one file that cites every research question id it covers in its description.
- A split question gets a file for each part when it is recorded; nothing is written for the part still queued.
- A reorder changes the queue only. `depends_on` on a file records a real prerequisite, set when the file is written, because a decision file is not edited afterwards. `blocks` may list existing decisions that wait on the new one.
- An answer that contradicts an earlier record reopens it. Say so, confirm with the operator, and write the new decision with `supersedes`.
- A new question the answer raises goes to the end of the queue and, if it bears on the plan, into the consolidation.

## Work that waits

The skill does not read or edit work files to change their readiness; **revising-plan** does. It does list, in the consolidation, the work files whose `waits_on` names a decision that this run settled or superseded, so the operator can see what can start now and what the next revision has to update. When no work files exist yet, say so.
