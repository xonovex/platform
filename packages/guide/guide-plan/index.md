---
name: plan-guidelines
description: Trigger on planning sessions, research tasks, architecture decisions. Use when planning research and analysis tasks. Apply for code alignment, simplification, hardening, template extraction. Keywords: research, planning, task breakdown, code analysis, refactoring, template extraction, shared libraries, TODO scanning.
---

# Planning Research Guidelines

## Core Principles

- **Research First** - Analyze before creating implementation plans, see [reference/general-research.md](reference/general-research.md)
- **Use Task Agents** - Leverage Explore subagent with model=haiku, see [reference/general-research.md](reference/general-research.md)
- **Validation Required** - All analysis validates with typecheck/lint/build
- **Reports, Not Code** - Research generates reports without modifications

## Research Workflow

- **Launch Exploration** - Use Task tool with subagent_type=Explore and model=haiku
- **Analyze and Report** - Process results, identify patterns, document findings

## Planning Workflow

- **Create Implementation Plan** - Document architecture, technology choices, create tasks, see [reference/plan-create.md](reference/plan-create.md)
- **Break Down Work** - Use TaskCreate for each step with subject/description/activeForm
- **Define Dependencies** - Use TaskUpdate to set blockedBy/blocks relationships
- **Track Progress** - Use TaskList, TaskGet, TaskUpdate to manage execution

## Skills to Consult

- **Include in all plans** - Add "Skills to consult:" section with comma-separated list, see [reference/plan-create.md](reference/plan-create.md)

## Progressive Disclosure

### Research Analysis

- Read [reference/general-research.md](reference/general-research.md) - Research codebase and web for requirements
- Read [reference/code-align.md](reference/code-align.md) - Compare two similar implementations
- Read [reference/code-barrels-remove.md](reference/code-barrels-remove.md) - Analyze barrel exports for removal
- Read [reference/code-comments-remove.md](reference/code-comments-remove.md) - Identify non-essential comments
- Read [reference/code-harden.md](reference/code-harden.md) - Improve type safety and error handling
- Read [reference/code-shared-extract.md](reference/code-shared-extract.md) - Find duplicated patterns to extract
- Read [reference/code-simplify.md](reference/code-simplify.md) - Reduce code complexity
- Read [reference/code-template-extract.md](reference/code-template-extract.md) - Create reusable templates
- Read [reference/code-template-scaffold.md](reference/code-template-scaffold.md) - Generate from templates
- Read [reference/todos.md](reference/todos.md) - Scan and group TODO comments

### Plan Creation

- Read [reference/plan-create.md](reference/plan-create.md) - Create plan with task breakdown and dependencies
- Read [reference/plan-tdd-create.md](reference/plan-tdd-create.md) - Create TDD plan with sequential RED-GREEN-COMMIT tasks

## Task Management

- **TaskCreate** - Create with subject, description, activeForm
- **TaskUpdate** - Set dependencies (addBlockedBy/addBlocks), update status
- **TaskList** - View all tasks with status and dependencies
- **TaskGet** - Retrieve full details before starting work
- **Dependencies** - Independent (no blockedBy), Sequential (set blockedBy), TDD (strict order)
