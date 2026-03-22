---
type: plan
has_subplans: false
parent_plan: plans/k8s-operator-security-hardening.md
parallel_group: 3
status: complete
dependencies:
  plans:
    - plans/k8s-operator-security-hardening/03-network-policy.md
    - plans/k8s-operator-security-hardening/04-agent-policy-crd.md
  files:
    - packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go
    - packages/agent/agent-operator-go/internal/builder/job.go
    - packages/agent/agent-operator-go/internal/builder/toolchain_nix.go
    - packages/agent/agent-operator-go/internal/builder/workspace.go
    - packages/agent/agent-operator-go/config/crd/bases/agent.xonovex.com_agentruns.yaml
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pending
---

# Resource Hardening

## Objective

Four small, independent resource hardening fixes: (1) add `SizeLimit` to the Nix EmptyDir volume, (2) add `ephemeral-storage` to `ResourceRequirements`, (3) add `TTLSecondsAfterFinished` to Jobs, (4) propagate `runtimeClassName` to the workspace init job.

## Tasks

### 1. Add ephemeral storage to AgentRunSpec

In `packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go`, the `Resources` field is already `corev1.ResourceRequirements`. No type change needed — it already supports `ephemeral-storage` as a resource name. Document it in the sample.

Update `config/samples/agentrun_sample.yaml`:

```yaml
spec:
  resources:
    requests:
      cpu: "500m"
      memory: "512Mi"
      ephemeral-storage: "1Gi"
    limits:
      cpu: "2"
      memory: "2Gi"
      ephemeral-storage: "5Gi"
```

### 2. Add NixEmptyDirSizeLimit field to NixSpec

In `packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go`, update `NixSpec`:

```go
type NixSpec struct {
    // Packages are nixpkgs attribute names to install
    Packages []string `json:"packages,omitempty"`
    // Image is the Nix container image for the init container
    Image string `json:"image,omitempty"`
    // StoreSizeLimit is the size limit for the Nix store EmptyDir volume (default: "10Gi")
    StoreSizeLimit string `json:"storeSizeLimit,omitempty"`
}
```

In `packages/agent/agent-operator-go/internal/builder/toolchain_nix.go`, update `Volumes()`:

```go
import "k8s.io/apimachinery/pkg/api/resource"

func (n *NixToolchain) Volumes() []corev1.Volume {
    sizeLimit := resource.MustParse("10Gi")
    if n.nix.StoreSizeLimit != "" {
        sizeLimit = resource.MustParse(n.nix.StoreSizeLimit)
    }

    return []corev1.Volume{{
        Name: nixVolumeName,
        VolumeSource: corev1.VolumeSource{
            EmptyDir: &corev1.EmptyDirVolumeSource{
                SizeLimit: &sizeLimit,
            },
        },
    }}
}
```

### 3. Add TTLSecondsAfterFinished to AgentRunSpec

In `packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go`, add to `AgentRunSpec`:

```go
// TTLSecondsAfterFinished is the number of seconds the Job and its pods are retained
// after completion before automatic cleanup. Defaults to 3600 (1 hour). Set to 0 to
// clean up immediately.
TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
```

Also add to `AgentSpec` for harness-level defaults:

```go
DefaultTTLSecondsAfterFinished *int32 `json:"defaultTtlSecondsAfterFinished,omitempty"`
```

In `packages/agent/agent-operator-go/internal/builder/job.go`, apply the TTL in `BuildJob`. Pass the resolved TTL as a parameter:

```go
func BuildJob(run *agentv1alpha1.AgentRun, ..., ttl *int32) *batchv1.Job {
    // Resolve TTL: use spec value, fallback to default 3600
    var ttlSeconds *int32
    if ttl != nil {
        ttlSeconds = ttl
    } else {
        defaultTTL := int32(3600)
        ttlSeconds = &defaultTTL
    }

    job := &batchv1.Job{
        Spec: batchv1.JobSpec{
            ActiveDeadlineSeconds:   &activeDeadlineSeconds,
            BackoffLimit:            &backoffLimit,
            TTLSecondsAfterFinished: ttlSeconds,
            // ...
        },
    }
    return job
}
```

Apply the same to `BuildWorkspaceJob`. Update callers in `agentrun_controller.go` to pass the resolved TTL:

```go
// In resolver/defaults.go, add TTL resolution:
defaults.TTL = run.Spec.TTLSecondsAfterFinished
if defaults.TTL == nil && harness != nil {
    defaults.TTL = harness.Spec.DefaultTTLSecondsAfterFinished
}

// In controller, pass defaults.TTL to BuildJob
job := builder.BuildJob(agentRun, providerEnv, pvcName, defaults.Image, defaults.Timeout, agentType, wsType, tc, defaults.TTL)
```

### 4. Propagate runtimeClassName to workspace init job

In `packages/agent/agent-operator-go/internal/builder/workspace.go`, update `BuildWorkspaceInitJob` to accept an optional `runtimeClassName`:

```go
func BuildWorkspaceInitJob(ws *agentv1alpha1.AgentWorkspace, pvcName, image string, runtimeClassName *string) *batchv1.Job {
    // ...
    Spec: corev1.PodSpec{
        RestartPolicy:    corev1.RestartPolicyNever,
        RuntimeClassName: runtimeClassName,
        // ...
    },
}
```

In `packages/agent/agent-operator-go/internal/controller/agentworkspace_controller.go`, pass the workspace's runtimeClassName. The `AgentWorkspace` spec doesn't currently have a `RuntimeClassName` field — add it:

In `agentworkspace_types.go`, add to `AgentWorkspaceSpec`:

```go
// RuntimeClassName sets the pod runtimeClassName for the workspace init Job.
// Use to ensure the clone container runs inside the same isolation as agent pods.
RuntimeClassName *string `json:"runtimeClassName,omitempty"`
```

In the workspace controller:

```go
job := builder.BuildWorkspaceInitJob(&ws, workspacePVCName, image, ws.Spec.RuntimeClassName)
```

Update the E2E gVisor test that explicitly asserts `workspace init Job correctly has NO runtimeClassName` — this assertion must be changed to `workspace init Job has runtimeClassName if set on AgentWorkspace`.

### 5. Update resolver defaults for TTL

In `packages/agent/agent-operator-go/internal/resolver/defaults.go`, add TTL field to `ResolvedDefaults`:

```go
type ResolvedDefaults struct {
    Image   string
    Timeout time.Duration
    TTL     *int32
}
```

### 6. Manually update CRD YAML

In `config/crd/bases/agent.xonovex.com_agentruns.yaml`, add to `spec.properties`:

```yaml
ttlSecondsAfterFinished:
  description: Seconds to retain completed Job before cleanup. Defaults to 3600.
  type: integer
  format: int32
  minimum: 0
```

In `config/crd/bases/agent.xonovex.com_agentworkspaces.yaml`, add to `spec.properties`:

```yaml
runtimeClassName:
  description: RuntimeClassName for the workspace init Job pod.
  type: string
```

In `config/crd/bases/agent.xonovex.com_agenttoolchains.yaml` (if AgentToolchain embeds NixSpec) and the AgentRun CRD's toolchain inline, add `storeSizeLimit`.

### 7. Update unit tests

In `packages/agent/agent-operator-go/internal/builder/`:

- `container_test.go` — `TestNixToolchain_Volumes_DefaultSizeLimit` → EmptyDir SizeLimit is `10Gi`
- `container_test.go` — `TestNixToolchain_Volumes_CustomSizeLimit` → custom value respected
- `job_test.go` — `TestBuildJob_DefaultTTL` → TTL=3600 when nil passed
- `job_test.go` — `TestBuildJob_ExplicitTTL` → provided TTL value used
- `workspace_test.go` — `TestBuildWorkspaceInitJob_WithRuntimeClass` → runtimeClassName propagated

Update E2E gVisor test to match new workspace init job behaviour.

## Validation Steps

```bash
cd packages/agent/agent-operator-go
go build ./...
go vet ./...
go test ./internal/builder/...
go test ./internal/resolver/...
go test ./internal/controller/...
golangci-lint run ./...
```

## Success Criteria

- [ ] Nix EmptyDir has `SizeLimit: 10Gi` by default; configurable via `spec.toolchain.nix.storeSizeLimit`
- [ ] `TTLSecondsAfterFinished` defaults to 3600 on all Jobs
- [ ] Workspace init job accepts optional `runtimeClassName`
- [ ] `AgentWorkspaceSpec` has optional `runtimeClassName` field
- [ ] CRD YAML files manually updated for all new fields
- [ ] All new unit tests pass
- [ ] E2E gVisor workspace test updated

## Files Modified/Created

- `api/v1alpha1/agentrun_types.go` — `NixSpec.StoreSizeLimit`, `AgentRunSpec.TTLSecondsAfterFinished`, `AgentSpec.DefaultTTLSecondsAfterFinished`
- `api/v1alpha1/agentworkspace_types.go` — `AgentWorkspaceSpec.RuntimeClassName`
- `api/v1alpha1/zz_generated.deepcopy.go` — manual updates for new pointer fields
- `internal/builder/toolchain_nix.go` — EmptyDir SizeLimit
- `internal/builder/job.go` — TTL parameter
- `internal/builder/workspace.go` — runtimeClassName parameter in `BuildWorkspaceInitJob`
- `internal/controller/agentrun_controller.go` — pass TTL to builder
- `internal/controller/agentworkspace_controller.go` — pass runtimeClassName to builder
- `internal/resolver/defaults.go` — TTL field in ResolvedDefaults
- `config/crd/bases/agent.xonovex.com_agentruns.yaml` — TTL field
- `config/crd/bases/agent.xonovex.com_agentworkspaces.yaml` — runtimeClassName field
- `config/samples/agentrun_sample.yaml` — ephemeral-storage example

## Estimated Duration

Small — ~2h
