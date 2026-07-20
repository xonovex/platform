# Lifecycle Workflow Commands

Independently invocable commands for discovery through retirement. Each command owns its
operation and loads only the guideline skills needed for that invocation.

## Guides

Role quickstarts are lenses over the same actor-neutral commands — no command knows a job
title. Each names the subset one perspective runs and the gates it answers for:

- [Developer quickstart](docs/developer-quickstart.md) — [diagram](../../diagram/diagram-agent-workflow/developer-workflow.png)
- [PM quickstart](docs/pm-quickstart.md) — [diagram](../../diagram/diagram-agent-workflow/pm-workflow.png)
- [QA quickstart](docs/qa-quickstart.md) — [diagram](../../diagram/diagram-agent-workflow/qa-workflow.png)
- [UX quickstart](docs/ux-quickstart.md) — [diagram](../../diagram/diagram-agent-workflow/ux-workflow.png)

## Installation

### Claude Code

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-workflow@xonovex-marketplace
```

### Codex

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-workflow@xonovex-marketplace
```

## Dependencies

Commands explicitly load their owning guideline skills. Plugin dependencies make planning,
review, code-quality, and testing guidance available; no command must use every capability.

Methods and native adapters are soft dependencies selected per operation. User stories,
BDD, example mapping, user research, accessibility, architecture, Git/worktrees, GitHub,
GitLab, databases, work-item trackers, and artifact providers are not universal
prerequisites. When an explicitly selected capability is unavailable, the command fails
visibly and names what is missing; it never silently substitutes a local file or provider.

## Early lifecycle

```text
Discovery -> Research -> Formulation -> optional Experience Design / Solution Design
         -> Decision -> Planning -> child Planning results -> Development
```

Each capability publishes its own provider-native result and opaque reference. Critique,
revision, and authority-bound acceptance remain independent operations. Fresh-context
resume resolves native references; conversation and execution traces are not persistent
identity.

| Command                                             | Description                                                                                 |
| --------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| `discovery-run`                                     | Iterate observations, assumptions, affected context, and unknowns without forcing stories   |
| `research-run`                                      | Produce reusable evidence, provenance, confidence, uncertainty, and bounded synthesis       |
| `formulation-run`                                   | Formulate candidate behavior, examples, constraints, and ambiguities with a neutral default |
| `experience-design-{create,critique,revise,accept}` | Manage an optional Experience Design result at exact revisions                              |
| `solution-design-{create,critique,revise,accept}`   | Manage an optional Solution Design result at exact revisions                                |
| `decision-{create,critique,revise,accept}`          | Keep evidence, recommendation, authority, and supersession separate                         |

## Planning and execution

| Command                       | Description                                                             |
| ----------------------------- | ----------------------------------------------------------------------- |
| `plan-research`               | Specialized read-only codebase/web research or code-quality analysis    |
| `plan-create`                 | Publish a high-level Planning result from opaque lifecycle references   |
| `plan-revise`                 | Apply feedback to an exact Planning revision and publish a new revision |
| `plan-critique`               | Independently stress-test an exact Planning revision                    |
| `plan-accept` / `plan-reject` | Record an authority-bound status decision against an exact revision     |
| `plan-subplans-create`        | Publish detailed child Planning results and dependency/execution groups |
| `plan-continue`               | Reconstruct native state and complete one actionable child result       |
| `plan-update`                 | Publish current status and exact-revision validation evidence           |
| `plan-validate`               | Validate success criteria and Definition of Done without mutation       |

## Development and assurance

| Command               | Description                                                                  |
| --------------------- | ---------------------------------------------------------------------------- |
| `develop-run`         | Execute exact Planning assignments and publish independent results           |
| `develop-consolidate` | Combine exact Development results without claiming Acceptance or Integration |
| `develop-abandon`     | Preserve partial Development state, evidence, reason, cleanup, and retry     |
| `deliver-publish`     | Publish a provider-native reviewable candidate at an immutable revision      |
| `inventory-generate`  | Generate deterministic SBOM, AI/ML/CBOM, service, or agent inventories       |
| `assessment-run`      | Assess any exact workflow result against pinned applicable criteria          |
| `review-run`          | Publish deliverable-specific findings with explicit reviewer independence    |
| `qa-run`              | Publish deliverable-specific test and environment evidence                   |

Development assigns exact Planning revisions to isolated workspaces and publishes one
result per assignment. Consolidation produces another Development result; it is not the
accepted-target mutation owned by Integration. Inventory facts come from deterministic or
external sources. Review, QA, and Assessment preserve evaluator origin and become stale
when their bound subject, policy, evaluator, or required environment changes.

## Acceptance and operational lifecycle

| Command                 | Description                                                                                       |
| ----------------------- | ------------------------------------------------------------------------------------------------- |
| `acceptance-validate`   | Assemble fresh exact-revision evidence without claiming accountable sign-off                      |
| `acceptance-decide`     | Record human Acceptance bound to subject, target, evidence, policy, actor, and expiry             |
| `integration-validate`  | Preflight authorization and protected target capabilities without mutation                        |
| `integration-run`       | Execute Integration only through an explicit externally enforced target capability                |
| `transition-run`        | Plan, execute, verify, or roll back data, users, providers, flags, support, and resilience        |
| `release-run`           | Execute, verify, roll back, or recover through controlled automation and protected environments   |
| `observe-run`           | Publish monitoring, user, security, AI, cost, accessibility, and delivery evidence                |
| `incident-run`          | Declare, update, contain, recover, escalate, or close an urgent Incident                          |
| `corrective-action-run` | Plan, execute, verify, and close corrective work with effectiveness and learning evidence         |
| `retirement-run`        | Retire models, data, credentials, features, APIs, infrastructure, dependencies, and configuration |

Evidence assembly may use bounded agents or models, but only a provider-authenticated
accountable human records Acceptance. Integration, target-changing Transition and Release,
data deletion, and Retirement revalidate exact subject, target, evidence, policy, actor, and
expiry bindings at a non-bypassable external enforcement point. Ordinary tool access is
not authorization. Exceptions and emergency exceptions remain scoped, expiring, compensated,
notified, revoked, and reviewed.

## Delivery

| Command                                        | Description                                                                 |
| ---------------------------------------------- | --------------------------------------------------------------------------- |
| `git-commit`                                   | Commit/push through an installed Git capability                             |
| `plan-worktree-{create,merge,abandon,cleanup}` | Optional Git-worktree workspace operations                                  |
| `pr-create`                                    | Open a provider-native pull/merge request through the detected host adapter |
| `pr-review-{analyze,refine,post,resolve}`      | Produce, refine, publish, and resolve review findings                       |

## Design decisions

- Commands delegate their procedures to declared guideline skills.
- Neutral methods are available without story/Gherkin skills; specialist methods remain selectable.
- Local files, Git repositories, hosted trackers, and databases are peers selected by context; none is the universal fallback.
