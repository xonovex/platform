---
name: plan-guide
description: "Use when researching, creating, reviewing, revising, expanding, continuing, updating, or validating implementation plans, or when auditing code for hardening, simplification, alignment, shared extraction, templates, or TODO inventory. Triggers on plan and subplan work, plan-driven implementation, codebase research for future work, and read-only code-quality analysis, even when the user doesn't say 'plan'."
---

# Planning and Code-Research Guidelines

Provide planning and code-research procedures selected by a caller or generic operation. This skill owns the planning behavior, not invocation, lifecycle governance, or persistence.

## Core Principles

- **Explicit subjects** — accept an inline plan or an opaque provider-native reference plus an optional native revision
- **Provider-owned references** — let a selected provider resolve, version, relate, or persist native resources; otherwise return inline results
- **Research first** — ground plans in codebase evidence, constraints, uncertainty, and relevant external sources
- **Read-only analysis** — research, audits, critique, and validation report findings without changing their subjects; only continuation implements planned work
- **One requested operation** — research, create, critique, revise, expand, continue, update, or validate without silently performing another operation
- **Skills to consult** — plans name applicable implementation capabilities, and continuation loads them before editing
- **Evidence-based validation** — check explicit success criteria and Definition of Done evidence, not merely command exit codes

## Planning Operations

- **Research** — investigate codebase and external evidence for a future plan — see [references/plan-research.md](references/plan-research.md)
- **Create** — author one high-level plan and stop before detailed subplans — see [references/plan-create.md](references/plan-create.md)
- **Critique** — independently stress-test one plan without revising it — see [references/plan-critique.md](references/plan-critique.md)
- **Revise** — apply explicit feedback to a new plan revision — see [references/plan-revise.md](references/plan-revise.md)
- **Expand** — derive focused child plans and their execution ordering — see [references/plan-subplans-create.md](references/plan-subplans-create.md)
- **Continue** — implement one plan or child plan and stop — see [references/plan-continue.md](references/plan-continue.md)
- **Update** — refresh descriptive progress and validation evidence — see [references/plan-update.md](references/plan-update.md)
- **Validate** — evaluate explicit success criteria without changing the plan — see [references/plan-validate.md](references/plan-validate.md)

## Gotchas

- An opaque reference is provider data — never infer its storage shape or silently replace its provider with a local file
- Status is optional descriptive metadata — do not use approval or status fields to authorize or gate another planning operation
- Successful validation is evidence about criteria, not approval, acceptance, publication, or permission to proceed
- A plan without `skills_to_consult` leaves implementers without project-specific guidance
- Toolchain discovery limited to one manifest misses workspace-level or task-runner validation
- Critique needs fresh independent context; continuation needs reconstructed subject context after session loss
- Expansion may use any explicit parent plan, regardless of whether it has an approval field
- Continuation completes one target and stops instead of silently chaining into the next child
- Subplans with more than seven tasks risk silent drops — target five to seven focused tasks
- Re-read tasks and success criteria before reporting completion; green tests alone do not prove the plan is done

## Progressive Disclosure

### Planning

- Read [references/plan-research.md](references/plan-research.md) - Load when researching codebase and external evidence for a future plan, or running a read-only code-quality audit
- Read [references/plan-create.md](references/plan-create.md) - Load when authoring a high-level implementation plan from explicit inputs
- Read [references/plan-critique.md](references/plan-critique.md) - Load when independently stress-testing an exact plan without changing it
- Read [references/plan-revise.md](references/plan-revise.md) - Load when applying explicit feedback or annotations to a plan
- Read [references/plan-subplans-create.md](references/plan-subplans-create.md) - Load when expanding a parent plan into detailed child plans
- Read [references/plan-continue.md](references/plan-continue.md) - Load when resuming implementation from an existing plan
- Read [references/plan-update.md](references/plan-update.md) - Load when refreshing a plan with current progress and validation evidence
- Read [references/plan-validate.md](references/plan-validate.md) - Load when checking a plan's success criteria and Definition of Done without mutation

### Code research

- Read [references/code-barrels-remove.md](references/code-barrels-remove.md) - Load when analyzing barrel exports for removal
- Read [references/code-comments-remove.md](references/code-comments-remove.md) - Load when identifying non-essential comments
- Read [references/code-shared-extract.md](references/code-shared-extract.md) - Load when finding duplicated patterns to extract
- Read [references/code-template-extract.md](references/code-template-extract.md) - Load when creating reusable templates from existing code
- Read [references/code-template-scaffold.md](references/code-template-scaffold.md) - Load when generating new code from templates
- Read [references/todos.md](references/todos.md) - Load when scanning and grouping TODO comments
