# plan-create: Create High-Level Plan with Research

Generate a high-level plan from spec / requirements using research already in conversation context. Presents architecture decisions for review BEFORE generating detailed subplans.

## Prerequisites

Run research first (see `plan-research.md`, including its code-quality audits — harden / simplify / align); this command assumes research context is already in the conversation and does NOT perform codebase exploration.

## Goal

- Generate plan with architecture decisions and technology choices from conversation context
- Present plan for user review / discussion
- Save plan and STOP (user reviews, then runs `plan-subplans-create`)

## Core Workflow

**IMPORTANT: Do NOT switch into a plan-authoring mode. Do NOT delegate to codebase-exploration agents — assume research is already done.**

1. **Gather requirements** — read spec or analyze conversation; ask clarifications if interactive mode was requested
2. **Synthesize context** — use research findings from conversation
3. **Document key decisions** — technology choices with versions, rationale, alternatives
4. **Assess risks** — trade-offs, alternatives considered, open questions
5. **Propose subplan structure** — list subplan names without detailed implementation
6. **Write plan** — save to `plans/<feature-name>.md` (standalone file, not inside an existing plan dir unless requested)
7. **Show summary** — display plan for user review; STOP

## Test-first plans

When the request is test-first, apply **tdd-guide**'s red-green-refactor (or **bdd-guide** for acceptance-first): structure each step as failing-test → implement → refactor, keep one assertion failing at a time, and list the test doubles per **testing-guide**. The plan document is the same shape — the steps just lead with the test.

## Implementation Details

### Interactive mode

ask about architectural preferences, library choices, error handling, testing depth, accessibility

### Plan frontmatter

`type: plan`, `has_subplans: true`, `status: pending-approval`, `dependencies.plans: []`, `proposed_subplans: []`, `skills_to_consult: [skill-names]`, `research_sources: {documentation: [], versions: {}}`

### Skills to consult

plan MUST include `skills_to_consult` listing applicable coding guidelines (e.g. `typescript-guide`, `react-guide`, `testing-guide`) so implementers know project conventions

### Plan sections

Overview (2-3 sentences), Goals (bullets), Current State (stack / integration), Research Findings (recommended library with version / rationale / pros-cons / docs, alternatives), Proposed Approach (numbered components / files), Risk Assessment, Proposed Child Plans (with execution groups), Success Criteria, Estimated Effort

## Output

```
Created plan: plans/feature-name.md

Research Summary:
- Analyzed codebase: 15 files in services/packages/example
- Current stack: React 19, Astro 5.15, Tailwind 4.1
- Recommended: canvas-confetti@1.9.3 (latest stable)
- Documentation: https://example.com/docs

Plan Summary:
- Goals: Add celebration animation on form submission
- Approach: 3 components (library, modal, integration)
- Proposed Child Plans: 4 (2 parallel, 2 sequential)
- Estimated Effort: 2-3 hours
```

## Error Handling

- Error: spec doesn't exist, plan already exists, output dir fails
- Warning: no requirements, dependency missing, library research failed

## Gotchas

- Skipping research and going straight to plan-create produces vague plans — the prerequisite isn't optional
- "Skills to consult" empty means implementers won't read project conventions — at least list language guides
- Mixing detailed implementation steps into the parent plan defeats the create → review → subplans flow
- A plan that proposes >10 subplans is too coarse — split into multiple parent plans
