---
name: plan-guide
description: "Use when scoping, researching, interrogating, planning, clarifying, refining, critiquing, continuing, updating, or validating a feature, refactor, or analysis task — covers the whole plan-document lifecycle plus codebase research (alignment, hardening, simplification, template extraction, TODO scanning, shared-library design). Triggers on prompts about planning, designing, scoping, breaking down, architecting, interrogating requirements, stress-testing / critiquing a plan, code analysis / cleanup / hardening / alignment, even when the user doesn't say 'plan'."
---

# Planning & Code-Research Guidelines

Author and maintain plan documents across their full lifecycle (research → interrogate → clarify → create → refine → critique → subplans → continue → update → validate) and run code-research operations (align, harden, simplify, template extraction, etc.) that feed into those plans.

## Core Principles

- **Research first** — analyze before authoring; never plan blind
- **Reports, not code** — research operations produce reports; only `plan-continue` modifies the codebase
- **Validation required** — every plan's success criteria back-checks with typecheck / lint / build / tests
- **Skills to consult** — every plan lists which guideline skills implementers must read first
- **Read-only research** — exploration / analysis operations never modify files; they generate reports for `plan-create` to consume
- **Use available tools** — if the environment provides task tracking, file search, code analysis, or other tools, use them rather than working from memory or assumptions

## Gotchas

- Skipping `plan-research` and going straight to `plan-create` produces vague plans built on assumed context
- A plan without `skills_to_consult` leaves implementers ignoring project conventions
- Auto-detecting toolchain via `package.json` only misses Moon/Makefile-driven projects — check both
- Approving a parent plan is mandatory before generating subplans (`plan-subplans-create`)
- "Tests pass" doesn't mean "success criteria met" — `plan-validate` reads the criteria, not just exit codes
- Auto-continuing to the next plan after completion silently chains work — `plan-continue` STOPS after one
- Subplans with >7 tasks risk silent drops — target 5–7 tasks each
- Skipping the verification re-read before marking complete is the #1 cause of incomplete work

## Plan Lifecycle

1. **Research** — `plan-research` for general; for a code-quality audit (hardening / simplification / alignment) it applies the **code-quality-guide** dimensions and reports findings
2. **Interrogate** — `plan-interrogate` surfaces unknown decisions by walking the design tree one question at a time (codebase-aware), upstream of clarify
3. **Clarify** — `plan-clarify` walks known open decisions one by one in plain prose (after research, or after create to settle direction)
4. **Create** — `plan-create` authors the parent plan; test-first plans apply **tdd-guide** (or **bdd-guide** for acceptance-first)
5. **Refine** — `plan-refine` processes user feedback (annotations + prompt instructions)
6. **Critique** — `plan-critique` adversarially stress-tests the plan (red-team / pre-mortem), feeding findings back into refine
7. **Expand** — `plan-subplans-create` generates detailed child plans
8. **Execute** — `plan-continue` works through subplans one at a time
9. **Update** — `plan-update` refreshes status / phase / validation results
10. **Validate** — `plan-validate` confirms success criteria are met (read-only)

## Code-Research Operations

- **Audit code quality (harden / simplify / align)** — `plan-research` applies the **code-quality-guide** dimensions and reports findings for `plan-create`
- **Remove barrel exports** — see [references/code-barrels-remove.md](references/code-barrels-remove.md)
- **Remove non-essential comments** — see [references/code-comments-remove.md](references/code-comments-remove.md)
- **Extract shared utilities** — see [references/code-shared-extract.md](references/code-shared-extract.md)
- **Extract reusable templates** — see [references/code-template-extract.md](references/code-template-extract.md)
- **Scaffold from templates** — see [references/code-template-scaffold.md](references/code-template-scaffold.md)
- **Scan / group TODO comments** — see [references/todos.md](references/todos.md)

## Plan Operations

- **Research** the codebase + web — see [references/plan-research.md](references/plan-research.md)
- **Interrogate** to surface unknown decisions — see [references/plan-interrogate.md](references/plan-interrogate.md)
- **Clarify** known open decisions one by one — see [references/plan-clarify.md](references/plan-clarify.md)
- **Create** a plan with research — see [references/plan-create.md](references/plan-create.md)
- **Refine** from feedback (annotations / prompt) — see [references/plan-refine.md](references/plan-refine.md)
- **Critique** the plan adversarially — see [references/plan-critique.md](references/plan-critique.md)
- **Generate subplans** from an approved parent — see [references/plan-subplans-create.md](references/plan-subplans-create.md)
- **Continue** work from an existing plan — see [references/plan-continue.md](references/plan-continue.md)
- **Update** plan status / progress — see [references/plan-update.md](references/plan-update.md)
- **Validate** plan achievement — see [references/plan-validate.md](references/plan-validate.md)

## Progressive Disclosure

### Research

- Read [references/plan-research.md](references/plan-research.md) - Load when researching codebase + web for a future plan, or running a read-only code-quality audit (harden / simplify / align — applies code-quality-guide)
- Read [references/code-barrels-remove.md](references/code-barrels-remove.md) - Load when analyzing barrel exports for removal
- Read [references/code-comments-remove.md](references/code-comments-remove.md) - Load when identifying non-essential comments
- Read [references/code-shared-extract.md](references/code-shared-extract.md) - Load when finding duplicated patterns to extract
- Read [references/code-template-extract.md](references/code-template-extract.md) - Load when creating reusable templates from existing code
- Read [references/code-template-scaffold.md](references/code-template-scaffold.md) - Load when generating new code from templates
- Read [references/todos.md](references/todos.md) - Load when scanning and grouping TODO comments

### Plan lifecycle

- Read [references/plan-interrogate.md](references/plan-interrogate.md) - Load when interrogating the user to surface unknown decisions before a plan exists, one question at a time, exploring the codebase to self-answer
- Read [references/plan-clarify.md](references/plan-clarify.md) - Load when walking the user through known open decisions one at a time, in plain prose, after research or plan creation
- Read [references/plan-create.md](references/plan-create.md) - Load when authoring a high-level plan from research (test-first plans route to **tdd-guide** / **bdd-guide**)
- Read [references/plan-refine.md](references/plan-refine.md) - Load when iterating on a plan from inline annotations and/or prompt feedback
- Read [references/plan-critique.md](references/plan-critique.md) - Load when adversarially stress-testing a plan to expose weaknesses (red-team / pre-mortem / falsify / steelman), read-only
- Read [references/plan-subplans-create.md](references/plan-subplans-create.md) - Load when expanding an approved plan into detailed parallelizable subplans
- Read [references/plan-continue.md](references/plan-continue.md) - Load when resuming implementation work from an existing plan
- Read [references/plan-update.md](references/plan-update.md) - Load when refreshing a plan with current status / validation / progress
- Read [references/plan-validate.md](references/plan-validate.md) - Load when verifying a plan's success criteria are met (read-only)
