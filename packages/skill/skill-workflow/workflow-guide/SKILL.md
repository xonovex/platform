---
name: workflow-guide
description: "Use when coordinating one explicit artifact-neutral workflow operation: create, review, revise, decide, execute, validate, publish, abandon, or isolated workspace create, merge, abandon, or cleanup. Triggers on workflow-command delegation, forwarding decision context across roles or provider comments, and operation-level requests that must select installed capabilities, preserve effect boundaries, or compose without a prescribed lifecycle, even when the user doesn't say 'workflow'."
---

# Workflow Operations

Perform one explicit operation while preserving its effect boundary. This skill owns
operation semantics; selected domain skills own the subject-specific procedure. The
architecture these operations implement is in
[references/contract.md](references/contract.md).

## Core Principles

- **One operation** — perform only the requested operation; never infer a lifecycle or
  silently chain the next one
- **Explicit subject** — act on the supplied inline subject, path, or opaque native
  reference without inventing missing identity or revision data
- **Cold-boundary handoffs** — chain freely inside one session; produce a handoff only
  when a session or role ends, in the shape
  [references/handoffs.md](references/handoffs.md) defines
- **Decisions carry anchors** — record what was decided, why, and where in the code, so
  a later session can find it
- **Soft capability selection** — choose relevant installed skills from the subject and
  their routing descriptions; keep this general skill free of concrete domain
  dependencies
- **Dependency-first composition** — load each selected guide's exact manifest
  dependencies before the guide itself, once per operation
- **Operation boundary wins** — use only the part of a selected procedure that fits the
  requested effect and persistence boundary; block when that procedure cannot be
  separated from a broader effect
- **Visible boundaries** — report degraded coverage, blockers, evidence, and every
  planned or observed effect instead of hiding uncertainty
- **Authority stays explicit** — see [references/governance.md](references/governance.md)
- **Workspace resources stay separate** — workspace operations manage isolation and
  integration resources, not the work performed inside them
- **No operation, no command** — a request that fits no operation gets none; a freeform
  session is the sanctioned way to work outside this catalog

## Operation Procedure

1. Identify the requested operation, exact subject, supplied context, constraints,
   requested effect mode, and any handoff relationships.
2. Reject or separate additional operations whose effect boundaries differ.
3. Resolve every active context reference, then resolve required and preferred
   capability needs and load relevant installed capabilities.
4. Reject stale, conflicting, invalid, or unresolved active context before effects.
5. Adapt selected procedures to the operation's effect and persistence boundary.
6. Perform only the requested operation without assuming publication, approval, or a
   following lifecycle step.
7. Preserve material identity, context, evidence, degradation, blockers,
   relationships, and effects in the returned handoff.

Use these headings only when they add information:

```markdown
## Outcome

Status: completed | partial | blocked

## Result

## Subject

## Decisions

## Evidence

## Effects

## Degradation or blockers
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
- **Execute** — carry out previously specified work under an explicit effect mode — see
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
- Caller-declared preferred capability needs degrade visibly when unavailable;
  caller-declared required needs block before effects
- A `--request` file follows the Markdown handoff contract; shorthand arguments remain
  useful for requests that do not need cross-role traceability
- A selected guide never widens the operation: persistence remains Publish, workspace
  removal remains Workspace cleanup, and mutation still requires `apply`
- Execute expects work that was already specified; a request with no antecedent belongs
  to Create or a freeform session
- Workspace merge never implies branch, reference, or worktree cleanup
- Provider-native references are opaque; pass them to the selected provider capability
  instead of parsing their shape

## Progressive Disclosure

- Read [references/contract.md](references/contract.md) - Load when judging whether an
  artifact, operation, handoff, or effect boundary belongs in the catalog at all
- Read [references/governance.md](references/governance.md) - Load when authority,
  approval, evidence, or fetched provider content is in question
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
  resolving soft selections, hard-dependency order, degradation, or blockers
- Read [references/handoffs.md](references/handoffs.md) - Load when a session or role ends
  and the next one needs the subject, decisions, and open issues
- Read [references/context-forwarding.md](references/context-forwarding.md) - Load when
  carrying decisions into later operations or provider-native notes
- Read [references/sdlc.md](references/sdlc.md) - Load when composing the atomic
  operations across the frozen scenario families
