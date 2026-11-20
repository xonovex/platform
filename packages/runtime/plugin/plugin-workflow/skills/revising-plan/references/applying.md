# Applying the dispositions

## Order of work

1. Choose the name of this run's log entry first (`log/<date>-revising-plan-<suffix>`), because a deprecated work file and the table refer to it.
2. New decision files, then the deprecation of the files they supersede.
3. Work files, then `plan.md` outside its generated block, as [work-and-plan.md](work-and-plan.md) says.
4. The log entry with the dispositions table.
5. Removal of the resolved `FIXME` notes.
6. Propagation and regeneration, in [propagation.md](propagation.md).

Announce each write in one line.

## Items to decisions

| The item | Change |
| --- | --- |
| The operator's confirmed answer settles an open, paused or partly settled decision | A new decision file with state `settled` that supersedes it. |
| The answer settles only part of the question | A new `partly-settled` file; the remaining part is sent to **settling-questions**. |
| Evidence removes the basis of a settled decision | A new decision file with state `open`, or `paused` when it names the fact it waits for, that supersedes it. The question is sent to **settling-questions**. |
| Evidence bears on an open or paused decision | No decision file. The evidence goes to **settling-questions** with the question. |
| A new question | No decision file. The question goes to **settling-questions**, which owns new decisions. |
| A change of wording, scope, a criterion or a work file | No decision file. Apply it to the plan or a work file, else not applied with a reason. |

## Decision files

Follow the shared decision template and write rules of **recording-handoff**.

- The new file gets `supersedes` with the old concept ID. `depends_on` and `blocks` carry over, pointing at the current file of each decision.
- `generated.by` is `revising-plan/<version>`. A decision the operator settled or confirmed in this session gets a `verified` entry for the operator, with the id from the invoker. A reopened decision that nobody confirmed has no `verified` entry; say in the reply that the owner may confirm it.
- `# Decision` is one testable sentence in the form "When ..., the system shall ...", or for a reopened decision the question to settle. `# Reason` cites the note, the critique finding or the research file by its source. `# Options considered` keeps the options of the old file and what the item said about them.
- Another decision file that names the old one in `depends_on` or `blocks` is not edited. Readers follow `superseded_by`.
