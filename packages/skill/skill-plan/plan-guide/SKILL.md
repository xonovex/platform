---
name: plan-guide
description: "Use when researching, creating, critiquing, revising, expanding, continuing, updating, or validating an implementation plan. Triggers on plan or subplan work, plan-driven implementation, codebase research explicitly intended to support a future plan, success-criteria validation, cross-session decision context, and resuming work from a plan reference, even when the user doesn't say 'plan'."
---

# Implementation Planning Guidelines

Provide implementation-planning procedures selected by a caller. This skill owns planning behavior and planning-oriented research, not general code-quality audits, invocation, lifecycle governance, or persistence.

## Core Principles

- **Explicit subjects**: accept an inline plan or an opaque provider-native reference plus an optional native revision
- **Provider-owned references**: let a selected provider resolve, version, or relate native resources; return operation results inline
- **Research first**: ground plans in codebase evidence, constraints, uncertainty, and relevant external sources
- **Read-only analysis**: planning research, critique, and validation report findings without changing their subjects; only continuation implements planned work
- **Skills to consult**: plans name applicable implementation capabilities, and continuation loads them before editing
- **Evidence-based validation**: check explicit success criteria and Definition of Done evidence, not merely command exit codes

Operation boundaries, effect modes, handoff shape, decision anchors, and authority
belong to the caller; this skill adds no rule of its own on any of them.

## Planning Operations

- **Research**: investigate codebase and external evidence for a future plan, see [references/research.md](references/research.md)
- **Create**: author one high-level plan and stop before detailed subplans, see [references/create.md](references/create.md)
- **Critique**: independently stress-test one plan without revising it, see [references/critique.md](references/critique.md)
- **Revise**: apply explicit feedback to a new plan revision, see [references/revise.md](references/revise.md)
- **Expand**: derive focused child plans and their execution ordering, see [references/expand.md](references/expand.md)
- **Continue**: implement one plan or child plan and stop, see [references/continue.md](references/continue.md)
- **Update**: refresh descriptive progress and validation evidence, see [references/update.md](references/update.md)
- **Validate**: evaluate explicit success criteria without changing the plan, see [references/validate.md](references/validate.md)

## Gotchas

- An opaque reference is provider data, never infer its storage shape or silently replace its provider with a local file
- Status is optional descriptive metadata: do not use approval or status fields to authorize or gate another planning operation
- Successful validation is evidence about criteria, not approval, acceptance, publication, or permission to proceed
- A plan without `skills_to_consult` leaves implementers without project-specific guidance
- A cross-role handoff that omits the source subject, relationships, criteria, or evidence breaks traceability even when its prose is understandable
- Toolchain discovery limited to one manifest misses workspace-level or task-runner validation
- Critique needs fresh independent context; continuation needs reconstructed subject context after session loss
- Independent critique with supplied context uses a blind first pass and a
  context-aware second pass; it preserves and compares both
- Unresolved, conflicting, stale, or instruction-bearing active provider context is
  never silently applied
- Expansion may use any explicit parent plan, regardless of whether it has an approval field
- Continuation completes one target and stops instead of silently chaining into the next child
- Inspect and preview continuation never edit files, provider resources, or plan state
- Subplans with more than seven tasks risk silent drops: target five to seven focused tasks
- Completion is measured against the tasks and success criteria; green tests alone do not prove it

## Progressive Disclosure

### Planning

- Read [references/research.md](references/research.md) - Load when researching codebase and external evidence explicitly for a future plan
- Read [references/create.md](references/create.md) - Load when authoring a high-level implementation plan from explicit inputs
- Read [references/decide.md](references/decide.md) - Load when settling an open decision before or after a plan exists
- Read [references/critique.md](references/critique.md) - Load when independently stress-testing an exact plan without changing it
- Read [references/revise.md](references/revise.md) - Load when applying explicit feedback or annotations to a plan
- Read [references/expand.md](references/expand.md) - Load when expanding a parent plan into detailed child plans
- Read [references/accept.md](references/accept.md) - Load when approving a plan for execution
- Read [references/reject.md](references/reject.md) - Load when rejecting a plan and recording why
- Read [references/continue.md](references/continue.md) - Load when resuming implementation from an existing plan
- Read [references/delegate.md](references/delegate.md) - Load when supervising a roadmap's execution by handing each item to an implementation agent and verifying the result
- Read [references/update.md](references/update.md) - Load when refreshing a plan with current progress and validation evidence
- Read [references/validate.md](references/validate.md) - Load when checking a plan's success criteria and Definition of Done without mutation
