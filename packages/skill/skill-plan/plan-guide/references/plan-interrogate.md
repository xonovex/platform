# plan-interrogate: Surface Unknown Decisions by Interrogation

Relentlessly interview the user about a feature idea or design direction to surface the decisions, edge cases, and hidden assumptions nobody has written down yet. One question at a time, each with a recommended answer; explore the codebase to self-answer rather than asking what the code already states. Ends with a consolidated shared-understanding summary that feeds `plan-clarify` or `plan-create`. Does **not** author or edit plans.

## Interrogate vs Clarify

- **plan-interrogate is divergent** — it _discovers_ unknown decisions: generates questions on the fly, walks the design tree depth-first, explores the codebase, and keeps going until the tree is resolved (often many questions).
- **plan-clarify is convergent** — it _resolves_ a known, bounded decision set in plain prose, with no question tools.
- Run interrogate upstream (an idea exists but the requirements don't); run clarify downstream (the decisions are enumerated and each needs settling).

## Core Workflow

**Stay in interrogation mode — no plan authoring, no implementation.**

1. **Frame the subject** — the feature / idea / direction under question; read any named file or infer it from the conversation
2. **Walk the design tree** — start at the top-level intent, branch into sub-decisions, resolve dependencies depth-first (an upstream answer prunes or reshapes downstream branches)
3. **One question at a time** — ask a single question, give your **recommended answer** with a one-line rationale, then STOP and wait
4. **Self-answer from the codebase** — if a question is answerable by reading the repo (existing patterns, types, conventions), explore it instead of asking
5. **Record each answer with its nuance** and push back when it conflicts with a discovered constraint
6. **Continue until the tree is resolved** — no unaddressed branches, no unstated assumptions
7. **Consolidate** — a shared-understanding summary: decisions made, open risks, the governing intent, and the next command

## Question Format

```
<single question about one branch>
Recommended: <your pick> — <one-line why>
```

Then STOP — one question per message; the answer reshapes the next.

## Why One Question at a Time

A batch forces the user to context-switch and buries the dependency between answers. One question, answered, changes which question comes next — and the recommended answer lets the user reply "yes" instead of writing an essay.

## Example Output

```
Shared understanding: 14 questions resolved

Intent: real-time notifications, in-app + email
Decisions:
- Transport   → SSE (matches the existing stream in api/)
- Offline     → persist 7 days, replay on reconnect
- Fan-out     → per-user queue, capped at 1k
Open risks: load-test fan-out before committing
Next: plan-clarify (settle the 3 flagged trade-offs) → plan-create
```

## Gotchas

- Asking what the codebase already answers wastes the user's time — grep/read first, ask only what isn't discoverable
- Dropping the recommended answer turns a fast "yes" into an essay — always recommend
- Batching questions defeats the dependency walk — one at a time, and propagate each answer before the next
- Recording a decision but losing its nuance silently drops a future plan constraint
- Drifting into authoring a plan — interrogate surfaces decisions; `plan-create` writes them up
