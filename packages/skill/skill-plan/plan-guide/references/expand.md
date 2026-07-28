# Expand: A Plan into Child Plans

Expand any explicit parent plan, inline or provider-native, into focused child plans. An approval or status field is not a precondition. Stop after returning the child plans and execution ordering.

## Core Workflow

1. Resolve the explicit parent plan, optional native revision, active canonical
   context, supporting references, and selected source-provider capabilities.
2. Extract the objective, components, constraints, risks, dependencies, success
   criteria, applicable context, and `skills_to_consult` capabilities.
3. Create focused child plans with roughly five to seven tasks, affected resources or
   packages, applicable decisions, validation, success criteria, dependencies,
   and any descriptive status metadata the provider needs.
4. Derive execution groups from explicit dependencies and overlapping mutable files, packages, environments, or provider resources. Non-overlapping children may run concurrently; conflicting children are sequenced.
5. Return inline child plans and ordering, including the parent, child, and dependency relationships a later Publish operation should preserve.
6. Stop before implementation.

## Gotchas

- Do not reject an explicit parent because it lacks approval metadata or uses an unfamiliar status.
- A parent status may describe progress but never grants permission for expansion.
- Keep children independently completable; overlapping ownership without sequencing creates merge risk.
- Source-provider relationships are native resources, not a universal wrapper or required file layout.
