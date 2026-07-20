# discovery-run: Discover a Problem or Opportunity

## Core workflow

1. Resolve the subject and any native input references using [early-lifecycle-contracts.md](early-lifecycle-contracts.md).
2. Select the profile, provider, neutral-or-explicit method, and least-adaptive executor. Record any applicable data, privacy, or regulated constraints as workflow requirements without selecting runtime controls.
3. Collect observations deterministically. Use a bounded model for classification or synthesis; use an adaptive agent only when the discovery path genuinely branches.
4. Separate observed facts, stakeholder statements, assumptions, hypotheses, conflicts, and unknowns. Identify affected context and stakeholders without turning them into mandatory story roles.
5. Publish a canonical Discovery result with problem or opportunity, affected context, observations, assumptions, unknowns, source references, limitations, and executor origin.
6. Ask whether another discovery iteration is needed. Each iteration publishes a new native revision; stop when remaining unknowns are accepted, delegated to Research, or block progress.

## Boundaries

- Discovery does not silently become Formulation, requirements, user stories, or Gherkin.
- Interviews, journey mapping, user research, and other installed skills are optional methods selected by context.
- A model may suggest questions but cannot fabricate stakeholder statements.
