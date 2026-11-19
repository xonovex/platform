# Discovering questions and finding facts

## Discovering questions

Use it when nothing is queued, or the operator brings only an idea.

1. Frame the subject in one sentence and read any file the operator named.
2. Start at the top-level intent and branch into sub-questions. Resolve upstream branches first; an answer prunes or reshapes the branches under it.
3. Ask one question at a time, with your recommended answer and a one-line reason, then stop. The brief may be short: the question, the options and the recommendation.
4. Answer yourself what the repository answers: existing patterns, types, conventions, written decisions. Ask only what cannot be found.
5. Continue until no branch is unaddressed and no assumption is unstated.

Each answer is recorded like any other. The walk ends when the tree is resolved or the operator stops.

## When a fact is needed

Go back to the code, or to research, in four cases. Outside them stay in the walk.

| Trigger | What to do |
| --- | --- |
| The operator asks a factual question ("did the web client ever retry this call?") | Run the commands, then answer with what they showed and what they did not show, before any further brief. |
| The operator asks for more checking ("make sure both clients behave the same way") | Start a read-only agent in the background with a stated question and reason, continue with the next brief, fold it in later. |
| The operator reframes the question | Check the facts first, then present a new brief for the reframed question. |
| A brief would list a fact as "to verify while planning" | Leave it as an open item in the record. Do not resolve it by assertion. |

A finding that bears on an earlier record reopens that decision: say so, and offer to record the change as a new decision that supersedes it. When the fact needs a wider read than a few commands, pause the question, name the fact, and offer **researching-change** as the next step. Never start it.

A read-only agent reads and reports; it writes no file and changes no state. Run only a bounded number at once, with the lowest model and effort that does the part reliably, and re-check the claim that would change the answer yourself.
