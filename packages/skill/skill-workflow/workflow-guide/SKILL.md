---
name: workflow-guide
description: "Use when work needs a stage-neutral create, review, revise, decide, execute, validate, publish, or abandon contract; exact multi-provider inputs; durable continuation; story decomposition; policy-shaped SDLC composition; or isolated workspace creation, integration, abandonment, or cleanup. Triggers on operation selection, structured handoffs, criteria or perspective resolution, provider-native references and revisions, preview/apply effects, partial or unknown outcomes, and cross-session workflow state, even when the user doesn't say 'workflow architecture'."
---

# Composable Workflow Operations

Perform one explicit operation on exact inputs. Return a human summary plus the
machine-readable `OperationResult` defined in
[references/contracts.md](references/contracts.md); compose later calls from that
result instead of conversational inference.

## Core Principles

- **Keep layers distinct** — operation, semantic selection, resource binding,
  invocation context, and derived implementation have different owners, see
  [references/composition.md](references/composition.md)
- **Bind every resource independently** — preserve each provider-native reference and
  revision; use named bindings for multi-provider work, see
  [references/provider-native-references.md](references/provider-native-references.md)
- **Publish domain results separately** — non-publish operations return inline
  results; only Publish persists or transmits a domain result
- **Resolve criteria visibly** — assisted resolution is the default; record source,
  reason, confidence, and binding status without promoting advice, see
  [references/criteria-and-perspectives.md](references/criteria-and-perspectives.md)
- **Derive implementation** — derive skills, adapters, and capabilities from the
  request and policy; consume the host resolver result, stop when blocked, report
  degradation, load its ordered guides, and honor only exact expert overrides, see
  [references/skill-resolution.md](references/skill-resolution.md)
- **Separate intent from authority** — effect mode requests inspect, preview, or
  apply; deterministic runtime policy verifies authority, approvals, idempotency,
  budgets, and retries, see
  [references/effects-and-authority.md](references/effects-and-authority.md)
- **Checkpoint administration separately** — runtime-managed durable checkpoints are
  policy-authorized administrative records, not domain-result publication, see
  [references/durable-work.md](references/durable-work.md)

## Core Operations

- **Create** — produce a new inline result without changing a source — see
  [references/create.md](references/create.md)
- **Review** — return evidence-linked judgment about an exact subject — see
  [references/review.md](references/review.md)
- **Revise** — produce an inline successor from explicit feedback — see
  [references/revise.md](references/revise.md)
- **Decide** — record a descriptive outcome without granting authority — see
  [references/decide.md](references/decide.md)
- **Execute** — inspect, preview, or apply one bounded action — see
  [references/execute.md](references/execute.md)
- **Validate** — return criterion-by-criterion evidence without changing the subject
  — see [references/validate.md](references/validate.md)
- **Publish** — preview or apply persistence of one exact result to one destination —
  see [references/publish.md](references/publish.md)
- **Abandon** — return why work stopped and its safe retry boundary — see
  [references/abandon.md](references/abandon.md)

## Workspace Transactions

- **Workspace create** — preview or create one isolated workspace — see
  [references/workspace-create.md](references/workspace-create.md)
- **Workspace merge** — preview or integrate one workspace without removing it — see
  [references/workspace-merge.md](references/workspace-merge.md)
- **Workspace abandon** — report abandonment without removing resources — see
  [references/workspace-abandon.md](references/workspace-abandon.md)
- **Workspace cleanup** — exclusively preview or remove named workspace resources —
  see [references/workspace-cleanup.md](references/workspace-cleanup.md)

## Gotchas

- A provider binding is a typed envelope around an opaque native locator, not a
  universal identifier.
- An explicit perspective selects a lens; it does not make that lens's suggested
  criteria binding.
- An apply request is effect intent, not proof of authorization.
- Execute and workspace transactions may perform their named effects, but they never
  persist their `OperationResult`; Publish owns that separate effect.
- A ready decision, passing validation, or approved review never implies merge,
  release, deployment, or publication authority.
- Same-conversation shorthand may use only the exact immediately preceding result;
  continuation elsewhere needs a persisted native binding.

## Progressive Disclosure

- Read [contracts](references/contracts.md) - Load when performing any operation
- Read [composition](references/composition.md) - Load when composing operations
- Read [bindings](references/provider-native-references.md) - Load when resolving resources
- Read [criteria](references/criteria-and-perspectives.md) - Load when resolving semantics
- Read [skill resolution](references/skill-resolution.md) - Load when deriving implementation
- Read [catalog](assets/composition-catalog.json) - Load when matching skill provisions
- Read [effects](references/effects-and-authority.md) - Load when an effect is possible
- Read [invocation](references/invocation-and-execution.md) - Load when calls start or run
- Read [role lenses](references/role-lenses.md) - Load when a role should suggest concerns
- Read [durable work](references/durable-work.md) - Load when work crosses sessions
- Read [story decomposition](references/story-decomposition.md) - Load when splitting stories
- Read [SDLC composition](references/sdlc-composition.md) - Load when policy defines a graph
- Read [recovery](references/recovery.md) - Load when outcomes are partial or uncertain
- Read [references/create.md](references/create.md) - Load when performing Create
- Read [references/review.md](references/review.md) - Load when performing Review
- Read [references/revise.md](references/revise.md) - Load when performing Revise
- Read [references/decide.md](references/decide.md) - Load when performing Decide
- Read [references/execute.md](references/execute.md) - Load when performing Execute
- Read [references/validate.md](references/validate.md) - Load when performing Validate
- Read [references/publish.md](references/publish.md) - Load when performing Publish
- Read [references/abandon.md](references/abandon.md) - Load when performing Abandon
- Read [workspace create](references/workspace-create.md) - Load when creating a workspace
- Read [workspace merge](references/workspace-merge.md) - Load when integrating a workspace
- Read [workspace abandon](references/workspace-abandon.md) - Load when abandoning a workspace
- Read [workspace cleanup](references/workspace-cleanup.md) - Load when removing a workspace
