---
name: workflow-guide
description: "Use when selecting, composing, executing, or explaining stage-neutral workflow operations and workspace utilities across independent kind, perspective, method, capability, trigger, executor, provider, reference, revision, criteria, and result dimensions. Triggers on create, review, revise, decide, execute, validate, publish, abandon, provider-native destinations, inline results, workspace lifecycle, operation ordering, or questions about which workflow operation fits, even when the user doesn't say 'workflow architecture'."
---

# Symmetric Workflow Operations

Perform one explicit operation on one subject, then compose additional calls only when the work requires them. Operations are siblings rather than lifecycle stages.

## Core Principles

- **Select one operation** — choose exactly one of the eight core verbs for each call; role, stage, trigger, executor, and maturity never change its semantics, see [references/composition.md](references/composition.md)
- **Keep dimensions independent** — resolve kind, perspective, method, capability, executor, trigger, provider, reference, and result separately, see [references/composition.md](references/composition.md)
- **Preserve native references** — let the selected provider interpret opaque locators, revisions, authentication, idempotency, and side effects, see [references/provider-native-references.md](references/provider-native-references.md)
- **Load capabilities narrowly** — use only explicitly selected or unambiguous domain, method, and provider capabilities; stop rather than substitute when an explicit capability is unavailable
- **Return inline by default** — persist only when the caller explicitly supplies a destination, then return its provider-native locator and revision
- **Separate invocation from execution** — manual calls, hooks, CI/CD, webhooks, and schedules share the same operation contract; humans, tools, models, and agents do not redefine it, see [references/invocation-and-execution.md](references/invocation-and-execution.md)
- **Keep authority explicit** — a descriptive decision grants no authority, and every publishing, cleanup, merge, or other side effect requires exact scope and authorization
- **Keep workspaces orthogonal** — core operations never create, merge, abandon, or clean a workspace implicitly

## Core Operations

- **Create** — produce a new result without changing the source — see [references/create.md](references/create.md)
- **Review** — evaluate an exact subject and report evidence-linked findings — see [references/review.md](references/review.md)
- **Revise** — produce a traceable new revision from explicit feedback — see [references/revise.md](references/revise.md)
- **Decide** — record a descriptive outcome and rationale without granting authority — see [references/decide.md](references/decide.md)
- **Execute** — carry out one bounded subject without implicit publication or workspace changes — see [references/execute.md](references/execute.md)
- **Validate** — check an exact subject against explicit criteria without changing it — see [references/validate.md](references/validate.md)
- **Publish** — publish an exact subject to one explicit destination — see [references/publish.md](references/publish.md)
- **Abandon** — stop work while preserving the reason, partial result, and retry boundary — see [references/abandon.md](references/abandon.md)

## Workspace Utilities

- **Workspace create** — create one isolated workspace from an exact source — see [references/workspace-create.md](references/workspace-create.md)
- **Workspace merge** — validate and merge one workspace into one destination — see [references/workspace-merge.md](references/workspace-merge.md)
- **Workspace abandon** — preserve recoverable state before optional removal — see [references/workspace-abandon.md](references/workspace-abandon.md)
- **Workspace cleanup** — preview and remove only explicitly selected targets — see [references/workspace-cleanup.md](references/workspace-cleanup.md)

## Gotchas

- Composition examples are illustrative; any core operation may stand alone, repeat, or appear in another order.
- `review` produces judgment and findings; `validate` returns a result for every explicit criterion.
- `decide` never approves, rejects, authorizes, promotes, or changes a gate.
- `execute` never publishes or manages a workspace implicitly; request those effects separately.
- `publish` changes a destination but never revises the source or implies approval.
- A workspace merge is a workspace utility, not a lifecycle stage or proof of acceptance.
- A missing selected provider or capability is an error; never fall back silently to a local file, another provider, or an invented umbrella capability.

## Progressive Disclosure

- Read [references/composition.md](references/composition.md) - Load when selecting operations or composing independent dimensions and calls
- Read [references/provider-native-references.md](references/provider-native-references.md) - Load when resolving or persisting opaque subjects, revisions, evidence, or result destinations
- Read [references/invocation-and-execution.md](references/invocation-and-execution.md) - Load when a manual call, hook, CI/CD event, webhook, schedule, human, tool, model, or agent initiates or performs an operation
- Read [references/role-lenses.md](references/role-lenses.md) - Load when applying a PM, UX, developer, QA, or reviewer perspective without creating role-specific commands
- Read [references/create.md](references/create.md) - Load when producing a new result without changing its source
- Read [references/review.md](references/review.md) - Load when evaluating an exact subject and reporting evidence-linked findings
- Read [references/revise.md](references/revise.md) - Load when producing a new revision from explicit feedback
- Read [references/decide.md](references/decide.md) - Load when recording a descriptive outcome without granting authority
- Read [references/execute.md](references/execute.md) - Load when carrying out one bounded subject
- Read [references/validate.md](references/validate.md) - Load when checking an exact subject against explicit criteria
- Read [references/publish.md](references/publish.md) - Load when publishing an exact subject to an explicit provider destination
- Read [references/abandon.md](references/abandon.md) - Load when stopping work while preserving reason, partial result, and retry state
- Read [references/workspace-create.md](references/workspace-create.md) - Load when creating one explicitly targeted isolated workspace
- Read [references/workspace-merge.md](references/workspace-merge.md) - Load when validating and merging one workspace into one explicit destination
- Read [references/workspace-abandon.md](references/workspace-abandon.md) - Load when abandoning one workspace while preserving recoverable state
- Read [references/workspace-cleanup.md](references/workspace-cleanup.md) - Load when previewing and removing explicitly selected stale or merged workspaces
