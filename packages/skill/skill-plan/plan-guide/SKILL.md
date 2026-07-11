---
name: plan-guide
description: "Use when scoping, researching, deciding, planning, revising, critiquing, continuing, updating, or validating a feature, refactor, or analysis task — covers the whole plan-document lifecycle plus codebase research (alignment, hardening, simplification, template extraction, TODO scanning, shared-library design). Triggers on prompts about planning, designing, scoping, breaking down, architecting, settling or interrogating requirements and decisions, stress-testing / critiquing a plan, code analysis / cleanup / hardening / alignment, even when the user doesn't say 'plan'."
---

# Planning & Code-Research Guidelines

Author and maintain plan documents across their full lifecycle (research → decide → create → revise ⇄ critique → subplans → continue → update → validate) and run code-research operations (align, harden, simplify, template extraction, etc.) that feed into those plans.

## Core Principles

- **Research first** — analyze before authoring; never plan blind
- **Reports, not code** — all research/analysis operations are read-only and generate reports for `plan-create` to consume; only `plan-continue` modifies the codebase
- **Validation required** — every plan's success criteria back-checks with typecheck / lint / build / tests
- **Skills to consult** — every plan lists which guideline skills implementers must read first
- **Use available tools** — prefer environment-provided task tracking, file search, and code analysis over working from memory

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
2. **Decide** — `plan-decide` settles decisions one at a time: walks known open decisions as prose briefs, or discovers unknown ones by questioning the user down the design tree (codebase-aware) when nothing is queued
3. **Create** — `plan-create` authors the parent plan; test-first plans apply **tdd-guide** (or **bdd-guide** for acceptance-first)
4. **Revise** — `plan-revise` applies user feedback to the plan document (annotations + prompt instructions); `--final` marks it approved
5. **Critique** — `plan-critique` adversarially stress-tests the plan (red-team / pre-mortem), feeding findings back into revise
6. **Expand** — `plan-subplans-create` generates detailed child plans
7. **Execute** — `plan-continue` works through subplans one at a time
8. **Update** — `plan-update` refreshes status / phase / validation results
9. **Validate** — `plan-validate` confirms success criteria are met (read-only)

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

- Read [references/plan-decide.md](references/plan-decide.md) - Load when settling decisions one at a time — walking known open decisions as prose briefs, or discovering unknown ones by questioning the user, before or after a plan exists
- Read [references/plan-create.md](references/plan-create.md) - Load when authoring a high-level plan from research (test-first plans route to **tdd-guide** / **bdd-guide**)
- Read [references/plan-revise.md](references/plan-revise.md) - Load when applying inline annotations and/or prompt feedback to a plan document
- Read [references/plan-critique.md](references/plan-critique.md) - Load when adversarially stress-testing a plan to expose weaknesses (red-team / pre-mortem / falsify / steelman), read-only
- Read [references/plan-subplans-create.md](references/plan-subplans-create.md) - Load when expanding an approved plan into detailed parallelizable subplans
- Read [references/plan-continue.md](references/plan-continue.md) - Load when resuming implementation work from an existing plan
- Read [references/plan-update.md](references/plan-update.md) - Load when refreshing a plan with current status / validation / progress
- Read [references/plan-validate.md](references/plan-validate.md) - Load when verifying a plan's success criteria are met (read-only)
