---
type: plan
has_subplans: false
parent_plan: ../symmetric-workflow-commands.md
parallel_group: 1
status: complete
updated: 2026-07-20
completed_date: "2026-07-20"
dependencies:
  plans: []
  files: []
skills_to_consult:
  - kubernetes-guide
  - orthogonal-pattern-guide
  - testing-guide
  - git-guide
validation:
  type_check: passed
  lint: passed
  build: passed
  tests: passed
  integration: passed
---

# Subplan 04: External trigger boundary

## Objective

Remove schedules, event ingress, and trigger interpretation from the base agent
operator. Leave `AgentRun` as the direct execution request API so manual tools,
harness hooks, CI/CD systems, webhooks, Kubernetes CronJobs, and other external
integrations can create runs without an operator-owned trigger framework.

Preserve agent execution, workspace, provider, harness, toolchain, admission,
and sandbox-security behavior. This is a breaking API removal, not a security
policy reduction or compatibility migration.

## Tasks

### 1. Remove the trigger and schedule API types

- Delete `api/v1alpha1/agenttrigger_types.go`, including `AgentSchedule`,
  `AgentTrigger`, their specs/status/lists, concurrency policy, token selector,
  and type registration.
- Regenerate or precisely update `api/v1alpha1/zz_generated.deepcopy.go` to
  remove only generated trigger/schedule functions (baseline starts near line
  481).
- Verify `groupversion_info.go` and test schemes still register every retained
  API type and contain no trigger-specific assumptions.
- Do not add replacement generic-trigger fields to `AgentRun`.

### 2. Remove trigger/schedule controllers and tests

- Delete `internal/controller/agentschedule_controller.go` and
  `agentschedule_controller_test.go`.
- Delete `internal/controller/agenttrigger_controller.go` and
  `agenttrigger_controller_test.go`, including the HTTP receiver,
  authentication, idempotency, readiness, and `buildTriggeredRun` helpers.
- Search retained controller and test utilities for annotations/constants used
  only by generated trigger runs; remove them only when now unused.
- Preserve all `AgentRun`, provider, workspace, isolation, network, resolver,
  validator, and webhook tests.

### 3. Reduce operator startup to retained reconcilers

- In `cmd/operator/main.go:29-112`, remove `triggerBindAddress`, the
  `--trigger-bind-address` flag, `AgentScheduleReconciler`,
  `AgentTriggerReconciler`, and `AgentTriggerReceiver` registration.
- Keep manager, health/readiness, `AgentRun`, provider, workspace, and admission
  setup unchanged except for imports made unused by removal.
- Verify there is no additional listener or service port dedicated to triggering.

The retained control path should be conceptually:

```text
external caller -> Kubernetes API -> AgentRun -> AgentRunReconciler -> secured workload
```

### 4. Remove Kubernetes installation surfaces

- Delete the `agentschedules` and `agenttriggers` CRD YAML files.
- Remove their entries from `config/crd/kustomization.yaml`.
- Delete `config/manager/trigger-service.yaml` and
  `trigger-network-policy.yaml`, and remove them from manager/default
  kustomizations.
- Remove schedule/trigger verbs and resources from `config/rbac/role.yaml` and
  any generated webhook/manager manifests.
- Remove the trigger receiver argument/port from
  `config/manager/manager.yaml:31` and related container/service declarations.
- Build all kustomizations to prove no deleted resource remains referenced.

### 5. Remove the cron dependency and reconcile generated metadata

- Remove `github.com/robfig/cron/v3` from `go.mod` and `go.sum` once the schedule
  controller is gone.
- Run `go mod tidy` from the operator package.
- Run the repository's Go formatting, generation, build, lint, unit, and
  integration paths; do not invent a new generator task if the workspace uses
  direct tooling.
- Confirm no retained dependency exists solely for the deleted receiver.

### 6. Align operator and agent documentation

- Rewrite trigger/workflow sections in `agent-operator-go/README.md` and the
  local `AGENTS.md`/`CLAUDE.md` mirrors to describe direct `AgentRun` submission.
- Update `architecture.dot` and regenerate `architecture.png` without the
  trigger/schedule ingress components.
- Remove the stale `Governance & oversight` claims at
  `packages/agent/AGENTS.md:38-40` and any assertions about accountable owners,
  AIBOM journals, anomaly stops, or workflow enforcement that no retained code
  implements.
- Preserve and accurately name `AgentPolicy` and shared sandbox controls as
  execution security.
- Correct any adjacent stale toolchain resolver wording only where verified by
  current declarations; do not broaden this into an unrelated operator refactor.

### 7. Prove direct execution and retained security still work

- Keep or add a focused integration fixture that creates an `AgentRun` directly
  and reaches the existing reconciliation path without a trigger CRD.
- Run retained webhook, admission, workspace, isolation, network, and policy
  tests unchanged where possible.
- Search tracked agent/shared files for `AgentTrigger`, `AgentSchedule`,
  trigger receiver addresses/services, and cron usage.
- Verify `packages/shared/shared-agent-go` has no functional changes and its
  policy tests still pass.

## Validation steps

1. `go mod tidy` in `packages/agent/agent-operator-go`.
2. `npx moon run agent-operator-go:ci-check`.
3. `npx moon run shared-agent-go:ci-check`.
4. Build each retained operator kustomization with the repository's configured
   Kubernetes tooling.
5. `rg -n 'AgentTrigger|AgentSchedule|trigger-bind-address|trigger-service|robfig/cron' packages/agent packages/shared/shared-agent-go` must return no active-code/config references.
6. Inspect the regenerated `architecture.png`.
7. `git diff --check`.

## Success criteria

- [x] Trigger/schedule Go types, controllers, receiver, tests, CRDs, RBAC,
      service, network policy, arguments, and dependencies are gone.
- [x] The operator exposes no trigger-specific network listener.
- [x] A direct `AgentRun` remains the complete execution entry point.
- [x] Retained controller, admission, integration, and security tests pass.
- [x] `AgentPolicy` and shared sandbox policy are unchanged except for accurate
      documentation.
- [x] Agent documentation makes trigger ownership external and contains no
      stale lifecycle-governance claims.
- [x] No replacement trigger abstraction or compatibility API is introduced.

## Files modified/created

- Delete: `api/v1alpha1/agenttrigger_types.go`.
- Modify: `api/v1alpha1/zz_generated.deepcopy.go` and scheme metadata as needed.
- Create: `api/v1alpha1/groupversion_info_test.go`.
- Delete: `internal/controller/agentschedule_controller.go` and its test.
- Delete: `internal/controller/agenttrigger_controller.go` and its test.
- Modify: `cmd/operator/main.go`.
- Delete: trigger/schedule CRDs and manager trigger service/network policy.
- Modify: CRD, manager, default, RBAC, and generated manifest kustomizations.
- Modify: `go.mod` and `go.sum`.
- Modify: operator README/instructions and `architecture.dot/.png`.
- Modify: `packages/agent/AGENTS.md` and its generated/instruction mirrors as
  required by the instruction synchronization rules.
- Verify only: `packages/shared/shared-agent-go/**`.

## Dependencies

- No child-plan dependency.
- Can run in parallel with subplan 01 because their files do not overlap.
- Must finish before subplan 03 publishes Kubernetes invocation examples and
  before subplan 05 runs final residue/release validation.

## Validation Results

- `npx moon run agent-operator-go:ci-check` passes formatting, lint, typecheck,
  build, unit/coverage, and envtest integration tasks. The retained API scheme
  and runtime deep-copy test keeps aggregate coverage at 45.9% against the 35%
  gate.
- `npx moon run shared-agent-go:ci-check` passes with no functional diff in the
  shared policy package.
- `kubectl kustomize` builds all six retained roots: cert-manager, CRDs,
  default, manager, RBAC, and webhooks.
- The trigger/schedule residue scan returns no matches, the manager exposes
  only health and webhook ports, and `git diff --check` passes.
- `architecture.png` was regenerated from `architecture.dot` at 150 DPI and
  visually inspected for readable direct-submission labels.

## Estimated duration

One to two focused implementation sessions because Kubernetes API removal must
stay synchronized across Go, generated code, manifests, tests, and diagrams.
