---
description: >-
  Create a high-level plan with research for user review before detailed
  subplans
allowed-tools:
  - Write
  - Read
  - Glob
  - Grep
  - TaskCreate
  - TaskUpdate
  - AskUserQuestion
argument-hint: "[spec-file-or-requirements] [--interactive] [--depends-on <plan>] [--dry-run]"
---

# /plan-create – Create Plan with Research

Generate high-level plan from spec/requirements using deep codebase and web research. Presents architecture decisions for user review BEFORE generating detailed subplans.

## Prerequisites

Run research before using this command:

- `/plan-research` - General codebase and web research
- `/code-simplify` - Code simplification analysis
- `/code-harden` - Code hardening analysis
- Or other `/plan-research-*` commands as appropriate

This command assumes research context is already in the conversation. It does NOT perform codebase exploration.

## Goal

- Generate plan with architecture decisions and technology choices from conversation context
- Present plan for user review/discussion
- Save plan and STOP (user reviews, then runs /plan-subplans-create)

## Arguments

- `spec-file-or-requirements` (optional): Path to spec or inline requirements (defaults to conversation context)
- `--interactive` (optional): Ask context-dependent technical questions during research
- `--depends-on <plan>` (optional): Mark dependency on another plan
- `--dry-run` (optional): Preview without writing files

## Core Workflow

**IMPORTANT: Do NOT use EnterPlanMode. Do NOT use Task/Explore agents - assume research is already done.**

1. **Gather requirements**: Read spec file or analyze conversation context; ask clarifications if `--interactive`
2. **Synthesize context**: Use research findings from conversation context
3. **Document key decisions**: Technology choices with versions, rationale, alternatives
4. **Assess risks**: Trade-offs, alternatives considered, open questions
5. **Propose subplan structure**: List subplan names without detailed implementation
6. **Write plan**: Save to `plans/<feature-name>.md` (standalone file, NOT inside existing plan directories unless explicitly requested)
7. **Show summary**: Display plan for user review; STOP (no subplans, no implementation)

## Implementation Details

**Interactive Mode**: Ask about architectural preferences, library choices, error handling, testing depth, accessibility

**Plan Frontmatter**: `type: plan`, `has_subplans: true`, `status: pending-approval`, `dependencies.plans: []`, `proposed_subplans: []`, `skills_to_consult: [skill-names]`, `research_sources: {documentation: [], versions: {}}`

**Skills to Consult**: Plan MUST include `skills_to_consult` array listing applicable coding guidelines based on languages/frameworks (e.g., `typescript-guidelines`, `react-guidelines`, `testing-guidelines`). This ensures implementers know which project conventions to follow.

**Plan Sections**: Overview (2-3 sentences), Goals (bullet list), Current State (stack/integration), Research Findings (recommended library with version/rationale/pros-cons/docs, alternatives), Proposed Approach (numbered list of components/files), Risk Assessment, Proposed Child Plans (with execution groups), Success Criteria, Estimated Effort

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

Next Steps:
1. Review the plan for architecture decisions and technology choices
2. Discuss and evaluate research findings, alternatives, and trade-offs
3. Verify proposed child plan structure and execution groups
4. Generate child plans: /plan-subplans-create plans/feature-name.md
5. Alternative: Modify plan if different approach needed, then re-run child plan creation
```

## Examples

```bash
# Create plan with research
/plan-create specs/feature.md

# Interactive with questions
/plan-create --interactive

# With dependency
/plan-create specs/profile.md --depends-on plans/auth.md
```

## Error Handling

- Error: spec doesn't exist, plan exists, output dir fails
- Warning: no requirements, dependency missing, library research failed
