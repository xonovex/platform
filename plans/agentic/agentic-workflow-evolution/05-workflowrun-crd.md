---
type: plan
has_subplans: false
parent_plan: plans/agentic/agentic-workflow-evolution.md
parallel_group: 4
status: pending
dependencies:
  plans:
    - plans/agentic/agentic-workflow-evolution/04-agent-cli-workflow-run.md
    - plans/agent-security-hardening/01-operator-webhook-deployment.md
    - plans/agent-security-hardening/03-operator-fail-closed-enforcement.md
  files:
    - packages/agent/agent-operator-go/api/v1alpha1/**
    - packages/agent/agent-operator-go/internal/controller/**
    - packages/agent/agent-operator-go/internal/webhook/**
    - packages/agent/agent-operator-go/config/**
    - packages/agent/agent-operator-go/test/**
    - nix/agent-env.nix
skills_to_consult:
  - kubernetes-guide
  - testing-guide
  - fp-guide
  - docker-guide
  - moon-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 05 — WorkflowRun CRD: The L3 Interpreter (Engineering Only)

## Objective

Cluster-autonomous execution as tested engineering (deployment is the
drodan-side runtime plan): a `WorkflowRun` CRD + reconciler in
agent-operator-go reusing `pkg/workflow` (04), mapping steps to
AgentRuns and parallel groups to AgentWorkspace worktrees, gates as CR
conditions, reconciler-side state via step-result summaries in status
(parent critique resolution), the AgentPolicy gate-autonomy floor
(decision 10), and the plugin set baked into the agent image
(decision 9).

## Context (read this first — no other context is assumed)

Anchors as of `main` @ `2b276a7f` (2026-07-03); re-read cited files —
subplans 01–04 land first, and `plans/agent-security-hardening`
subplans 01 (webhook deployment) and 03 (fail-closed enforcement) are
HARD prerequisites (listed in this plan's dependencies.plans): they
make operator webhooks deployable and set the enforcement posture this
subplan's webhook and floor work build on. Verify both are
`status: complete` before starting.

1. Operator today: v1alpha1 AgentRun / AgentHarness / AgentProvider /
   AgentWorkspace / AgentToolchain / AgentPolicy; AgentRun phase machine
   Pending → Initializing → Running → {Succeeded, Failed, TimedOut};
   workspace path uses RWX PVC + worktree init. KNOWN TOOLCHAIN CAVEAT:
   controller-gen is broken with Go 1.25+ here — CRDs under
   `config/crd/bases/` and `zz_generated.deepcopy.go` are maintained
   MANUALLY (parent research finding); budget time for that.
2. **Reconciler-side state (critique resolution)**: the controller
   never mounts the workspace PVC and never clones the repo. Each
   step's AgentRun reports its result SUMMARY into WorkflowRun status —
   mechanism: the run's final container writes the result YAML to the
   pod termination message (`/dev/termination-log`, 4KiB cap — the
   summary shape must fit; evidence stays in the journal in git). The
   WorkflowRun reconciler watches owned AgentRuns and ingests
   termination messages on completion.
3. Gates as conditions: a gate step sets a `GateApproval` condition
   `status: False, reason: AwaitingApproval`; approval = human
   annotation (`workflow.xonovex.com/approve: <gate>`) validated by
   webhook against the gate's policy + the AgentPolicy floor; assisted
   gates spawn a reviewer AgentRun first and record its verdict in
   status.
4. Decision 10 floor: AgentPolicy gains `maxGateAutonomy`
   (`manual | assisted | auto`); effective policy =
   min(definition policy from trusted ref, floor); enforced in the
   reconciler and validated in the webhook (fail-closed, matching the
   operator's existing policy posture).
5. Decision 9 at L3: the pinned plugin set from 04's manifest is baked
   into the agent image (`nix/agent-env.nix` +
   `agent-operator-go:agent-image-build`, digest-pinned like the rest of
   the closure) — preferred over workspace-init installation for
   reproducibility; fall back to init-job materialization only if image
   size becomes a problem (record the measurement).
6. Definition + plan state enter the cluster as the repository the
   AgentWorkspace clones; the trusted-ref rule means the reconciler
   passes the definition CONTENT (from the WorkflowRun spec, set at
   creation from main) rather than letting runs re-read their branch.

## Tasks

1. **API** — `workflowrun_types.go`: spec (workflowName + resolved
   definition snapshot, planFile, workspaceRef, harness/provider refs,
   budgets), status (phase, currentSteps[], stepResults[] summaries,
   conditions incl. per-gate), manual deepcopy + CRD yaml.
2. **AgentPolicy floor** — add `maxGateAutonomy` to
   `agentpolicy_types.go` + defaulting/validation webhook +
   most-restrictive-wins resolution helper (shared with pkg/workflow's
   hook from 04).
3. **Reconciler** — drive pkg/workflow.Next() from status-held state:
   create AgentRuns per step (workspace worktrees for parallel groups),
   ingest termination-message results, synthesize failed on
   AgentRun Failed/TimedOut without a result (01's rule), honor retry
   caps and budgets, set gate conditions, resume on approval
   annotation.
4. **Webhooks** — validating: definition snapshot parses against the
   schema, gate approvals respect effective policy, budgets within
   AgentPolicy caps; defaulting: timeout/budget defaults. (Assumes the
   security-hardening plan's webhook deployment work — verify its
   status first.)
5. **Image plugin baking** — extend `nix/agent-env.nix` + the
   agent-image-build task to include the pinned plugin set at the
   harness's expected config path; record image-size delta.
6. **Tests** — envtest: reconciler state machine (sequence, parallel
   group, gate wait/approve, retry, budget halt); e2e (kind, fake
   agent image): a 2-step definition runs to a gate, approval
   annotation resumes it, termination-message results appear in status
   (parent success criterion 5).

## Validation Steps

- `npx moon run agent-operator-go:test` (envtest suite) green;
  e2e suite green on kind.
- Manual CRD/deepcopy consistency check (the repo's established
  practice, given controller-gen breakage).
- `:lint :typecheck :build` green; shared-agent-go untouched or
  re-tested if the floor hook changed it.

## Success Criteria

- [ ] WorkflowRun e2e: sequenced AgentRuns with a gate condition
      requiring approval; parallel group uses one AgentWorkspace with
      worktrees.
- [ ] Step results ingested from termination messages; missing result →
      synthesized failed + bounded retry.
- [ ] Effective gate policy = min(definition, AgentPolicy floor),
      webhook-enforced fail-closed.
- [ ] Agent image contains the pinned plugin set; size delta recorded.

## Files Modified/Created

- Created: `api/v1alpha1/workflowrun_types.go` + CRD yaml,
  `internal/controller/workflowrun_controller.go`,
  `internal/webhook/workflowrun_webhook.go`, envtest/e2e suites
- Modified: `agentpolicy_types.go` + webhook, `nix/agent-env.nix`,
  image-build task, `config/` kustomize wiring

## Dependencies

Requires 04 (pkg/workflow). Coordinate with
`plans/agent-security-hardening` (same operator; webhook deployment).
Parallel with 07 (disjoint files). Blocks 06.

## Estimated Duration

~1.5 weeks.
