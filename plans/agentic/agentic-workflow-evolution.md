---
type: plan
has_subplans: true
status: approved
updated: 2026-07-03
feature: agentic-workflow-evolution
dependencies:
  plans:
    - plans/agentic/platform-workflow-improvement.md
  subplans:
    01-step-result-contract: []
    02-workflow-definitions: [01-step-result-contract]
    03-workflow-next-command: [02-workflow-definitions]
    04-agent-cli-workflow-run: [02-workflow-definitions]
    05-workflowrun-crd: [04-agent-cli-workflow-run]
    06-sensors-and-outer-loops: [05-workflowrun-crd]
    07-workflow-diagram-refresh: [03-workflow-next-command, 04-agent-cli-workflow-run]
proposed_subplans:
  - 01-step-result-contract
  - 02-workflow-definitions
  - 03-workflow-next-command
  - 04-agent-cli-workflow-run
  - 05-workflowrun-crd
  - 06-sensors-and-outer-loops
  - 07-workflow-diagram-refresh
parallel_groups:
  - group: 1
    plans: [01-step-result-contract]
    note: "The contract every later layer consumes; nothing else starts
      until results are machine-readable."
  - group: 2
    plans: [02-workflow-definitions]
    depends_on: [1]
  - group: 3
    plans: [03-workflow-next-command, 04-agent-cli-workflow-run]
    depends_on: [2]
    note: "L1 (prompt interpreter) and L2 (Go interpreter) implement the
      same semantics against the same definitions; independent codebases."
  - group: 4
    plans: [05-workflowrun-crd, 07-workflow-diagram-refresh]
    depends_on: [3]
    note: "L3 engineering builds on the L2 interpreter package. Diagram
      refresh documents the landed L1/L2 shape."
  - group: 5
    plans: [06-sensors-and-outer-loops]
    depends_on: [4]
skills_to_consult:
  - plan-guide
  - skill-guide
  - command-guide
  - moon-guide
  - kubernetes-guide
  - git-guide
  - testing-guide
  - fp-guide
  - shell-scripting-guide
  - zod-guide
research_sources:
  documentation:
    - packages/diagram/diagram-agent-workflow/workflow-diagram.dot
    - packages/diagram/diagram-agent-workflow/target-architecture.dot
    - packages/diagram/diagram-agent-workflow/maturity-ladder.dot
    - ../../drodan/drodan-platform/plans/drodan/drodan-multi-cluster-rollout.md
    - ../../drodan/drodan-platform/packages/drodan/drodan-docs/architecture/drodan-multi-cluster-kubernetes-blueprint.dot
  versions:
    agent-operator-go: "functional v1alpha1: AgentRun, AgentHarness,
      AgentProvider, AgentWorkspace, AgentToolchain, AgentPolicy;
      gVisor/Kata runtime classes; default-deny egress; webhooks"
    agent-cli-go: "run command with isolation x provision x network axes;
      headless mode supported; worktree support; claude + opencode via
      shared-agent-go registries"
    drodan-fleet: "five-cluster Talos/Cilium rollout approved 2026-07-02;
      agent runtime targets the ci cluster (subplan 05 of that plan)"
---

# Agentic workflow evolution

## Overview

Evolve the plan/PR command workflow from manually driven (L0) through
assisted (L1) and supervised-autonomous (L2) to cluster-autonomous (L3),
per `packages/diagram/diagram-agent-workflow/target-architecture.dot` and
`maturity-ladder.dot`. Commands become reusable workflow steps with a
machine-readable result contract; declarative workflow definitions with
policy-carrying gates let the same steps run manually, semi-automatically,
or autonomously without rewriting anything. Follow-up to
`plans/agentic/platform-workflow-improvement.md`, which supplies the enabling
layer (standardized `[plan-file]` args, `plan-list` discovery, CI gates,
thin commands). Only the parent's subplans 01-04 (gates, references,
distillation, plumbing) are true prerequisites; its catalog subplans
(05-catalog-dedup, 06-sources-eval-backfill) are not, and this plan may
start once the parent's 04-workflow-plumbing merges. Cross-repo
execution order: see `plans/agentic/roadmap.md`.

## Execution Context

Cross-repo ordering, the ecosystem glossary (skill/command package
anatomy, moon commands, lockstep releases, workflow/step/gate/L0-L3
terminology, plan-status conventions), and the per-subplan execution
protocol live in `plans/agentic/roadmap.md` — read it before any subplan. When
a subplan cites "decision N", the Design Decisions below are normative.
Each subplan is self-contained given its own Context section, this
decision log, and its `skills_to_consult` — assume no other session
context exists.

## Design Decisions (2026-07-03)

1. **Deterministic orchestration**: no LLM decides the control flow.
   Intelligence lives inside steps and in gate reviewers; the orchestrator
   reads definitions + state, computes the next step, dispatches, and
   stops at gates.
2. **One semantics, three interpreters**: L1 `workflow-next` command
   (in-session), L2 `agent-cli workflow run` (headless Go loop), L3
   `WorkflowRun` controller (agent-operator-go). L2 and L3 share a Go
   interpreter package in shared-agent-go.
3. **Gates carry policy**: `manual | assisted | auto`, per gate, with
   reviewer agents (plan-critique, pr-review-analyze) and an audit trail.
   Progressive automation = flipping gate policies, never editing flow.
4. **Git is the source of truth**: plan documents hold instance state; an
   append-only run journal (`plans/<plan>/runs/*.yaml`) holds step
   results. Cluster CRs hold run mechanics plus mirrored step-result
   summaries (the reconciler's working state); the journal in git remains
   the durable record.
5. **Agents stop at the PR**: no agent mutates a cluster. Deliverables
   move as commits/PRs through GitLab CI (Trivy, cosign v3) and Flux;
   runtime signals return only through sensors.
6. **Agent stack placement**: the whole stack (operator, WorkflowRun
   controller, AgentRun jobs, sensor receiver) runs on the drodan ci
   cluster; mgmt hosts no agent stack and holds no spoke credentials
   (drodan rollout decision 5). Egress: deny-all plus a named
   model-provider API exception.
7. **Custom-thin over Argo**: the model (plan docs + a dozen steps +
   gates + foreach/loop) is far smaller than Argo Workflows; the operator
   and L1/L2 must run outside Kubernetes anyway. Revisit if step DAGs
   outgrow sequence + foreach + bounded loop.
8. **Repo split**: engine work (this plan) lives in xonovex-platform;
   fleet deployment (kubernetes package, flux tree, egress exception,
   Alertmanager receiver, GitLab webhook) is a separate
   `agentic-workflow-runtime` plan to be created in
   drodan-platform/plans/drodan/, dependent on rollout subplans 03/04/05.
9. **Runners provision the plugin set**: steps are marketplace plugins;
   every runner installs a pinned xonovex plugin set into the session's
   harness environment (sandbox home for L2; agent image or workspace
   init for L3), versioned like the nix toolchain. Without this, no step
   exists inside a headless or cluster session.
10. **Layered gate authority** (settled 2026-07-03): changing a gate
    policy requires human review via protected `workflows/**` paths
    (CODEOWNERS-style); at L3 an `AgentPolicy` gate-autonomy floor caps
    the maximum policy per namespace; most-restrictive-wins precedence.
    Interpreters read gate policies from the trusted base ref (main),
    never from the run's working branch.
11. **Notify via Discord, approve on the artifact** (settled 2026-07-03):
    gate halts and escalations send a Discord webhook message with a
    direct link (same channel pattern as drodan alert routing); the
    audited approval is recorded on the gate's natural surface - PR
    approval where a PR exists, the journal/gate verdict for pre-PR
    gates. Email is not used for workflow notifications.
12. **Runner-scoped commit identity** (settled 2026-07-03): L0/L1
    commits are the operator's; L2/L3 commits use a dedicated bot
    identity with a `Workflow-Run:` trailer linking to the journal,
    signed, and pushable only to `agent/*` branches. Key management:
    local keychain at L2; cluster Secret at L3 (drodan-side).

## Goals

- Every workflow step ends with a machine-readable result (status,
  evidence, artifacts, suggested next) that any outer loop can consume.
- `workflows/*.yml` encodes the current feature workflow declaratively:
  steps, named gates with policies, foreach over parallel subplan groups,
  bounded loops, budgets.
- L1: `workflow-next` suggests or runs the correct next step in-session.
- L2: `agent-cli workflow run` executes a plan headless - fresh sandboxed
  session per step, worktrees for parallel groups - stopping at manual/
  assisted gates, enforcing budgets and iteration caps.
- L3 engineering: `WorkflowRun` CRD + controller mapping steps to
  AgentRuns and parallel groups to AgentWorkspace worktrees, gates as CR
  conditions; tested via envtest/e2e (deployment is drodan-side).
- Sensors: a webhook receiver that turns CI/observability events into
  proposed workflow runs, with dedup and cooldown; CI Sentinel and
  Validation Auditor as the first outer-loop agents.
- The L0 experience is unchanged throughout: same commands, same skills,
  same plan documents.

## Current State

- 40 thin commands over plan/git/pr skills; state in plan frontmatter +
  `git config branch.<branch>.plan`; sessions are disposable and resume
  from git state (plan-continue) - the durable-state model already works.
- agent-cli-go runs claude/opencode headless in sandboxes (bwrap/docker,
  nix-pinned toolchains, network none/proxy); no loop/workflow mode.
- agent-operator-go is functional (v1alpha1, webhooks, e2e incl. gVisor/
  Kata) with nothing orchestrating above AgentRun.
- The workflow exists only as documentation
  (`diagram-agent-workflow/workflow-diagram.dot`); decisions are unnamed
  inline "(HITL or AI)" diamonds; step outcomes are session prose.
- Drodan fleet: five-cluster rollout approved; ci cluster (runner,
  autoscaler, deny-all egress baseline) is the agent runtime target; obs
  Alertmanager routes Discord/email and can add a sensor receiver.

## Research Findings

- The three maturity ingredients already exist (steps, git state,
  runners); the missing layer is contract + definition + orchestrator +
  sensors - all additive.
- controller-runtime's workqueue/requeue model fits the WorkflowRun
  reconciler; no external engine or broker needed at this scale.
- Drodan rollout constraints adopted: mgmt-no-spoke-credentials rules out
  mgmt-hosted orchestration; ci egress baseline requires an explicit
  provider-API exception; Alertmanager (decision 13) is the one
  observability signal source, so a single receiver covers Mimir ruler,
  Falco, and Flagger events.
- Known toolchain caveat: controller-gen is broken with Go 1.25+ in this
  repo; CRDs and deepcopy are maintained manually - the WorkflowRun CRD
  subplan inherits that practice.

## Proposed Approach

1. **01-step-result-contract** - result schema (status, summary,
   evidence, artifacts, suggested next), journal convention
   `plans/<plan>/runs/<run-id>-<step>.yaml` where run-id embeds the
   subplan/branch so parallel worktrees cannot collide on merge;
   plan-guide documentation, step idempotency requirement stated;
   plan-* / pr-* / git-commit commands emit results only when a workflow
   context is active (branch plan association set, or an explicit flag) -
   casual manual runs write nothing; plan-update sweeps journal files
   into commits; missing/invalid-result handling is part of the
   contract: a runner synthesizes a `failed` result when a session ends
   without a valid result file; schema validation added to the ci-check
   gates from the parent plan.
2. **02-workflow-definitions** - `workflows/feature.yml` encoding the
   current diagram (steps, named gates with `manual|assisted|auto`
   policies and reviewers, foreach over parallel groups, bounded loops,
   per-step retry caps, budgets); definition schema + validation; a new
   `workflow-guide` skill owning definition semantics, gate policy
   rules, and the authority mechanics from decision 10 (protected
   `workflows/**` paths, trusted-ref policy reads, most-restrictive-wins
   against the L3 floor).
3. **03-workflow-next-command** - new thin command + skill operation: read
   definition + plan + journal, compute the next step; suggest (default)
   or invoke (`--auto` within current gate policy, executing at most ONE
   step per session and handing off to a fresh session - preserving the
   session-per-step model and keeping L1 semantics identical to L2); wire
   reviewer agents into assisted gates.
4. **04-agent-cli-workflow-run** - Go interpreter in
   `shared-agent-go/pkg/workflow` (parse definition, compute next,
   evaluate gate policy) + `agent-cli workflow run`: headless loop, fresh
   sandboxed session per step with the pinned plugin set provisioned into
   the sandbox home (decision 9), worktree-per-parallel-subplan, journal
   writes, synthesized `failed` results + bounded retries on session
   death, budget/iteration enforcement, commits under the decision-12
   bot identity with `Workflow-Run:` trailers, stop-and-notify at
   manual/assisted gates (Discord webhook with a direct link, decision
   11).
5. **05-workflowrun-crd** - `WorkflowRun` CRD + reconciler in
   agent-operator-go reusing the pkg/workflow interpreter: steps ->
   AgentRuns, parallel groups -> AgentWorkspace worktrees, gates -> CR
   conditions (flipped by human annotation/PR approval or reviewer
   AgentRun), budgets/timeouts, envtest + e2e coverage; `AgentPolicy`
   gains the gate-autonomy floor field (decision 10), enforced
   most-restrictive-wins in the reconciler. State access is
   reconciler-side by design: each step's AgentRun reports its result
   summary into WorkflowRun status (completion hook / termination
   message) so the controller reconciles from CR status and never mounts
   the workspace or clones the repo; the plugin set is baked into the
   agent image or provisioned by the workspace init job (decision 9).
6. **06-sensors-and-outer-loops** - sensor receiver (GitLab pipeline
   events; Alertmanager webhook format) with fingerprint dedup and
   cooldown, mapping events to proposed WorkflowRuns (paused at a gate by
   default); CI Sentinel and Validation Auditor agent definitions; Fleet
   Supervisor rules (stuck detection, budget cap, kill-switch) escalating
   via the decision-11 Discord channel.
7. **07-workflow-diagram-refresh** - apply the recorded changes to
   `workflow-diagram.dot` (named gates with policies, journal artifact,
   trigger entry point, runner variants, loop caps, delivery hand-off);
   extend the moon `graph-build` task to render all `.dot` files in the
   package.

## Risk Assessment

- **Prose-vs-structure tension**: the result contract adds rigidity to
  prompt-based commands. Mitigation: additive (a final block/file),
  harness-neutral plain files, validated in CI rather than at runtime.
- **Parallel runs conflicting on plan docs**: journal is append-only
  per-run files; plan frontmatter updates serialize through merge steps.
- **Runaway loops / cost**: every loop has a max, every run a budget, the
  Fleet Supervisor a global cap and kill-switch; sensors dedup by
  fingerprint with cooldown so a flapping alert cannot fan out.
- **Interpreter drift between L1 (prompt) and L2/L3 (Go)**: workflow-guide
  is the single semantics document; the Go package is the reference
  implementation; L1 stays suggest-first where ambiguity is cheap.
- **Credential blast radius (cluster)**: the decision-12 bot identity -
  signed commits, `Workflow-Run:` trailers, push restricted to agent/*
  branches, protected main; enforcement designed in 05 and deployed
  drodan-side.
- **Two-repo coordination**: L3 deployment value only lands with the
  drodan-side runtime plan; engineering here is testable without it
  (envtest/e2e), so the boundary costs no schedule coupling.
- **No Go guideline skill exists** for the plan's largest work items
  (04/05); accepted gap - fp-guide is the nearest proxy, and a
  `go-guide` skill is queued as catalog follow-up work.

## Success Criteria

- [ ] Running any plan-* command manually is unchanged (L0 invariant);
      with a workflow context active it also writes a schema-valid
      result to the run journal; without one it writes nothing.
- [ ] `workflows/feature.yml` validates and encodes the current workflow
      diagram, including gates and parallel groups.
- [ ] L1: on a real plan mid-flight, `workflow-next` names the correct
      next step and honors gate policy (suggests at manual, runs at
      auto).
- [ ] L2: `agent-cli workflow run` takes a two-subplan test plan from
      research to merge-gate headless, one sandboxed session per step
      with the pinned plugin set provisioned, journal complete, and
      stops at the manual gate; exceeding the budget or iteration cap
      halts the run; killing a step session mid-run yields a synthesized
      `failed` result and a bounded retry.
- [ ] L3: WorkflowRun e2e maps a definition to sequenced AgentRuns with a
      gate condition requiring approval before the next step; parallel
      group runs use one AgentWorkspace with worktrees.
- [ ] A simulated GitLab pipeline-failed webhook produces exactly one
      proposed WorkflowRun (dedup verified) paused at its entry gate.
- [ ] workflow-diagram.dot shows named gates/policies, the trigger entry
      point, runners, and loop caps; `moon run
      diagram-agent-workflow:graph-build` renders all diagrams.
- [ ] `npx moon run :lint :typecheck :test :build` green, including Go
      packages.

## Estimated Effort

~4-6 weeks elapsed: 01 ~3-4 days, 02 ~3-4 days (incl. gate-authority
clarify), 03 ~2-3 days, 04 ~1.5 weeks (Go interpreter + CLI loop),
05 ~1.5 weeks (CRD + reconciler + e2e), 06 ~1 week, 07 ~1-2 days.
