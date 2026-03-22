---
type: plan
has_subplans: true
status: complete
completed_date: "2026-03-22"
dependencies:
  plans: []
  subplans:
    - plans/k8s-operator-security-hardening/01-shell-injection-fix.md
    - plans/k8s-operator-security-hardening/02-pod-security-context.md
    - plans/k8s-operator-security-hardening/03-network-policy.md
    - plans/k8s-operator-security-hardening/04-agent-policy-crd.md
    - plans/k8s-operator-security-hardening/05-audit-events.md
    - plans/k8s-operator-security-hardening/06-resource-hardening.md
    - plans/k8s-operator-security-hardening/07-confidential-computing.md
    - plans/k8s-operator-security-hardening/08-webhook-validation.md
parallel_groups:
  - group: 1
    parallel: true
    plans:
      - plans/k8s-operator-security-hardening/01-shell-injection-fix.md
      - plans/k8s-operator-security-hardening/02-pod-security-context.md
  - group: 2
    parallel: true
    depends_on: [1]
    plans:
      - plans/k8s-operator-security-hardening/03-network-policy.md
      - plans/k8s-operator-security-hardening/04-agent-policy-crd.md
  - group: 3
    parallel: true
    depends_on: [2]
    plans:
      - plans/k8s-operator-security-hardening/05-audit-events.md
      - plans/k8s-operator-security-hardening/06-resource-hardening.md
  - group: 4
    parallel: true
    depends_on: [3]
    plans:
      - plans/k8s-operator-security-hardening/07-confidential-computing.md
      - plans/k8s-operator-security-hardening/08-webhook-validation.md
skills_to_consult:
  - skill-general-fp
research_sources:
  documentation:
    - docs/gap-analysis-k8s-operator.md
  versions:
    go: "1.25+"
    controller-runtime: "current"
    kubernetes: "1.28+"
---

# K8s Operator Security Hardening

Close all 10 security gaps identified in `docs/gap-analysis-k8s-operator.md` for enterprise adoption (AKS, OWASP-SAMM L2, EU AI Act Article 4).

## Goals

- Fix 2 critical vulnerabilities (SecurityContext, shell injection)
- Add network isolation via NetworkPolicy generation
- Add central policy enforcement CRD with admission validation
- Add structured audit events for security-relevant actions
- Harden resource controls (ephemeral storage, TTL, EmptyDir limits)
- Add confidential computing support for AKS
- Complete webhook validation coverage

## Current State

- **Package**: `packages/agent/agent-operator-go/`
- **CRDs**: AgentRun, AgentHarness, AgentProvider, AgentWorkspace, AgentToolchain (v1alpha1)
- **Controllers**: agentrun_controller.go, agentprovider_controller.go, agentworkspace_controller.go
- **Builders**: job.go, container.go, workspace.go, pvc.go, env.go, toolchain_nix.go, harness*.go, workspace_vcs.go/git.go/jj.go
- **Webhooks**: agentrun, agentharness, agenttoolchain, agentworkspace (validated), agentprovider (empty)
- **Tests**: Unit (builders, resolvers, webhooks), integration (kubebuilder), E2E (kind + gVisor + Kata)
- **Runtime isolation**: Kata and gVisor via runtimeClassName, comprehensive E2E tests
- **Secret management**: K8s SecretKeyRef, never stored in CRDs
- **Note**: controller-gen broken with Go 1.25+ — CRDs and deepcopy must be maintained manually

## Research Findings

### Shell Injection (GAP 2)

Both `buildCloneScript` (`container.go:70-91`) and `buildWorkspaceCloneScript` (`workspace.go:168-188`) concatenate user-provided `repo.URL`, `repo.Branch`, and `repo.Commit` into shell scripts without sanitization. The `BuildWorktreeInitContainers` (`workspace.go:191-220`) also passes `worktreeBranch` and `sourceBranch` unsanitized.

**Approach**: Validate format in webhooks via regex (URL must match `^https?://` or `^git@`, branch/commit must match `^[a-zA-Z0-9._/-]+$`). Also add shell quoting in scripts as defense-in-depth.

### SecurityContext (GAP 1)

Neither `BuildJob` (`job.go`) nor `BuildWorkspaceJob` (`workspace.go`) set SecurityContext on containers or PodSecurityContext on the PodSpec. The operator's own `manager.yaml` has correct hardening but agent pods have none.

**Approach**: Add `SecurityContext` field to `AgentRunSpec` (optional override), with hardened defaults applied when nil. Add `PodSecurityContext` to PodSpec with `RunAsNonRoot: true`, `SeccompProfile: RuntimeDefault`.

### NetworkPolicy (GAP 3)

No NetworkPolicy resources are ever created. Agent pods have unrestricted network access.

**Approach**: Add `NetworkPolicySpec` to `AgentRunSpec` with `DenyAll` bool and optional `EgressRules`. Controller creates a NetworkPolicy owned by the AgentRun, targeting pods by label. Requires adding `networking.k8s.io` to RBAC.

### Policy CRD (GAP 4)

AgentHarness provides defaults but no enforcement. Any AgentRun can override image, resources, runtimeClassName, timeout.

**Approach**: New `AgentPolicy` CRD (namespace-scoped) with `enforced` (cannot override) and `defaults` (can override) sections. AgentRun webhook looks up the policy in the namespace and rejects violating specs. Covers: runtimeClassName, securityContext, maxTimeout, maxResources, requiredNetworkPolicy, allowedImages.

### Audit Events (GAP 6)

Controller uses `controller-runtime/pkg/log` for operational logging only.

**Approach**: Use `k8s.io/client-go/tools/record.EventRecorder` to emit K8s Events on: AgentRun creation (with resolved provider, runtime class), secret access, phase transitions, timeout, policy violations. Add structured log fields (agentRun, provider, runtimeClass) to controller logs.

## Proposed Approach

1. **Shell injection fix** — Add input validation to webhooks (AgentRun + AgentWorkspace) for URL/branch/commit format. Add shell quoting in `buildCloneScript` and `buildWorkspaceCloneScript` as defense-in-depth.

2. **Pod SecurityContext** — Add `SecurityContext` and `PodSecurityContext` fields to `AgentRunSpec`. Apply hardened defaults in `BuildJob` and `BuildWorkspaceJob`. Also apply to workspace init jobs and toolchain init containers.

3. **NetworkPolicy generation** — Add `NetworkPolicy` spec to `AgentRunSpec`. Controller creates NetworkPolicy owned by AgentRun. Add RBAC for `networking.k8s.io`. Default: deny-all egress when no rules specified, opt-in via explicit egress rules.

4. **AgentPolicy CRD** — New CRD with enforced constraints and overridable defaults. AgentRun webhook validates against namespace's AgentPolicy. New controller watches for policy changes.

5. **Audit events** — Add EventRecorder to controllers. Emit events on creation, secret access, phase transitions, timeout, policy violations. Add structured log fields.

6. **Resource hardening** — Add `SizeLimit` to Nix EmptyDir, add `ephemeral-storage` support, add `TTLSecondsAfterFinished` to JobSpec, propagate runtimeClassName to workspace init jobs.

7. **Confidential computing** — Add `ConfidentialComputing` field to AgentRunSpec (TEE type, attestation policy). Auto-add node affinity for AKS confidential compute labels.

8. **Webhook validation** — Complete AgentProvider webhook (validate authTokenSecretRef format, environment key safety, cliArgs injection). Add URL/branch format validation to AgentWorkspace webhook.

## Risk Assessment

- **CRD schema changes**: controller-gen is broken with Go 1.25+. CRD YAML must be updated manually. Risk of schema drift.
- **Breaking changes**: Adding `securityContext` defaults may break existing deployments where agents require root. Mitigate with opt-out field.
- **NetworkPolicy CNI dependency**: NetworkPolicy requires a CNI that supports it (Calico, Cilium). On AKS with Azure CNI, this works by default. Document requirement.
- **Policy CRD complexity**: Adding a 6th CRD increases operator surface area. Keep the policy CRD simple — enforced limits only, no complex rule engine.
- **E2E test updates**: All existing E2E tests (gVisor, Kata) need updating for new SecurityContext and potentially NetworkPolicy. Budget extra time.

## Proposed Child Plans

### Group 1 — Critical fixes (parallel)

- **`shell-injection-fix`**: Webhook input validation + shell quoting in clone scripts. Files: `agentrun_webhook.go`, `agentworkspace_webhook.go`, `container.go`, `workspace.go`, `workspace_git.go`, `workspace_jj.go`. Tests: webhook unit tests, builder unit tests.
- **`pod-security-context`**: SecurityContext on all containers and PodSecurityContext on PodSpec. Fields added to `agentrun_types.go`. Files: `agentrun_types.go`, `job.go`, `container.go`, `workspace.go`, CRD YAML. Tests: builder unit tests, E2E updates.

### Group 2 — Network and policy (parallel, after Group 1)

- **`network-policy`**: NetworkPolicy spec, controller creation, RBAC. Files: `agentrun_types.go`, `agentrun_controller.go`, new `builder/networkpolicy.go`, `config/rbac/role.yaml`, CRD YAML. Tests: builder unit tests, controller unit tests.
- **`agent-policy-crd`**: New AgentPolicy CRD, types, webhook enforcement, controller. Files: new `api/v1alpha1/agentpolicy_types.go`, new `internal/webhook/agentpolicy_webhook.go`, `agentrun_webhook.go` (policy lookup), new `internal/controller/agentpolicy_controller.go`, CRD YAML, RBAC. Tests: webhook tests, integration tests.

### Group 3 — Observability and hardening (parallel, after Group 2)

- **`audit-events`**: EventRecorder integration in controllers, structured logging. Files: `agentrun_controller.go`, `agentworkspace_controller.go`, `agentprovider_controller.go`, `cmd/operator/main.go`. Tests: controller unit tests.
- **`resource-hardening`**: EmptyDir SizeLimit, ephemeral-storage, TTL, workspace init runtimeClassName. Files: `toolchain_nix.go`, `job.go`, `workspace.go`, `agentrun_types.go`, CRD YAML. Tests: builder unit tests.

### Group 4 — Enterprise features (sequential, after Group 3)

- **`confidential-computing`**: ConfidentialComputing field, AKS node affinity, attestation config. Files: `agentrun_types.go`, `job.go`, `workspace.go`, CRD YAML. Tests: builder unit tests.
- **`webhook-validation`**: Complete AgentProvider webhook, add URL/branch validation to AgentWorkspace webhook. Files: `agentprovider_webhook.go`, `agentworkspace_webhook.go`. Tests: webhook unit tests.

## Success Criteria

- All 10 gaps from `docs/gap-analysis-k8s-operator.md` are closed
- All existing tests pass (unit, integration, E2E gVisor, E2E Kata)
- New tests cover: webhook validation rejects malicious input, SecurityContext is applied, NetworkPolicy is created, policy enforcement blocks violating AgentRuns, audit events are emitted
- CRD YAML files are updated manually (controller-gen limitation)
- `go test ./...` passes
- `golangci-lint run` passes

## Estimated Effort

- Group 1 (critical fixes): Small — 2 subplans, parallel
- Group 2 (network + policy): Medium — 2 subplans, parallel, AgentPolicy CRD is the largest piece
- Group 3 (audit + hardening): Small-Medium — 2 subplans, parallel
- Group 4 (enterprise): Small — 2 subplans, can be sequential
