---
type: plan
has_subplans: false
parent_plan: plans/k8s-operator-security-hardening.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go
    - packages/agent/agent-operator-go/internal/builder/job.go
    - packages/agent/agent-operator-go/internal/builder/container.go
    - packages/agent/agent-operator-go/internal/builder/workspace.go
    - packages/agent/agent-operator-go/internal/builder/toolchain_nix.go
    - packages/agent/agent-operator-go/config/crd/bases/agent.xonovex.com_agentruns.yaml
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Pod SecurityContext

## Objective

Add hardened `SecurityContext` (container-level) and `PodSecurityContext` (pod-level) to all agent Jobs and workspace init Jobs. Apply secure defaults while allowing overrides via `AgentRunSpec`. Match the existing hardening already present on the operator's own manager pod.

## Tasks

### 1. Add SecurityContext fields to AgentRunSpec

In `packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go`, add to `AgentRunSpec`:

```go
import corev1 "k8s.io/api/core/v1"

// AgentRunSpec defines the desired state of AgentRun
type AgentRunSpec struct {
    // ... existing fields ...

    // SecurityContext overrides the default container security context.
    // Defaults apply runAsNonRoot, drop all capabilities, and RuntimeDefault seccomp.
    SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`

    // PodSecurityContext overrides the default pod-level security context.
    PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
}
```

Also add to `AgentSpec` (used by AgentHarness) so defaults can be set at the harness level:

```go
type AgentSpec struct {
    // ... existing fields ...
    DefaultSecurityContext    *corev1.SecurityContext    `json:"defaultSecurityContext,omitempty"`
    DefaultPodSecurityContext *corev1.PodSecurityContext `json:"defaultPodSecurityContext,omitempty"`
}
```

**Note**: controller-gen is broken with Go 1.25+. Update CRD YAML manually (see Task 6).

### 2. Add defaultSecurityContexts helpers in builder package

Create `packages/agent/agent-operator-go/internal/builder/security.go`:

```go
package builder

import corev1 "k8s.io/api/core/v1"

// DefaultContainerSecurityContext returns hardened defaults matching the operator manager pod.
// Callers may override by passing a non-nil override which takes precedence field-by-field.
func DefaultContainerSecurityContext(override *corev1.SecurityContext) *corev1.SecurityContext {
    allowPrivEsc := false
    runAsNonRoot := true
    readOnlyRoot := true
    seccomp := corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault}
    dropAll := []corev1.Capability{"ALL"}

    sc := &corev1.SecurityContext{
        AllowPrivilegeEscalation: &allowPrivEsc,
        RunAsNonRoot:             &runAsNonRoot,
        ReadOnlyRootFilesystem:   &readOnlyRoot,
        SeccompProfile:           &seccomp,
        Capabilities:             &corev1.Capabilities{Drop: dropAll},
    }

    if override != nil {
        if override.AllowPrivilegeEscalation != nil {
            sc.AllowPrivilegeEscalation = override.AllowPrivilegeEscalation
        }
        if override.RunAsNonRoot != nil {
            sc.RunAsNonRoot = override.RunAsNonRoot
        }
        if override.ReadOnlyRootFilesystem != nil {
            sc.ReadOnlyRootFilesystem = override.ReadOnlyRootFilesystem
        }
        if override.SeccompProfile != nil {
            sc.SeccompProfile = override.SeccompProfile
        }
        if override.Capabilities != nil {
            sc.Capabilities = override.Capabilities
        }
        if override.RunAsUser != nil {
            sc.RunAsUser = override.RunAsUser
        }
        if override.RunAsGroup != nil {
            sc.RunAsGroup = override.RunAsGroup
        }
    }

    return sc
}

// DefaultPodSecurityContext returns hardened pod-level defaults.
func DefaultPodSecurityContext(override *corev1.PodSecurityContext) *corev1.PodSecurityContext {
    runAsNonRoot := true
    seccomp := corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault}

    psc := &corev1.PodSecurityContext{
        RunAsNonRoot:   &runAsNonRoot,
        SeccompProfile: &seccomp,
    }

    if override != nil {
        if override.RunAsNonRoot != nil {
            psc.RunAsNonRoot = override.RunAsNonRoot
        }
        if override.RunAsUser != nil {
            psc.RunAsUser = override.RunAsUser
        }
        if override.RunAsGroup != nil {
            psc.RunAsGroup = override.RunAsGroup
        }
        if override.FSGroup != nil {
            psc.FSGroup = override.FSGroup
        }
        if override.SeccompProfile != nil {
            psc.SeccompProfile = override.SeccompProfile
        }
        if override.Sysctls != nil {
            psc.Sysctls = override.Sysctls
        }
    }

    return psc
}
```

### 3. Apply SecurityContext in BuildJob and BuildWorkspaceJob

In `packages/agent/agent-operator-go/internal/builder/job.go`, update `BuildJob`:

```go
func BuildJob(run *agentv1alpha1.AgentRun, ...) *batchv1.Job {
    // ... existing volume setup ...

    job := &batchv1.Job{
        // ... existing ObjectMeta ...
        Spec: batchv1.JobSpec{
            // ... existing deadline/backoff ...
            Template: corev1.PodTemplateSpec{
                // ... existing ObjectMeta ...
                Spec: corev1.PodSpec{
                    RestartPolicy:       corev1.RestartPolicyNever,
                    SecurityContext:     DefaultPodSecurityContext(run.Spec.PodSecurityContext),
                    InitContainers:      BuildInitContainers(run, image, wsType, tc, run.Spec.SecurityContext),
                    Containers:          BuildMainContainers(run, providerEnv, image, agentType, tc, run.Spec.SecurityContext),
                    Volumes:             volumes,
                    NodeSelector:        run.Spec.NodeSelector,
                    Tolerations:         run.Spec.Tolerations,
                    RuntimeClassName:    run.Spec.RuntimeClassName,
                },
            },
        },
    }
    // ... existing resource assignment ...
    return job
}
```

Apply the same pattern to `BuildWorkspaceJob` in `workspace.go`.

### 4. Thread SecurityContext through container builders

Update `BuildMainContainers` and `BuildInitContainers` signatures in `container.go` to accept and apply `*corev1.SecurityContext`:

```go
func BuildMainContainers(run *agentv1alpha1.AgentRun, providerEnv map[string]string, image string, agentType agentv1alpha1.AgentType, tc *agentv1alpha1.ToolchainSpec, sc *corev1.SecurityContext) []corev1.Container {
    // ...
    return []corev1.Container{{
        Name:            "agent",
        Image:           image,
        Command:         command,
        Args:            args,
        Env:             env,
        WorkingDir:      workspaceMountPath,
        VolumeMounts:    volumeMounts,
        SecurityContext: DefaultContainerSecurityContext(sc),
    }}
}
```

Also add a writable `/tmp` EmptyDir volume since `ReadOnlyRootFilesystem: true` prevents writing to the root FS:

```go
// In BuildJob volumes setup:
volumes = append(volumes, corev1.Volume{
    Name: "tmp",
    VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
})

// In BuildMainContainers volumeMounts:
volumeMounts = append(volumeMounts, corev1.VolumeMount{
    Name:      "tmp",
    MountPath: "/tmp",
})
```

Apply the same SecurityContext to all init containers (git-clone, nix-env).

### 5. Apply SecurityContext to workspace init job

In `packages/agent/agent-operator-go/internal/builder/workspace.go`, update `BuildWorkspaceInitJob` to apply hardened defaults (no override field needed — workspace init job uses defaults only):

```go
func BuildWorkspaceInitJob(ws *agentv1alpha1.AgentWorkspace, pvcName, image string) *batchv1.Job {
    // ...
    Spec: corev1.PodSpec{
        RestartPolicy:   corev1.RestartPolicyNever,
        SecurityContext: DefaultPodSecurityContext(nil),
        Containers: []corev1.Container{{
            Name:            "git-clone",
            // ...
            SecurityContext: DefaultContainerSecurityContext(nil),
        }},
        // ...
    },
}
```

### 6. Update AgentSpec resolver defaults

In `packages/agent/agent-operator-go/internal/resolver/defaults.go`, propagate `DefaultSecurityContext` and `DefaultPodSecurityContext` from AgentHarness to AgentRun when not already set:

```go
func ApplyHarnessDefaults(run *agentv1alpha1.AgentRun, harness *agentv1alpha1.AgentHarness) ResolvedDefaults {
    // ...
    if run.Spec.SecurityContext == nil && harness != nil {
        defaults.SecurityContext = harness.Spec.DefaultSecurityContext
    }
    if run.Spec.PodSecurityContext == nil && harness != nil {
        defaults.PodSecurityContext = harness.Spec.DefaultPodSecurityContext
    }
    // ...
}
```

### 7. Manually update CRD YAML

Since controller-gen is broken with Go 1.25+, manually add the new fields to `config/crd/bases/agent.xonovex.com_agentruns.yaml` under `spec.properties`:

```yaml
securityContext:
  description: SecurityContext overrides the default container security context.
  type: object
  x-kubernetes-preserve-unknown-fields: true
podSecurityContext:
  description: PodSecurityContext overrides the default pod-level security context.
  type: object
  x-kubernetes-preserve-unknown-fields: true
```

Use `x-kubernetes-preserve-unknown-fields: true` for the nested SecurityContext/PodSecurityContext since their schemas are complex Kubernetes built-in types. Same additions needed in `agent.xonovex.com_agentharnesses.yaml`.

### 8. Update unit tests

In `packages/agent/agent-operator-go/internal/builder/`:

- `container_test.go` — update all `BuildMainContainers` calls to pass `nil` security context; add `TestBuildMainContainers_SecurityContextDefaults` verifying `AllowPrivilegeEscalation=false`, `RunAsNonRoot=true`, `ReadOnlyRootFilesystem=true`, `Capabilities.Drop=["ALL"]`; add `TestBuildMainContainers_SecurityContextOverride`.
- `job_test.go` — add `TestBuildJob_PodSecurityContext`.

In `packages/agent/agent-operator-go/internal/builder/security_test.go` (new) — test `DefaultContainerSecurityContext` with nil and non-nil overrides.

## Validation Steps

```bash
cd packages/agent/agent-operator-go
go build ./...
go vet ./...
go test ./internal/builder/...
go test ./internal/resolver/...
golangci-lint run ./...
```

Verify existing E2E tests still pass (SecurityContext is additive; gVisor/Kata runtime classes still propagate correctly).

## Success Criteria

- [ ] `DefaultContainerSecurityContext` and `DefaultPodSecurityContext` helpers exist in `builder/security.go`
- [ ] All agent containers have `AllowPrivilegeEscalation: false`, `RunAsNonRoot: true`, `ReadOnlyRootFilesystem: true`, `Capabilities.Drop: [ALL]`, `SeccompProfile: RuntimeDefault`
- [ ] All init containers (git-clone, nix-env, git-worktree) have the same SecurityContext
- [ ] Workspace init job has SecurityContext applied
- [ ] `/tmp` EmptyDir added to allow agent to write temporary files
- [ ] `AgentRunSpec` and `AgentSpec` have optional SecurityContext override fields
- [ ] CRD YAML files manually updated for new fields
- [ ] All existing unit tests pass with updated function signatures
- [ ] New SecurityContext unit tests pass

## Files Modified/Created

- `api/v1alpha1/agentrun_types.go` — add `SecurityContext`, `PodSecurityContext` to `AgentRunSpec` and `AgentSpec`
- `api/v1alpha1/zz_generated.deepcopy.go` — manually add DeepCopyInto for new pointer fields
- `internal/builder/security.go` (new)
- `internal/builder/security_test.go` (new)
- `internal/builder/job.go` — apply security contexts
- `internal/builder/container.go` — update signatures, apply SecurityContext, add tmp volume
- `internal/builder/workspace.go` — apply security contexts to all job/container builders
- `internal/resolver/defaults.go` — propagate SecurityContext from harness
- `config/crd/bases/agent.xonovex.com_agentruns.yaml` — manual update
- `config/crd/bases/agent.xonovex.com_agentharnesses.yaml` — manual update
- `internal/builder/container_test.go` — update signatures + new security tests
- `internal/builder/job_test.go` — new security tests

## Estimated Duration

Small-Medium — ~3h
