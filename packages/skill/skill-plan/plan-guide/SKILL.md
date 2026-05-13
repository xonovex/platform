---
name: plan-guide
description: "Use when scoping, researching, planning, refining, continuing, updating, or validating a feature, refactor, or analysis task — covers the whole plan-document lifecycle plus codebase research (alignment, hardening, simplification, template extraction, TODO scanning, shared-library design). Triggers on prompts about planning, designing, scoping, breaking down, architecting, code analysis / cleanup / hardening / alignment, even when the user doesn't say 'plan'."
---

# Planning & Code-Research Guidelines

Author and maintain plan documents across their full lifecycle (research → create → refine → subplans → continue → update → validate) and run code-research operations (align, harden, simplify, template extraction, etc.) that feed into those plans.

## Core Principles

- **Research first** — analyze before authoring; never plan blind
- **Reports, not code** — research operations produce reports; only `plan-continue` modifies the codebase
- **Validation required** — every plan's success criteria back-checks with typecheck / lint / build / tests
- **Skills to consult** — every plan lists which guideline skills implementers must read first
- **Read-only research** — exploration / analysis operations never modify files; they generate reports for `plan-create` to consume

## Gotchas

- Skipping `plan-research` and going straight to `plan-create` produces vague plans built on assumed context
- A plan without `skills_to_consult` leaves implementers ignoring project conventions
- Auto-detecting toolchain via `package.json` only misses Moon/Makefile-driven projects — check both
- Approving a parent plan is mandatory before generating subplans (`plan-subplans-create`)
- "Tests pass" doesn't mean "success criteria met" — `plan-validate` reads the criteria, not just exit codes
- Auto-continuing to the next plan after completion silently chains work — `plan-continue` STOPS after one

## Plan Lifecycle

1. **Research** — `plan-research` for general; `code-align` / `code-harden` / `code-simplify` for specific aspects
2. **Create** — `plan-create` (or `plan-tdd-create`) authors the parent plan
3. **Refine** — `plan-refine` processes user feedback (annotations + prompt instructions)
4. **Expand** — `plan-subplans-create` generates detailed child plans
5. **Execute** — `plan-continue` works through subplans one at a time
6. **Update** — `plan-update` refreshes status / phase / validation results
7. **Validate** — `plan-validate` confirms success criteria are met (read-only)

## Code-Research Operations

- **Align two implementations** — see [references/code-align.md](references/code-align.md)
- **Harden for production** — see [references/code-harden.md](references/code-harden.md)
- **Simplify / deduplicate** — see [references/code-simplify.md](references/code-simplify.md)
- **Remove barrel exports** — see [references/code-barrels-remove.md](references/code-barrels-remove.md)
- **Remove non-essential comments** — see [references/code-comments-remove.md](references/code-comments-remove.md)
- **Extract shared utilities** — see [references/code-shared-extract.md](references/code-shared-extract.md)
- **Extract reusable templates** — see [references/code-template-extract.md](references/code-template-extract.md)
- **Scaffold from templates** — see [references/code-template-scaffold.md](references/code-template-scaffold.md)
- **Scan / group TODO comments** — see [references/todos.md](references/todos.md)

## Plan Operations

- **Research** the codebase + web — see [references/plan-research.md](references/plan-research.md)
- **Create** a plan with research — see [references/plan-create.md](references/plan-create.md)
- **Create a TDD plan** — see [references/plan-tdd-create.md](references/plan-tdd-create.md)
- **Refine** from feedback (annotations / prompt) — see [references/plan-refine.md](references/plan-refine.md)
- **Generate subplans** from an approved parent — see [references/plan-subplans-create.md](references/plan-subplans-create.md)
- **Continue** work from an existing plan — see [references/plan-continue.md](references/plan-continue.md)
- **Update** plan status / progress — see [references/plan-update.md](references/plan-update.md)
- **Validate** plan achievement — see [references/plan-validate.md](references/plan-validate.md)

## Progressive Disclosure

### Research

- Read [references/plan-research.md](references/plan-research.md) - Load when researching codebase + web for a future plan
- Read [references/code-align.md](references/code-align.md) - Load when comparing two similar implementations for alignment
- Read [references/code-harden.md](references/code-harden.md) - Load when researching type-safety / validation / logging / error-handling improvements
- Read [references/code-simplify.md](references/code-simplify.md) - Load when researching deduplication, dead code, and over-engineering
- Read [references/code-barrels-remove.md](references/code-barrels-remove.md) - Load when analyzing barrel exports for removal
- Read [references/code-comments-remove.md](references/code-comments-remove.md) - Load when identifying non-essential comments
- Read [references/code-shared-extract.md](references/code-shared-extract.md) - Load when finding duplicated patterns to extract
- Read [references/code-template-extract.md](references/code-template-extract.md) - Load when creating reusable templates from existing code
- Read [references/code-template-scaffold.md](references/code-template-scaffold.md) - Load when generating new code from templates
- Read [references/todos.md](references/todos.md) - Load when scanning and grouping TODO comments

### Plan lifecycle

- Read [references/plan-create.md](references/plan-create.md) - Load when authoring a high-level plan from research
- Read [references/plan-tdd-create.md](references/plan-tdd-create.md) - Load when authoring a TDD plan with story-based test steps
- Read [references/plan-refine.md](references/plan-refine.md) - Load when iterating on a plan from inline annotations and/or prompt feedback
- Read [references/plan-subplans-create.md](references/plan-subplans-create.md) - Load when expanding an approved plan into detailed parallelizable subplans
- Read [references/plan-continue.md](references/plan-continue.md) - Load when resuming implementation work from an existing plan
- Read [references/plan-update.md](references/plan-update.md) - Load when refreshing a plan with current status / validation / progress
- Read [references/plan-validate.md](references/plan-validate.md) - Load when verifying a plan's success criteria are met (read-only)
