---
description: >-
  Walk open decisions one by one in plain text — files involved, explanation,
  pros, cons, recommendation — capturing the user's call on each before moving
  to the next
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Task
  - Write
argument-hint: "[input-file] [--save-to <file>]"
---

# /xonovex-workflow:plan-clarify – Clarify Open Decisions One by One

Walks the user through every open decision left by research or a draft plan, one decision per message, in plain prose. Each decision is presented as a full brief — files involved, explanation, options with pros and cons, an honest recommendation — then STOPS for the user's answer before presenting the next. Ends with a consolidated agreed-direction summary that feeds `/xonovex-workflow:plan-create` or `/xonovex-workflow:plan-refine`. Does NOT create or edit plans and does NOT implement anything.

## Prerequisites

- Research findings exist (from `/xonovex-workflow:plan-research` output in the conversation or a saved research file) OR a plan document exists (from `/xonovex-workflow:plan-create`)

## Goal

- Extract every open decision from the input: ambiguities, contradictions, leading assumptions, unsettled trade-offs
- Present decisions one at a time as plain-text briefs; STOP after each and wait for the user's call
- Record each answer verbatim including nuances and constraints the user adds
- Propagate each answer into the remaining decisions before presenting the next
- Close with a consolidated agreed-direction summary

## Why Plain Text, Not AskUserQuestion

**Do NOT use the AskUserQuestion tool in this command.** Architectural decisions need full context — file tables, consumer lists, dependency constraints, multi-paragraph trade-offs — which does not fit option chips. Users answer in free text, often partially ("option A, but don't use that word in the app layer"), redefine the decision, or push back on the framing itself. A question tool forces premature option-shaping; prose preserves the discussion.

## Arguments

- `input-file` (optional): Path to a research file or plan document. Auto-detects: research findings in the current conversation first, then the most recent `plans/*.md`
- `--save-to <file>` (optional): Save the consolidated agreed-direction summary to a file

## Core Workflow

**IMPORTANT: Do NOT switch into plan-authoring mode. Do NOT implement anything. This command only clarifies decisions.**

1. **Locate input**: Research findings in conversation, the given file, or the most recent plan
2. **Extract open decisions**: Scan for:
   - Ambiguities the research flagged ("needs a decision", "open question")
   - Contradictions between sources (a comment vs project docs, two conventions in tension)
   - Leading assumptions baked into the source material that deserve challenge
   - Trade-offs with no obviously correct answer
3. **Order by dependency and impact**: Upstream decisions first — an answer to a scoping decision reshapes or resolves placement decisions downstream
4. **Present one decision** as a brief (format below), then **STOP and wait**
5. **Record the answer**: Capture the choice AND every nuance ("yes, but…", vocabulary constraints, sequencing preferences). If the answer conflicts with a constraint found in research, push back honestly before recording — do not silently accept a choice that breaks something
6. **Propagate**: Re-evaluate remaining decisions against the new answer — merge, split, resolve, or reframe them as needed; mention when an earlier answer settled a later decision
7. **Repeat** from step 4 until all decisions are resolved
8. **Consolidate**: Present the full agreed direction — numbered decisions, governing principles the answers revealed, sequencing — and point to the next command (`plan-create` or `plan-refine`)

## Decision Brief Format

Each decision message contains, in order:

```
## Decision N of M: <title>

### Files involved
<table: file | role/consumers/markers — concrete, with paths>

### Explanation
<what the decision is, which constraint makes it non-obvious,
 what the source material assumes and whether that assumption holds>

### Options
<2-4 options, each with pros and cons — including the literal
 reading of the source material as one option when it is leading>

### Recommendation
<one honest pick with reasoning; name the trade-off being accepted>
```

Then STOP. No "next decision" content in the same message.

## Resolution Rules

- One decision per message — never bundle, never ask "and also…"
- Never proceed without an explicit answer; a question back from the user is part of this decision, not the next one
- The user may redefine the decision mid-flight — follow them, not the original framing
- Honest pushback is required when an answer conflicts with discovered constraints; agreement is not the goal, a sound direction is
- A vague answer ("do the middle one-ish") gets a one-line confirmation of the recorded interpretation, not a re-litigation

## Output

```
Clarified: 6 decisions (research: AGENT comment cleanup)

1. Scope of migration       → mechanism/policy split (data + behavior to script)
2. Contract files           → migrate semantics first, co-locate survivors in later subplan
3. Header placement         → co-locate with owning .c, same phase as 2
4. App-layer knowledge      → lifecycle stays in app, composition moves to script;
                              constraint: no "level" vocabulary in app/
5. Globals                  → all eliminated; world backing → caller-owned state
6. Wrappers                 → perf-counter + zeroed-alloc yes, memset wrapper no

Sequencing: 5 → 6 → (1,4) → (2,3)
Next: /xonovex-workflow:plan-create
```

## Examples

```bash
# Clarify decisions from research findings in the conversation
/xonovex-workflow:plan-clarify

# Clarify decisions from a saved research file
/xonovex-workflow:plan-clarify research/agent-comments.md

# Clarify a draft plan's open decisions and save the agreed direction
/xonovex-workflow:plan-clarify plans/layer-cleanup.md --save-to plans/layer-cleanup-direction.md
```

## Error Handling

- Error if no research findings or plan can be located
- Info if zero open decisions found — the input is already unambiguous; suggest proceeding to `plan-create`
- Warning if a decision needs codebase facts not in the input — run the lookup (read-only) before presenting the brief, never present a brief built on guesses

## Gotchas

- Reaching for AskUserQuestion is the #1 mistake — the whole point of this command is full-context prose briefs
- Bundling decisions to "save rounds" defeats the format; one answer changes the next brief
- Presenting options without a recommendation outsources the analysis to the user — always pick one and say why
- Recording the choice but dropping the nuance ("but don't mention level") silently loses a constraint the plan must honor
- Skipping propagation produces briefs that contradict earlier answers — re-evaluate the queue after every answer
- A leading source (annotations, review comments) is input to challenge, not instructions to execute — say so when its assumption fails
