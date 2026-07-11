# plan-create: Create High-Level Plan with Research

Generate a high-level parent plan from research already in the conversation, presenting architecture decisions for review BEFORE any detailed subplans. Save the plan and STOP (user reviews, then runs `plan-subplans-create`).

## Prerequisites

Run `plan-research` first (incl. its code-quality audits). This command assumes research is already in the conversation and does **NOT** explore the codebase or delegate to search agents.

## Core Workflow

**Do NOT switch into a plan-authoring mode. Do NOT implement.**

1. **Gather requirements** — read spec or conversation; ask clarifications only if interactive mode was requested
2. **Document key decisions** — technology choices with versions, rationale, alternatives
3. **Assess risks** — trade-offs, alternatives considered, open questions
4. **Propose subplan structure** — subplan names only, no implementation detail
5. **Write plan** — save to `plans/<feature-name>.md` (standalone file, not inside an existing plan dir unless requested); show summary; STOP

## Test-first plans

When the request is test-first, apply **tdd-guide**'s red-green-refactor (or **bdd-guide** for acceptance-first): structure each step as failing-test → implement → refactor, keep one assertion failing at a time, and list the test doubles per **testing-guide**. Same document shape — the steps just lead with the test.

## Plan shape

- **Frontmatter** — `type: plan`, `has_subplans: true`, `status: pending-approval`, `dependencies.plans: []`, `proposed_subplans: []`, `skills_to_consult: [skill-names]`, `research_sources: {documentation: [], versions: {}}`
- **`skills_to_consult`** — MUST list applicable coding guidelines (e.g. `typescript-guide`, `react-guide`, `testing-guide`); never empty
- **Sections** — Overview (2-3 sentences), Goals, Current State (stack/integration), Research Findings (recommended library with version/rationale/pros-cons/docs, alternatives), Proposed Approach (numbered components/files), Risk Assessment, Proposed Child Plans (with execution groups), Success Criteria, Estimated Effort

## Gotchas

- Skipping research and going straight to plan-create produces vague plans — the prerequisite isn't optional
- `skills_to_consult` empty means implementers won't read project conventions — at least list language guides
- Mixing detailed implementation steps into the parent plan defeats the create → review → subplans flow
- A plan proposing >10 subplans is too coarse — split into multiple parent plans
