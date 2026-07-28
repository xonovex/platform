# Revise: A Plan from Explicit Feedback

Create one revised plan from an inline plan or provider-native reference, explicit feedback, and an optional native revision. Preserve the source and return the new result inline.

## Feedback Inputs

Accept inline feedback, provider-native feedback references, or provider-native annotations. For Markdown plans, annotations may be HTML comments, callouts, inline notes, strike/replace markers, insertion/deletion markers, or `TODO`/`FIXME`/`QUESTION` lines.

## Core Workflow

1. Resolve the explicit plan, its carried decisions, feedback, supporting
   references, and optional native revision through selected providers when
   applicable.
2. Enumerate every distinct feedback item before editing. Stop on materially conflicting instructions that the available evidence cannot resolve.
3. Classify each item as applied, deferred, rejected, unresolved question, or scope change, with a concise rationale.
4. Propagate accepted changes through scope, approach, dependencies, risks, proposed children, validation, success criteria, and skills to consult.
5. Produce a traceable new revision without invisibly overwriting the source. Preserve
   still-active context and create versioned superseding or invalidating context
   records when accepted feedback changes it. Status remains descriptive metadata.
6. Return the revised plan and item-by-item disposition inline. Use a separate Publish operation if the revision must be persisted.

## Gotchas

- Do not implement the plan while revising it.
- Do not silently discard feedback or stale source evidence.
- A previous status value has no authorization effect on the revised result.
