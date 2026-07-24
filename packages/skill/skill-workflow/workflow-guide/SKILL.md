---
name: workflow-guide
description: "Use when coordinating one explicit artifact-neutral workflow operation: create, review, revise, decide, execute, validate, publish, abandon, or isolated workspace create, merge, abandon, or cleanup. Triggers on workflow-command delegation and operation-level requests that must select installed capabilities, preserve effect boundaries, or compose without a prescribed lifecycle, even when the user doesn't say 'workflow'."
---

# Workflow Operations

Perform one explicit operation while preserving its effect boundary. This skill owns
operation semantics; selected domain skills own the subject-specific procedure.

## Core Principles

- **One operation** — perform only the requested operation; never infer a lifecycle or
  silently chain the next operation
- **Explicit subject** — act on the supplied inline subject, path, or opaque native
  reference without inventing missing identity or revision data
- **Soft capability selection** — choose relevant installed skills from the subject,
  method, perspectives, and criteria; keep this general skill free of concrete skill
  dependencies
- **Dependency-first composition** — resolve selected skills' hard dependencies,
  preferred requirements, and applicable overlays before loading each guide once
- **Visible boundaries** — report degraded coverage, blockers, evidence, and every
  planned or observed effect instead of hiding uncertainty
- **Authority stays explicit** — review, validation, and decision results are
  descriptive; they never authorize publication, integration, cleanup, or another
  protected effect
- **Workspace resources stay separate** — workspace operations manage isolation and
  integration resources, not the work performed inside them

## Operation Procedure

1. Identify the requested operation, exact subject, constraints, and requested effect
   mode.
2. Reject or separate additional operations whose effect boundaries differ.
3. Select and load relevant installed capabilities using the subject, method,
   perspectives, and criteria.
4. Perform only the requested operation without assuming publication, approval, or a
   following lifecycle step.
5. Verify the result and report evidence, degradation, blockers, and effects.

Use these headings only when they add information:

```markdown
## Outcome

Status: completed | partial | blocked

## Result

## Evidence

## Effects

## Degradation or blockers

## Retry boundary
```

Do not force empty headings or serialize the result into a fixed envelope.

## Core Operations

- **Create** — produce a new inline result without publishing it — see
  [references/create.md](references/create.md)
- **Review** — return evidence-linked findings without changing the subject — see
  [references/review.md](references/review.md)
- **Revise** — produce a traceable successor while preserving the source — see
  [references/revise.md](references/revise.md)
- **Decide** — record one descriptive outcome without granting authority — see
  [references/decide.md](references/decide.md)
- **Execute** — perform bounded work under an explicit effect mode — see
  [references/execute.md](references/execute.md)
- **Validate** — evaluate binding criteria independently with evidence — see
  [references/validate.md](references/validate.md)
- **Publish** — persist one exact result to one explicit destination — see
  [references/publish.md](references/publish.md)
- **Abandon** — return the reason, partial state, and retry boundary — see
  [references/abandon.md](references/abandon.md)

## Workspace Operations

- **Workspace create** — create only the named isolated workspace resources — see
  [references/workspace-create.md](references/workspace-create.md)
- **Workspace merge** — validate and integrate without cleanup — see
  [references/workspace-merge.md](references/workspace-merge.md)
- **Workspace abandon** — record abandonment without mutation — see
  [references/workspace-abandon.md](references/workspace-abandon.md)
- **Workspace cleanup** — preview and remove only exact named resources — see
  [references/workspace-cleanup.md](references/workspace-cleanup.md)

## Gotchas

- A request containing several verbs still needs separate operation boundaries when
  their effects differ
- Installing a skill does not load it; load every selected guide explicitly
- Missing preferred capability degrades coverage; missing required capability blocks
  before effects
- Preview is not authorization to apply, and a successful decision is not approval
- Workspace merge never implies branch, reference, or worktree cleanup
- Provider-native references are opaque; pass them to the selected provider capability
  instead of parsing their shape

## Progressive Disclosure

- Read [references/create.md](references/create.md) - Load when creating an inline result
- Read [references/review.md](references/review.md) - Load when reviewing a subject
- Read [references/revise.md](references/revise.md) - Load when applying explicit feedback
- Read [references/decide.md](references/decide.md) - Load when recording a descriptive decision
- Read [references/execute.md](references/execute.md) - Load when executing bounded work
- Read [references/validate.md](references/validate.md) - Load when evaluating binding criteria
- Read [references/publish.md](references/publish.md) - Load when persisting an exact result
- Read [references/abandon.md](references/abandon.md) - Load when stopping work without cleanup
- Read [references/workspace-create.md](references/workspace-create.md) - Load when creating isolated workspace resources
- Read [references/workspace-merge.md](references/workspace-merge.md) - Load when integrating a workspace without cleanup
- Read [references/workspace-abandon.md](references/workspace-abandon.md) - Load when abandoning a workspace without mutation
- Read [references/workspace-cleanup.md](references/workspace-cleanup.md) - Load when removing exact workspace resources
- Read [references/effects.md](references/effects.md) - Load when an operation may read,
  preview, apply, publish, or otherwise affect external state
- Read [references/capability-selection.md](references/capability-selection.md) - Load when
  resolving capability order, overlays, degradation, or blockers
