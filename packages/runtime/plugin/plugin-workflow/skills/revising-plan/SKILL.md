---
name: revising-plan
description: "Use when applying explicit feedback, annotations, or critique findings to an existing plan as a traceable new revision. Triggers on review comments or inline annotations in a plan document, a disposition for each feedback item (applied, deferred, rejected, question, scope change), propagating an accepted change through scope, risks, and criteria, and conflicting feedback on a plan, even when the user doesn't say 'revise' and only says 'I left notes in the plan'."
---

# Revise a Plan from Explicit Feedback

Create one revised plan from an inline plan or provider-native reference, explicit feedback, and an optional native revision. Preserve the source and return the new result inline.

## Inputs

- **Plan** (optional): the plan document, inline or as a provider-native reference with an optional native revision; detected from the git configuration or the most recent plan in `plans/` when absent
- **Feedback** (optional): inline feedback or a reference to it; the annotations in the plan are used when absent

## Feedback Inputs

Accept inline feedback, provider-native feedback references, or provider-native annotations. For Markdown plans, annotations may be HTML comments, callouts, inline notes, strike/replace markers, insertion/deletion markers, or `TODO`/`FIXME`/`QUESTION` lines.

## Core Workflow

1. Resolve the explicit plan, its carried decisions, feedback, supporting references, and optional native revision through selected providers when applicable.
2. Enumerate every distinct feedback item before editing. Stop on materially conflicting instructions that the available evidence cannot resolve.
3. Classify each item as applied, deferred, rejected, unresolved question, or scope change, with a concise rationale.
4. Propagate accepted changes through scope, approach, dependencies, risks, proposed children, validation, success criteria, and skills to consult.
5. Produce a traceable new revision without invisibly overwriting the source. Preserve still-active context and create versioned superseding or invalidating context records when accepted feedback changes it. Status remains descriptive metadata.
6. Return the revised plan and item-by-item disposition inline. Use a separate Publish operation if the revision must be persisted.

## Boundaries

Findings to apply come from **critiquing-plan**. A piece of feedback that is an undecided question settles through **settling-questions**. The decision on the revised plan belongs to **approving-plan**.

## Gotchas

- Do not implement the plan while revising it.
- Do not silently discard feedback or stale source evidence.
- A previous status value has no authorization effect on the revised result.

## Progressive Disclosure

- Read [references/example.md](references/example.md) - Load when reporting the disposition of each feedback item
