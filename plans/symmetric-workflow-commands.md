---
type: plan
has_subplans: true
status: complete
updated: 2026-07-20
completed_date: "2026-07-20"
feature: symmetric-workflow-commands
dependencies:
  plans: []
  subplans:
    01-command-contract-and-inventory: []
    02-plan-skill-decoupling: [01-command-contract-and-inventory]
    03-role-reference-and-invocation-guides:
      [
        01-command-contract-and-inventory,
        02-plan-skill-decoupling,
        04-external-trigger-boundary,
      ]
    04-external-trigger-boundary: []
    05-validation-and-breaking-release:
      [
        01-command-contract-and-inventory,
        02-plan-skill-decoupling,
        03-role-reference-and-invocation-guides,
        04-external-trigger-boundary,
      ]
proposed_subplans:
  - 01-command-contract-and-inventory
  - 02-plan-skill-decoupling
  - 03-role-reference-and-invocation-guides
  - 04-external-trigger-boundary
  - 05-validation-and-breaking-release
parallel_groups:
  - group: 1
    plans: [01-command-contract-and-inventory, 04-external-trigger-boundary]
    note: "The command API and the removal of operator-owned triggering have independent implementation surfaces."
  - group: 2
    plans: [02-plan-skill-decoupling]
    depends_on: [01-command-contract-and-inventory]
    note: "Narrow the planning skill as soon as the generic command contract is fixed; trigger removal may continue independently."
  - group: 3
    plans: [03-role-reference-and-invocation-guides]
    depends_on: [02-plan-skill-decoupling, 04-external-trigger-boundary]
    note: "Document roles, references, and invocation using the completed command and skill boundaries."
  - group: 4
    plans: [05-validation-and-breaking-release]
    depends_on:
      [
        01-command-contract-and-inventory,
        02-plan-skill-decoupling,
        03-role-reference-and-invocation-guides,
        04-external-trigger-boundary,
      ]
    note: "Lock the new inventory and prepare the coordinated major release after every old surface is gone."
skills_to_consult:
  - command-guide
  - orthogonal-pattern-guide
  - plan-guide
  - skill-guide
  - testing-guide
  - typescript-guide
  - vitest-guide
  - kubernetes-guide
  - git-guide
research_sources:
  documentation: []
  versions:
    xonovex-platform-baseline: "ed367a4e4a8bb12c9e3251a18791cdc5041296ca"
    command-workflow: "6.1.0"
    skill-plan: "6.1.0"
    go: "1.26.0"
    controller-runtime: "0.24.1"
---

# Symmetric workflow commands

## Overview

Replace the role- and lifecycle-shaped workflow command catalog with a small,
symmetric operation API. Commands express what the caller wants to do;
independent dimensions select the subject, perspective, method, executor,
trigger, and provider. Roles become documented compositions over the same
operations instead of separate workflows. References remain native to their
provider, and triggering stays outside the command package.

This is an intentional breaking refactor. Remove the old commands, profiles,
approval gates, lifecycle governance, aliases, and compatibility paths rather
than adapting them to the new model. Retain execution-security controls such as
sandbox policy because they constrain a running workload rather than prescribe
a product or delivery workflow.

## Execution context

All paths and validation commands resolve from the Xonovex repository root. The
inspected baseline is `ed367a4e4a8bb12c9e3251a18791cdc5041296ca`; re-locate
named declarations if the checkout moves before execution. The working tree was
clean when this plan was created. Implementation must preserve unrelated user
changes if any appear later.

## Goals

- Expose eight consistent core operations: `create`, `review`, `revise`,
  `decide`, `execute`, `validate`, `publish`, and `abandon`.
- Keep workspace isolation as a separate, symmetric utility family:
  `workspace-create`, `workspace-merge`, `workspace-abandon`, and
  `workspace-cleanup`.
- Make operation, subject/result kind, perspective, method, executor, trigger,
  provider, and provider-native reference independent dimensions.
- Let PM/PO, UX, development, QA, and reviewers compose the same verbs while
  documenting their different subjects and perspectives.
- Keep PR review owned by a developer reviewer. QA can validate behavior and
  provide evidence without becoming the PR approval authority.
- Remove the remaining early-lifecycle, approval, authority, profile, and
  governed-tail assumptions from commands, skills, role guides, diagrams, and
  agent documentation.
- Remove operator-owned schedules and trigger receivers. Manual users, harness
  hooks, CI/CD hooks, schedulers, and other callers invoke commands or create
  `AgentRun` resources through their native integration.
- Strengthen validation so the repository cannot silently reintroduce an
  asymmetric command catalog or stale governance terminology.
- Release the result as a coordinated breaking major version without aliases,
  shims, or a migration document.

## Non-goals

- Defining a universal workflow, mandatory stage order, approval policy,
  governance runtime, or central command orchestrator.
- Defining universal PM, UX, developer, QA, or reviewer role permissions.
- Defining a universal A1/A2/A3 agent-maturity taxonomy. A harness may expose
  those labels as executor metadata, but the commands do not interpret them or
  turn them into gates.
- Inventing a Xonovex reference ID, provider registry, artifact store, or local
  persistence fallback.
- Moving every domain procedure into `command-workflow`. Domain skills remain
  optional capabilities selected for a subject, method, or provider.
- Removing `AgentPolicy`, shared sandbox policy, pod-security settings, or other
  runtime confinement controls.
- Renaming the `command-workflow` package or its marketplace plugin solely
  because its command inventory changes.
- Rewriting historical plans or artifacts to the new terminology.

## Current state

- `packages/command/command-workflow` contains 53 command prompts formed from
  artifact-specific verbs and a former governed lifecycle. Eighteen commands
  still name procedures that no installed skill owns, including acceptance,
  assessment, delivery, development, incident, integration, observation,
  release, retirement, review, and transition operations.
- Eighteen command and documentation files still expose `--profile`, even
  though no profile resolver or workflow runtime remains.
- The command README, role quickstarts, and four role diagrams still prescribe
  mandatory gates, authority, expiry, external enforcement, and a governed
  operational tail. The QA guide also assigns PR review to QA.
- `packages/skill/skill-plan/plan-guide` owns discovery, research,
  formulation, experience design, solution design, decisions, approval, and
  plan operations. Its references and evals still encode profiles, authority,
  provider-neutral persistence, and approval-gated subplans.
- The command documentation validator checks links and manifest dependency
  parity but does not derive or constrain the command inventory. Documentation
  command counts have already drifted from the files on disk while validation
  passes.
- The agent operator always registers `AgentTrigger` and `AgentSchedule`
  controllers and an HTTP trigger receiver. These resources fuse invocation
  mechanisms into the base execution operator even though callers can create
  `AgentRun` resources directly through the Kubernetes API.
- `packages/agent/AGENTS.md` still claims governance and oversight behavior
  that was removed with the earlier governance runtime refactor.
- `packages/shared/shared-agent-go` contains sandbox and execution policy but
  no workflow-runtime dependency or lifecycle machinery. It does not require a
  functional refactor for this work.

## Research findings and decisions

### 1. Separate the dimensions

The former catalog multiplied artifact types, lifecycle stages, roles, and
operations into command names. The replacement keeps each variation point
orthogonal:

| Dimension        | Meaning                                                                      | Owner                                |
| ---------------- | ---------------------------------------------------------------------------- | ------------------------------------ |
| Operation        | The requested transformation or effect                                       | Core command                         |
| Kind             | The subject/result type, such as plan, change, test result, or decision      | Caller or selected domain capability |
| Perspective      | The lens used during review or validation                                    | Caller or selected domain capability |
| Method           | A technique such as BDD, code review, or deterministic inventory generation  | Selected capability                  |
| Executor         | Human, script, LLM, or agent implementation                                  | Invoking environment                 |
| Agent capability | Optional provider/harness metadata such as A1/A2/A3                          | Agent provider or harness            |
| Trigger          | Manual, harness hook, CI/CD hook, schedule, webhook, or another event source | Invoking environment                 |
| Provider         | The system that reads or stores an artifact                                  | Selected provider capability         |
| Reference        | The provider-native locator and optional revision                            | Provider                             |

No dimension silently selects another. In particular, a hook does not imply an
agent, an agent maturity does not imply a workflow skill, a role does not imply
authority, and a reference format does not imply a universal provider.

### 2. Make eight operations the stable public API

| Operation  | Universal responsibility                                             | Explicitly does not own                        |
| ---------- | -------------------------------------------------------------------- | ---------------------------------------------- |
| `create`   | Produce a new result of the requested kind from supplied context     | A mandatory discovery or approval stage        |
| `review`   | Return findings from a named perspective against supplied criteria   | Revising, approving, or publishing the subject |
| `revise`   | Produce a new revision from a subject and explicit feedback          | Deciding whether the revision is accepted      |
| `decide`   | Record and return an optional decision and rationale                 | Gating later commands or granting authority    |
| `execute`  | Carry out a referenced intent using a selected capability            | Choosing hidden lifecycle stages               |
| `validate` | Compare evidence with explicit criteria and report the result        | PR approval or automatic remediation           |
| `publish`  | Create or update an explicit provider-native external result         | Silent persistence or unconfirmed side effects |
| `abandon`  | Record that work is being stopped and perform only requested cleanup | Global rollback or implied deletion            |

Exact flags and prompt schemas belong in the first child plan. The shared
semantic contract is fixed here: kinds, perspectives, methods, executors, and
providers remain open selections rather than centrally enumerated profiles.
Missing selected capabilities fail visibly.

The commands own their small universal operation contracts and may load an
explicitly selected domain or provider skill. They do not delegate to a new
umbrella workflow skill. This is a deliberate boundary: a central delegator
would recreate the workflow owner and coupling removed by this refactor.

### 3. Preserve workspace operations as a separate family

Worktree or workspace management changes execution isolation rather than an
artifact's lifecycle. Keep four sibling utilities with the same noun-verb
grammar and no hidden relationship to plan approval:

- `workspace-create`
- `workspace-merge`
- `workspace-abandon`
- `workspace-cleanup`

Core operations can run in the current workspace or one prepared by these
utilities. They never create, merge, or clean a workspace implicitly.

### 4. Replace old commands by semantics, not aliases

| Former capability                                                                              | New operation                                                  |
| ---------------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| Discovery, research, formulation, design, decision, or plan creation                           | `create --kind <kind>`                                         |
| Critiques and PR analysis                                                                      | `review`                                                       |
| Design/decision/plan revision and PR refinement                                                | `revise`                                                       |
| Accept/reject commands                                                                         | `decide --outcome <value>`                                     |
| Plan continuation, development, and QA runs                                                    | `execute`                                                      |
| Plan, assessment, acceptance, and resolution checks                                            | `validate`                                                     |
| Git commit, PR creation/posting, and delivery publication                                      | `publish`                                                      |
| Development abandonment                                                                        | `abandon`                                                      |
| Plan worktree operations                                                                       | The `workspace-*` family                                       |
| Inventory generation                                                                           | `create --kind inventory` with a deterministic method          |
| Incident, integration, release, transition, observation, corrective, and retirement procedures | Optional domain kinds/capabilities over the generic operations |

Delete the former command files. Do not retain aliases, wrappers, deprecation
markers, redirects, or migration documentation.

### 5. Treat roles as lenses over one operation grammar

| Role/lens          | Typical composition                                          | Typical subjects or evidence                     |
| ------------------ | ------------------------------------------------------------ | ------------------------------------------------ |
| PM/PO              | create -> review -> revise -> decide                         | Discovery, research, formulation, decision, plan |
| UX                 | create -> review -> revise -> decide                         | Experience design, research evidence             |
| Developer          | create -> review -> revise -> execute -> validate -> publish | Solution design, plan, code change               |
| QA                 | create -> review -> execute -> validate -> publish           | Test plan, test result, defect evidence          |
| Developer reviewer | review -> publish                                            | Change and PR review result                      |

These are examples, not enforced pipelines. Any role can invoke any operation
when its local organization allows it. QA validates product behavior and
evidence; a different developer reviews and approves the pull request.

### 6. Keep references provider-native

A generic command accepts an opaque provider locator and, when relevant, a
provider-native revision. The selected provider capability interprets both.
The command may infer a provider or kind only when the reference is
unambiguous; otherwise the caller supplies it explicitly.

Creation can return an inline/session result with no reference. When persistence
or publication is requested, it returns the new provider-native reference and
revision information the provider exposes. Copying or publishing between
providers creates a new destination reference and preserves the source
relationship only when the provider supports it. There is no silent local-file
fallback.

### 7. Keep triggers and executors outside command semantics

Manual invocation, a harness hook, a CI/CD hook, a webhook, and a scheduler all
call the same command contract. A script, LLM, or agent can execute that call;
the command neither detects nor changes behavior based on how it was invoked.
Agent labels such as A1/A2/A3 may help a harness select an executor but are not
accepted as workflow stages, permissions, or governance levels.

At the Kubernetes boundary, the base operator should reconcile `AgentRun` and
execution-security resources only. Remove `AgentTrigger`, `AgentSchedule`, the
HTTP trigger receiver, their controllers, CRDs, RBAC, service, network policy,
generated code, tests, and the now-unused cron dependency. External systems use
their native event mechanism to create `AgentRun` resources. Do not extract a
new trigger plugin until multiple real trigger implementations establish a
reusable contract.

### 8. Narrow planning to a domain capability

`skill-plan` remains useful for planning and code research, but it stops owning
the delivery lifecycle. Retain planning research, creation, critique, revision,
subplan creation, continuation, update, validation, and code-research cleanup.
Remove early-lifecycle discovery/formulation, experience and solution design,
decision, plan accept/reject, profile, authority, and provider-persistence
contracts from its skill text, references, manifests, source notes, and evals.

Plan status becomes descriptive or provider-owned metadata, not a gate.
Subplans can be created from an explicitly selected plan revision without an
approval transition. Existing historical plan documents remain readable and
are not rewritten.

### 9. Retain execution security as an independent concern

`AgentPolicy` and `packages/shared/shared-agent-go/pkg/policy` constrain runtime
filesystem, process, pod, and tool behavior. They do not order work, assign
roles, approve artifacts, or resolve workflow references. Keep them, correct
their documentation if necessary, and add residue checks that distinguish
security policy from removed lifecycle governance.

## Proposed approach

1. Replace the command catalog and manifests with the eight core commands and
   four workspace utilities. Remove hard dependencies on plan, review, testing,
   or pull-request skills when no single capability is required by every
   command.
2. Reduce `skill-plan` to its domain boundary and rewrite its eval corpus to
   exercise planning rather than lifecycle governance.
3. Replace the four role-specific quickstarts and diagrams with one operation
   model plus concise role-lens, provider-reference, and invocation guides.
4. Remove the agent operator's trigger/schedule subsystem and stale governance
   instructions while preserving the `AgentRun` and sandbox-security APIs.
5. Upgrade documentation and cross-package validation to derive the real
   command inventory, reject old concepts, and exercise every affected package.
6. Update marketplace metadata, package manifests, lockfiles, changelogs, and
   coordinated release metadata for the next major version through the normal
   version-packages PR flow.

## Affected surfaces

- `packages/command/command-workflow`
- `packages/skill/skill-plan`
- `packages/diagram/diagram-agent-workflow`
- `packages/script/script-moon-skill-validate`
- `packages/agent/agent-operator-go`
- `packages/agent/AGENTS.md`
- `packages/shared/shared-agent-go` (validation and documentation only)
- `.claude-plugin/marketplace.json`
- `.agents/plugins/marketplace.json`
- root package and release metadata affected by the coordinated version change

## Proposed child plans

1. `01-command-contract-and-inventory` freezes prompt inputs/outputs and replaces
   the command, manifest, and dependency inventory.
2. `02-plan-skill-decoupling` narrows the planning skill, references, and evals.
3. `03-role-reference-and-invocation-guides` replaces role workflows and
   diagrams with the orthogonal model and provider-native reference contract.
4. `04-external-trigger-boundary` removes operator-owned triggers/schedules and
   stale agent governance claims while preserving execution security.
5. `05-validation-and-breaking-release` adds residue/inventory tests, runs the
   full validation matrix, and prepares coordinated major-version metadata.

## Risk assessment

| Risk                                                         | Impact                                           | Mitigation                                                                                                         |
| ------------------------------------------------------------ | ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------ |
| Generic commands become vague or grow into god prompts       | Callers receive inconsistent results             | Give each operation one narrow invariant, require explicit criteria/capabilities, and test input/output contracts  |
| Open kinds and providers reduce discoverability              | Users do not know what to select                 | Provide concise examples and capability discovery without a closed registry or profile system                      |
| Provider inference resolves an ambiguous locator incorrectly | Work is read from or written to the wrong system | Infer only unambiguous formats; otherwise require an explicit provider and confirm external effects                |
| Removing aliases breaks existing invocations                 | Existing automation fails on upgrade             | Treat the release as intentionally breaking, list removals in the changelog, and do not claim compatibility        |
| Removing operator trigger CRDs breaks deployed manifests     | Existing clusters require coordinated changes    | Surface every removed Kubernetes kind in release notes and require external callers to create `AgentRun` directly  |
| Planning becomes too permissive after approval removal       | Organizations lose a desired local control       | Let provider, CI, or repository policy implement approval without embedding it in the reusable command/skill layer |
| Governance residue survives under new wording                | The package remains implicitly opinionated       | Validate forbidden concepts across commands, planning skill, role docs, diagrams, agent instructions, and evals    |
| Security policy is accidentally removed with governance      | Runtime confinement weakens                      | Treat sandbox policy as an explicit non-goal and run shared-agent/operator security tests unchanged                |

## Success criteria

- [x] The command plugin exposes exactly eight core operations and four
      workspace utilities with symmetric names and documented contracts.
- [x] No command name embeds a role, lifecycle stage, artifact-specific
      transition, profile, provider, trigger, executor, or maturity level.
- [x] PM/PO, UX, developer, QA, and developer-reviewer documentation uses the
      same verbs and clearly separates QA validation from PR review.
- [x] Manual, hook, CI/CD, scheduler, LLM, script, and agent invocation are
      documented as independent composition choices rather than command modes.
- [x] A1/A2/A3 or other maturity labels are neither defined nor enforced by the
      command package.
- [x] References are provider-native and opaque to generic commands; no central
      resolver, universal ID, or silent persistence fallback exists.
- [x] `skill-plan` contains planning and code-research guidance only and has no
      profiles, approval gates, authority model, or early-lifecycle ownership.
- [x] `AgentTrigger`, `AgentSchedule`, and the operator HTTP trigger receiver are
      removed; external event sources can create `AgentRun` directly.
- [x] Agent sandbox and execution policy remains intact and is not described as
      lifecycle governance.
- [x] Old commands, aliases, compatibility wrappers, migration documents, stale
      diagrams, and matching eval cases are absent.
- [x] Validators derive the command inventory from disk and fail on missing,
      extra, asymmetric, or stale surfaces.
- [x] All affected Moon checks, Go tests, formatting checks, and repository
      residue searches pass without warnings.
- [x] Marketplace and release metadata describe the new model and prepare a
      coordinated major release through the existing PR-only process.

## Validation strategy

- Run `npx moon run command-workflow:ci-check`.
- Run `npx moon run skill-plan:ci-check`.
- Run `npx moon run diagram-agent-workflow:ci-check`.
- Run `npx moon run agent-operator-go:ci-check`.
- Run `npx moon run shared-agent-go:ci-check` to prove retained security policy
  still works and no workflow dependency was introduced.
- Run the cross-package skill/command validation project and its Vitest suite.
- Search tracked files for every removed command, `--profile`, governed-tail,
  approval-gate, `AgentTrigger`, `AgentSchedule`, and deleted workflow-runtime
  identifier, excluding historical plans and changelogs where explicitly
  appropriate.
- Validate Kubernetes generation, CRD/RBAC manifests, Go formatting, generated
  code consistency, package manifests, marketplace metadata, and lockfiles.
- Run `git diff --check` and confirm the worktree contains only intended files.

## Estimated effort

Large. The prompt inventory itself is small, but this is a coordinated breaking
change across command contracts, planning-skill content and evals, role
documentation, diagrams, Kubernetes APIs/controllers, validation tooling, and
release metadata. The five child plans isolate those risks and keep the command
and trigger-boundary work parallel where safe.
