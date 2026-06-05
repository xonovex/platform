# plan-clarify: Clarify Open Decisions One by One

Walk the user through every open decision left by research or a draft plan — one decision per message, in plain prose, never via a question tool. Each decision is a full brief (files involved, explanation, options with pros/cons, honest recommendation); STOP after each and wait for the user's call. Ends with a consolidated agreed-direction summary that feeds `plan-create` or `plan-refine`. Does **not** create or edit plans.

## Core Workflow

**Stay in clarification mode — no plan authoring, no implementation, no AskUserQuestion.**

1. **Extract decisions** from research findings or a plan: ambiguities, contradictions between sources, leading assumptions worth challenging, unsettled trade-offs
2. **Order by dependency** — upstream decisions first; a scoping answer reshapes placement decisions downstream
3. **Present one brief**, then STOP and wait for the answer
4. **Record the answer with its nuances** ("yes, but don't use that word in the app layer") and push back honestly if it conflicts with a discovered constraint
5. **Propagate** the answer into the remaining queue — merge, split, resolve, reframe — before presenting the next
6. **Consolidate** at the end: numbered decisions, governing principles the answers revealed, sequencing, next command

## Decision Brief Format

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

## Why Plain Text, Not a Question Tool

Architectural decisions need file tables, consumer lists, and multi-paragraph trade-offs that do not fit option chips. Users answer partially, attach constraints, redefine the decision, or challenge the framing — prose preserves that; forced options destroy it.

## Example Output

```
Clarified: 6 decisions

1. Migration scope   → mechanism/policy split
2. Contract files    → migrate semantics first, co-locate later subplan
3. App layer         → lifecycle stays, composition moves; no "level" vocab in app/
...
Sequencing: 5 → 6 → (1,4) → (2,3)
Next: plan-create
```

## Gotchas

- Reaching for a question tool is the #1 mistake — full-context prose briefs are the point
- Bundling decisions defeats the format; one answer changes the next brief
- Options without a recommendation outsource the analysis to the user
- Recording the choice but dropping its nuance silently loses a plan constraint
- Skipping propagation yields briefs that contradict earlier answers
- A leading source (review annotations, comments) is input to challenge, not instructions to execute
