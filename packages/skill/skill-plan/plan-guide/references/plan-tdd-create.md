# plan-tdd-create: Create TDD Plan with Research

Generate a high-level TDD plan for user review BEFORE detailed RED-GREEN-COMMIT generation. For complex features with multiple stories.

## Prerequisites

Run research first (see `plan-research.md`, including its code-quality audits — harden / simplify / align); this command assumes research context is already in the conversation and does NOT perform codebase exploration.

## Goal

- Break requirements into stories with atomic test-step proposals from conversation context
- Present test strategy for review
- Save plan and STOP (user approves; subplans generated separately)

## Core Workflow

**IMPORTANT: Do NOT switch into a plan-authoring mode. Do NOT delegate to codebase-exploration agents — assume research is already done.**

1. **Gather context** — read spec; use research findings from conversation
2. **Decompose** — identify stories, propose atomic test steps (one test = one commit), note file impacts, identify dependencies
3. **Document** — write plan with research findings, proposed steps table, reference files, open questions
4. **Save** — write to `plans/<feature>-tdd.md` (standalone file); STOP for user review

## Plan Structure

```markdown
# [Feature] TDD Plan

**Frontmatter:**
type: plan
has_subplans: true
status: pending-approval

## Overview

[Description and approach]

## Skills to Consult

- [Language / framework-specific guidelines]

## Research Findings

- **Test Framework:** [Project's test framework]
- **Test Organization:** [conventions]
- **Validation Pattern:** [Project's validation approach]
- **Architecture:** [Project's architecture style]
- **Reference:** `path/to/reference-test-file`

## Stories & Proposed Steps

### Story 1: [Name]

| Step | Test Description | Files   |
| ---- | ---------------- | ------- |
| 1.1  | [Test verifies]  | [files] |

### Story 2: [Name]

…

## Execution Strategy

- Dependencies: [order, parallel opportunities]
- Estimated: Story 1 (N steps), Story 2 (M steps), Total (X commits)

## Risk Assessment

- Complex areas, edge cases, validation needs
- **Open Questions:** [clarifications needed]
```

## Step Patterns

- **New components:** basic creation test → property tests → behavior tests → edge cases → validation tests
- **Extending components:** new property/behavior test → integration → edge cases
- **Granularity check:** independent commit? tests ONE thing? clear description?

## Gotchas

- Story-level granularity ("Auth flow") is too coarse for TDD — break into atomic test steps that each map to one commit
- Skipping the test-step table in favor of prose makes execution counting impossible — keep the table
- A test step that tests two unrelated invariants belongs in two steps — RED-GREEN-COMMIT needs one assertion to fail at a time
