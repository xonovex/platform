# plan-revise: Revise an Exact Planning Result

Apply explicit feedback to one Planning revision, publish a new provider-native revision, and stop. Approval remains a separate authority action.

## Feedback inputs

Accept opaque native feedback/critique references, prompt instructions, or provider-native annotations. For a selected Markdown provider, annotation markers may include HTML comments, callouts, inline notes, strike/replace, insertion/deletion markers, or `TODO`/`FIXME`/`QUESTION` lines.

## Core workflow

1. Resolve the Planning provider context, native reference, exact revision, and all feedback/critique references; apply [early-lifecycle-contracts.md](early-lifecycle-contracts.md).
2. Enumerate every distinct item before mutation. If two authority-bearing instructions conflict, stop for the responsible actor.
3. Resolve each item as correction, deletion, addition, question, rejection, or scope change; preserve unresolved items with owner and rationale.
4. Propagate changed decisions through approach, risks, dependencies, child-plan proposal, validation, and success criteria. Re-evaluate stale source references.
5. Publish a new Planning native revision with supersession and source relationships; leave status pending approval unless an independent authority action already exists for this exact revision.
6. Return the new opaque reference/revision and item-by-item disposition; stop.

## Gotchas

- Never overwrite history invisibly or transfer an approval from an earlier revision.
- Editing a local file is valid only when that file provider was selected; its versioning behavior must still produce a resolvable revision.
- Implementing while revising conflates Planning with Development.
