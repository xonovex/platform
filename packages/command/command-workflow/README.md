# Composable Workflow Commands

Provider-native lifecycle commands for discovery through retirement. Canonical result
meaning stays stable while profiles independently select methods, executors, providers,
workspaces, and governance.

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

Commands load their owning guideline skills at runtime. Plugin dependencies guarantee the
core workflow, planning, governance, review, and testing capabilities are installed.

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
resume resolves native references; conversation and runtime traces are not persistent
identity.

| Command                                             | Description                                                                                      |
| --------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `discovery-run`                                     | Iterate observations, assumptions, affected context, and unknowns without forcing stories        |
| `research-run`                                      | Produce reusable evidence, provenance, confidence, uncertainty, and bounded synthesis            |
| `formulation-run`                                   | Formulate candidate behavior, examples, constraints, and ambiguities with a neutral default      |
| `experience-design-{create,critique,revise,accept}` | Manage an optional Experience Design result at exact revisions                                   |
| `solution-design-{create,critique,revise,accept}`   | Manage an optional Solution Design result at exact revisions                                     |
| `decision-{create,critique,revise,accept}`          | Keep evidence, recommendation, authority, and supersession separate                              |
| `workflow-onboard-advise`                           | Recommend profile-compatible methods, skills, providers, executors, and modules without applying |

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

## Delivery and governance

| Command                                        | Description                                                                       |
| ---------------------------------------------- | --------------------------------------------------------------------------------- |
| `acceptance-validate`                          | Validate an exact deliverable revision using neutral or selected criteria methods |
| `git-commit`                                   | Commit/push through an installed Git capability                                   |
| `plan-worktree-{create,merge,abandon,cleanup}` | Optional Git-worktree workspace operations                                        |
| `pr-create`                                    | Open a provider-native pull/merge request through the detected host adapter       |
| `pr-review-{analyze,refine,post,resolve}`      | Produce, refine, publish, and resolve review findings                             |
| `workflow-inspect`                             | Inspect workflow results, profile topology, evidence, and completion gaps         |
| `workflow-governance-inspect`                  | Inspect effective policies, modules, authority, enforcement, and exceptions       |
| `workflow-conformance`                         | Validate workflow and governance semantic contracts                               |
| `workflow-drift`                               | Compare intended and observed governance state                                    |
| `workflow-modules`                             | Inspect or manage governance modules through native adapters                      |

## Design decisions

- Lifecycle commands depend on semantic result and provider ports, not provider formats.
- Neutral methods are available without story/Gherkin skills; specialist methods remain selectable.
- Deterministic collection is preferred, model synthesis is bounded, agents are reserved for adaptive exploration, and human/qualified authority is never fabricated.
- Local files, Git repositories, hosted trackers, and databases are peers selected by profile/context; none is the universal fallback.

[View workflow diagram](../../diagram/diagram-agent-workflow/workflow-diagram.png)
