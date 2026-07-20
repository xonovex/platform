# plan-create: Create a High-Level Plan

Create one high-level plan from explicit inline inputs or provider-native references plus optional revisions. Return the plan inline unless the caller explicitly requests a provider-native destination, and stop before detailed child plans or implementation.

## Core Workflow

1. Resolve the subject and supporting references. Let the selected provider interpret opaque references and optional native revisions; do not parse their shapes locally.
2. Gather the objective, scope, exclusions, evidence, known constraints, unresolved questions, dependencies, risks, validation requirements, success criteria, and Definition of Done.
3. Propose parent-level components and child-plan names only. Leave detailed implementation tasks to plan expansion.
4. Add a non-empty `skills_to_consult` list naming the implementation capabilities required by the affected code and toolchain.
5. Create the plan with its assumptions and evidence links visible. Status may be included as descriptive provider metadata, but it has no gating or authorization meaning.
6. Return the plan inline or ask the selected provider to persist it to the explicitly requested destination and return the native reference and revision it supplies.

## Plan Contents

- Objective, scope, and explicit exclusions
- Evidence and supporting references
- Parent-level components and proposed child plans
- Dependencies, constraints, risks, and unresolved questions
- Validation steps, success criteria, and Definition of Done
- `skills_to_consult`

## Gotchas

- Missing material evidence remains an explicit gap; do not invent it.
- Creating detailed child plans here mixes creation with expansion.
- Do not require an approval field or assign authority semantics to status.
- Never replace an explicitly selected unavailable provider with a local plan file.
