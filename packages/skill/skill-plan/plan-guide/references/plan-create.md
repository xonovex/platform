# plan-create: Publish a High-Level Planning Result

Create one high-level parent Planning result from resolved lifecycle inputs, present it for review, and stop before detailed child plans or implementation.

## Preconditions

- Resolve material Research, Decision, Formulation, Experience Design, and Solution Design inputs through their provider contexts and opaque native references.
- Assess native revisions and freshness. If material research or decisions are missing, route to `research-run` or `decision-create` instead of inventing them.
- Apply the method, executor, policy, authority, and publication rules in [early-lifecycle-contracts.md](early-lifecycle-contracts.md).

## Core workflow

1. Resolve the selected workflow profile and independent result-provider, method, workspace, and policy axes.
2. Gather objective, scope, exclusions, source references, accepted decisions, dependencies, constraints, risks, validation, success criteria, and unresolved gaps.
3. Propose parent-level components and child Planning result names only; keep detailed implementation tasks for `plan-subplans-create`.
4. List non-empty `skills_to_consult` capabilities for implementers. If a selected method is test-first or acceptance-first, load a matching installed TDD/BDD/testing capability; do not require one when the method is not selected.
5. Publish a canonical Planning result through the selected provider with status `pending-approval`, provider-native revision, source relationships, profile/policy versions, and follow-up capabilities.
6. Present the provider context, opaque native reference, revision, risks, and unresolved gaps; stop for review.

## Planning semantics

The result contains objective, scope, tasks/components, dependencies, risks, validation, success criteria, status, and skills/capabilities to consult. A provider may represent these as a work item, database record, repository document, issue hierarchy, or another native resource. A local Markdown provider may use plan frontmatter and `plans/`, but that representation is not the workflow contract and Git is not required.

## Gotchas

- A conversation summary is not a Research or Decision reference after context loss; resolve native state.
- An explicitly selected unavailable hosted/database/work-item provider fails instead of falling back to a local file.
- Publishing detailed child plans in this operation bypasses parent review.
