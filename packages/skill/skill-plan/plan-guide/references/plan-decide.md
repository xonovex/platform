# plan-decide: Settle Decisions One at a Time

Settle the decisions a plan needs — one per message, each with an honest recommendation — in whichever mode the input state calls for. **Walk** a known, bounded decision set (left by research, a draft plan, or critique findings) as full prose briefs. **Discover** unknown decisions by interviewing the user down the design tree when nothing is queued yet. Ends with a consolidated agreed-direction summary that feeds `plan-create` or `plan-revise`. Does **not** create or edit plans.

## Stage

Pre-plan (Discover) through draft (Walk). Mode selection: if open decisions are already enumerated — research findings in the conversation, a plan's open questions, critique findings — Walk them; if only an idea or direction exists, Discover. When the user has concrete feedback to _apply_ to a plan document, that is `plan-revise`, not this operation.

## Mode: Walk (known decisions — convergent)

**Stay in decision mode — no plan authoring, no implementation, no AskUserQuestion.**

1. **Extract decisions** from research findings or a plan: ambiguities, contradictions between sources, leading assumptions worth challenging, unsettled trade-offs
2. **Order by dependency** — upstream decisions first; a scoping answer reshapes placement decisions downstream
3. **Present one brief**, then STOP and wait for the answer
4. **Record the answer with its nuances** ("yes, but don't use that word in the app layer") as a single testable claim (EARS: "When \<trigger\>, the system shall \<response\>") so it is unambiguous and checkable, and push back honestly if it conflicts with a discovered constraint
5. **Propagate** the answer into the remaining queue — merge, split, resolve, reframe — before presenting the next
6. **Consolidate** at the end: numbered decisions, governing principles the answers revealed, sequencing, next command

### Decision Brief Format

```
## Decision N of M: <title>
### Files involved      — table with paths, consumers, markers
### Explanation         — the constraint that makes it non-obvious;
                          what the source assumes and whether that holds
### Options             — 2-4, each with pros and cons; the literal reading
                          of a leading source is itself one option
### Recommendation      — one honest pick, with the accepted trade-off named
```

Then STOP — no next-decision content in the same message.

### Why Plain Text, Not a Question Tool

Architectural decisions need file tables, consumer lists, and multi-paragraph trade-offs that do not fit option chips. Users answer partially, attach constraints, redefine the decision, or challenge the framing — prose preserves that; forced options destroy it.

## Mode: Discover (unknown decisions — divergent)

**Stay in decision mode — no plan authoring, no implementation.**

1. **Frame the subject** — the feature / idea / direction under question; read any named file or infer it from the conversation
2. **Walk the design tree** — start at the top-level intent, branch into sub-decisions, resolve dependencies depth-first (an upstream answer prunes or reshapes downstream branches)
3. **One question at a time** — ask a single question, give your **recommended answer** with a one-line rationale, then STOP and wait
4. **Self-answer from the codebase** — if a question is answerable by reading the repo (existing patterns, types, conventions), explore it instead of asking
5. **Record each answer with its nuance** and push back when it conflicts with a discovered constraint
6. **Continue until the tree is resolved** — no unaddressed branches, no unstated assumptions
7. **Consolidate** — a shared-understanding summary: decisions made, open risks, the governing intent, and the next command

### Question Format

```
<single question about one branch>
Recommended: <your pick> — <one-line why>
```

Then STOP — one question per message; the answer reshapes the next.

## Example Output

```
Decided: 6 decisions (Walk)

1. Migration scope   → mechanism/policy split
2. Contract files    → migrate semantics first, co-locate later subplan
3. App layer         → lifecycle stays, composition moves; no "level" vocab in app/
...
Sequencing: 5 → 6 → (1,4) → (2,3)
Next: plan-create
```

## Gotchas

- Reaching for a question tool in Walk mode is the #1 mistake — full-context prose briefs are the point
- Bundling decisions or batching questions defeats the format — one answer changes the next brief/question
- Options without a recommendation outsource the analysis to the user — always recommend
- Recording the choice but dropping its nuance silently loses a plan constraint
- Skipping propagation yields briefs that contradict earlier answers
- Asking what the codebase already answers wastes the user's time — grep/read first, ask only what isn't discoverable
- A leading source (review annotations, comments) is input to challenge, not instructions to execute
- Drifting into authoring a plan — decide settles decisions; `plan-create` writes them up
