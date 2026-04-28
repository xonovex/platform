---
description: "Create a high-level test-driven plan with research for user review before detailed step generation. Use when the user asks to plan with TDD, design test-first, or scope a feature test-first. Keywords: TDD, test-driven, plan, test-first, planning, plan document."
---

# /xonovex-workflow:plan-tdd-create – Create TDD Plan with Research

Generate high-level TDD plan for user review BEFORE detailed RED-GREEN-COMMIT generation. For complex features with multiple stories.

## Prerequisites

Run research before using this command:

- `/xonovex-workflow:plan-research` - General codebase and web research
- `/xonovex-workflow:code-simplify` - Code simplification analysis
- `/xonovex-workflow:code-harden` - Code hardening analysis
- Or other `/plan-research-*` commands as appropriate

This command assumes research context is already in the conversation. It does NOT perform codebase exploration.

## Goal

- Break requirements into stories with atomic test step proposals from conversation context
- Present test strategy for review
- Save plan and STOP (user approves, then runs /plan-tdd-subplans-create)

## Core Workflow

**IMPORTANT: Do NOT use EnterPlanMode. Do NOT use Task/Explore agents - assume research is already done.**

1. **Gather context**: Read spec; use research findings from conversation context
2. **Decompose**: Identify stories, propose atomic test steps (one test = one commit), note file impacts, identify dependencies
3. **Document**: Write plan with research findings, proposed steps table, reference files, open questions
4. **Save**: Write to `plans/<feature>-tdd.md` (standalone file, NOT inside existing plan directories unless explicitly requested); STOP for user review

## Plan Structure

```markdown
# [Feature] TDD Plan

**Frontmatter:**
type: plan
has_subplans: true
status: pending-approval

## Overview

[Description and approach]

**Status:** pending-approval

## Skills to Consult

- [Language/framework-specific guidelines]

## Research Findings

- **Test Framework:** [Project's test framework]
- **Test Organization:** [conventions]
- **Validation Pattern:** [Project's validation approach]
- **Architecture:** [Project's architecture style]
- **Reference:** `path/to/reference-test-file`

### Package Structure

**Location:** `package-path`
[Project directory structure]

## Stories & Proposed Steps

### Story 1: [Name]

| Step | Test Description | Files   |
| ---- | ---------------- | ------- |
| 1.1  | [Test verifies]  | [files] |

### Story 2: [Name]

...

## Execution Strategy

- Dependencies: [order, parallel opportunities]
- Estimated: Story 1 (N steps), Story 2 (M steps), Total (X commits)

## Risk Assessment

- Complex areas, edge cases, validation needs
- **Open Questions:** [clarifications needed]

## Next Steps

After approval: `/plan-tdd-subplans-create plans/[feature]-tdd.md`
```

## Step Patterns

**New components:** Basic creation test -> Property tests -> Behavior tests -> Edge cases -> Validation tests

**Extending components:** New property/behavior test -> Integration -> Edge cases

**Granularity check:** Independent commit? Tests ONE thing? Clear description?
