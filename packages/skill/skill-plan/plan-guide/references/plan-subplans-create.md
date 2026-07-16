# plan-subplans-create: Publish Detailed Child Planning Results

Expand an approved parent Planning revision into detailed child Planning results, preserve provider-native relationships, and stop before implementation.

## Core workflow

1. Resolve the parent provider context, opaque native reference, exact approved revision, source references, profile, and provider capabilities.
2. Extract objective, components, decisions, risks, dependencies, success criteria, and required `skills_to_consult` capabilities.
3. Create child Planning results with focused objective, roughly 5-7 tasks, affected resources/packages, validation, success criteria, dependencies, and status pending.
4. Detect execution groups using explicit dependencies plus overlapping mutable resources, packages, environments, or provider objects. Non-overlapping children may run concurrently; conflicting ones are sequenced.
5. Publish each child through the selected provider and create native parent/child and dependency relationships. If relations are unsupported, use provider-native link semantics and state the limitation.
6. Publish a new parent revision that references the children and execution groups, then return their opaque references and stop.

Apply [early-lifecycle-contracts.md](early-lifecycle-contracts.md). A local Markdown provider may create a plan directory and child files; file paths, line numbers, Git configuration, and worktrees are optional provider/workspace behavior, not canonical Planning semantics.
