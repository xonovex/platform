---
type: plan
has_subplans: false
parent_plan: plans/k8s-operator-security-hardening.md
parallel_group: 4
status: complete
dependencies:
  plans:
    - plans/k8s-operator-security-hardening/05-audit-events.md
    - plans/k8s-operator-security-hardening/06-resource-hardening.md
  files:
    - packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go
    - packages/agent/agent-operator-go/internal/builder/job.go
    - packages/agent/agent-operator-go/internal/builder/workspace.go
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

# Confidential Computing Support (AKS)

## Objective

Add a `confidentialComputing` field to `AgentRunSpec` that, when set, automatically adds the correct node affinity for AKS confidential compute node pools and documents the supported TEE types (AMD SEV-SNP via `kata-cc`, Intel TDX). This does not implement attestation — it provides the K8s scheduling primitives needed to land agent pods on confidential compute nodes.

## Background

Azure Kubernetes Service supports confidential computing via:
- **AMD SEV-SNP**: `kata-cc` runtimeClass + nodes labelled `kubernetes.azure.com/confidential-computing: true`
- **Intel TDX**: `kata-tdx` runtimeClass on specific node SKUs

The operator already supports arbitrary `runtimeClassName`. What's missing is:
1. Automatic node affinity for confidential compute node pools
2. A dedicated field so users don't need to know the exact AKS node labels
3. Documentation of the recommended runtimeClassName values

## Tasks

### 1. Add ConfidentialComputing spec to AgentRunSpec

In `packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go`:

```go
// TEEType identifies the Trusted Execution Environment type.
type TEEType string

const (
    // TEETypeAMDSEVSNP uses AMD SEV-SNP via the kata-cc runtimeClass on AKS.
    TEETypeAMDSEVSNP TEEType = "amd-sev-snp"
    // TEETypeIntelTDX uses Intel TDX via the kata-tdx runtimeClass on AKS.
    TEETypeIntelTDX TEEType = "intel-tdx"
)

// ConfidentialComputingSpec configures TEE-based isolation for agent pods on AKS.
type ConfidentialComputingSpec struct {
    // TEE is the Trusted Execution Environment type.
    TEE TEEType `json:"tee"`

    // OverrideRuntimeClassName overrides the automatically selected runtimeClassName
    // for the TEE type. Leave empty to use the default for the TEE.
    OverrideRuntimeClassName *string `json:"overrideRuntimeClassName,omitempty"`

    // DisableNodeAffinity skips adding the AKS confidential compute node affinity.
    // Use only if you manage node selection via NodeSelector or Tolerations directly.
    DisableNodeAffinity bool `json:"disableNodeAffinity,omitempty"`
}

// AgentRunSpec — add field:
type AgentRunSpec struct {
    // ... existing fields ...

    // ConfidentialComputing configures TEE-based isolation for agent pods.
    // When set, the operator adds AKS confidential compute node affinity and
    // sets the appropriate runtimeClassName for the TEE type.
    // Takes precedence over RuntimeClassName if both are set.
    ConfidentialComputing *ConfidentialComputingSpec `json:"confidentialComputing,omitempty"`
}
```

Also add to `AgentSpec` for harness-level defaults:

```go
type AgentSpec struct {
    // ... existing fields ...
    DefaultConfidentialComputing *ConfidentialComputingSpec `json:"defaultConfidentialComputing,omitempty"`
}
```

### 2. Add teeRuntimeClassName helper in builder

Create `packages/agent/agent-operator-go/internal/builder/confidential.go`:

```go
package builder

import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// AKS node label for confidential compute nodes.
const aksConfidentialNodeLabel = "kubernetes.azure.com/confidential-computing"

// Default runtimeClassNames per TEE type on AKS.
var teeRuntimeClassNames = map[agentv1alpha1.TEEType]string{
    agentv1alpha1.TEETypeAMDSEVSNP: "kata-cc",
    agentv1alpha1.TEETypeIntelTDX:  "kata-tdx",
}

// TEERuntimeClassName returns the runtimeClassName for a TEE configuration.
func TEERuntimeClassName(cc *agentv1alpha1.ConfidentialComputingSpec) *string {
    if cc == nil {
        return nil
    }
    if cc.OverrideRuntimeClassName != nil {
        return cc.OverrideRuntimeClassName
    }
    name := teeRuntimeClassNames[cc.TEE]
    if name == "" {
        return nil
    }
    return &name
}

// TEENodeAffinity returns the node affinity for a TEE configuration.
// Returns nil if DisableNodeAffinity is true or cc is nil.
func TEENodeAffinity(cc *agentv1alpha1.ConfidentialComputingSpec) *corev1.NodeAffinity {
    if cc == nil || cc.DisableNodeAffinity {
        return nil
    }

    return &corev1.NodeAffinity{
        RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
            NodeSelectorTerms: []corev1.NodeSelectorTerm{{
                MatchExpressions: []corev1.NodeSelectorRequirement{{
                    Key:      aksConfidentialNodeLabel,
                    Operator: corev1.NodeSelectorOpIn,
                    Values:   []string{"true"},
                }},
            }},
        },
    }
}
```

Also create `packages/agent/agent-operator-go/internal/builder/confidential_test.go`:
- `TestTEERuntimeClassName_AMDSEVSNP` → returns `kata-cc`
- `TestTEERuntimeClassName_Override` → returns override value
- `TestTEERuntimeClassName_Nil` → returns nil
- `TestTEENodeAffinity_AMDSEVSNP` → returns NodeAffinity with AKS label
- `TestTEENodeAffinity_DisableNodeAffinity` → returns nil
- `TestTEENodeAffinity_Nil` → returns nil

### 3. Apply ConfidentialComputing in BuildJob and BuildWorkspaceJob

In `packages/agent/agent-operator-go/internal/builder/job.go`, apply TEE runtime class and node affinity. The CC spec takes precedence over `RuntimeClassName` if both are set:

```go
func BuildJob(run *agentv1alpha1.AgentRun, ...) *batchv1.Job {
    cc := run.Spec.ConfidentialComputing

    // TEE runtimeClassName takes precedence over direct RuntimeClassName
    runtimeClass := run.Spec.RuntimeClassName
    if teeRC := TEERuntimeClassName(cc); teeRC != nil {
        runtimeClass = teeRC
    }

    // Build node affinity from TEE config + existing affinity
    var affinity *corev1.Affinity
    if nodeAffinity := TEENodeAffinity(cc); nodeAffinity != nil {
        affinity = &corev1.Affinity{NodeAffinity: nodeAffinity}
    }

    job := &batchv1.Job{
        Spec: batchv1.JobSpec{
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    // ...
                    RuntimeClassName: runtimeClass,
                    Affinity:         affinity,
                    // ...
                },
            },
        },
    }
    return job
}
```

Apply the same pattern to `BuildWorkspaceJob`.

### 4. Propagate from harness resolver

In `packages/agent/agent-operator-go/internal/resolver/defaults.go`, add CC to ResolvedDefaults:

```go
type ResolvedDefaults struct {
    // ... existing ...
    ConfidentialComputing *agentv1alpha1.ConfidentialComputingSpec
}

func ApplyHarnessDefaults(run *agentv1alpha1.AgentRun, harness *agentv1alpha1.AgentHarness) ResolvedDefaults {
    // ...
    defaults.ConfidentialComputing = run.Spec.ConfidentialComputing
    if defaults.ConfidentialComputing == nil && harness != nil {
        defaults.ConfidentialComputing = harness.Spec.DefaultConfidentialComputing
    }
    return defaults
}
```

Pass `defaults.ConfidentialComputing` to `BuildJob` — fold it into the AgentRun spec before calling the builder, or pass as parameter.

### 5. Add validation to AgentRun webhook

In `agentrun_webhook.go`, add validation that the TEE type is a known value if `confidentialComputing` is set:

```go
if run.Spec.ConfidentialComputing != nil {
    cc := run.Spec.ConfidentialComputing
    switch cc.TEE {
    case agentv1alpha1.TEETypeAMDSEVSNP, agentv1alpha1.TEETypeIntelTDX:
        // valid
    default:
        return nil, fmt.Errorf("unknown TEE type %q", cc.TEE)
    }
    // Warn if RuntimeClassName is also set (CC takes precedence)
    if run.Spec.RuntimeClassName != nil {
        return admission.Warnings{"confidentialComputing.tee takes precedence over runtimeClassName"}, nil
    }
}
```

### 6. Manually update CRD YAML

In `config/crd/bases/agent.xonovex.com_agentruns.yaml`, add to `spec.properties`:

```yaml
confidentialComputing:
  description: Configures TEE-based isolation for agent pods on AKS.
  type: object
  required: [tee]
  properties:
    tee:
      type: string
      enum: [amd-sev-snp, intel-tdx]
    overrideRuntimeClassName:
      type: string
    disableNodeAffinity:
      type: boolean
```

### 7. Add sample

Add to `config/samples/agentrun_sample.yaml` (as a comment showing the AKS pattern):

```yaml
# For AKS Confidential Computing (AMD SEV-SNP):
# spec:
#   confidentialComputing:
#     tee: amd-sev-snp
#   resources:
#     requests:
#       cpu: "2"
#       memory: "4Gi"
```

## Validation Steps

```bash
cd packages/agent/agent-operator-go
go build ./...
go vet ./...
go test ./internal/builder/...
go test ./internal/resolver/...
go test ./internal/webhook/...
golangci-lint run ./...
```

## Success Criteria

- [ ] `ConfidentialComputingSpec` type with `TEEType` enum defined
- [ ] `TEERuntimeClassName` and `TEENodeAffinity` helpers implemented and unit tested
- [ ] `BuildJob` and `BuildWorkspaceJob` apply TEE runtimeClass and node affinity
- [ ] CC takes precedence over direct `RuntimeClassName`
- [ ] Webhook validates TEE type; warns when both `runtimeClassName` and CC are set
- [ ] CRD YAML updated with CC field and TEE enum
- [ ] Sample YAML documents AKS usage pattern
- [ ] All unit tests pass

## Files Modified/Created

- `api/v1alpha1/agentrun_types.go` — `TEEType`, `ConfidentialComputingSpec`, field in `AgentRunSpec`/`AgentSpec`
- `api/v1alpha1/zz_generated.deepcopy.go` — manual DeepCopyInto for new types
- `internal/builder/confidential.go` (new)
- `internal/builder/confidential_test.go` (new)
- `internal/builder/job.go` — apply TEE runtime class and affinity
- `internal/builder/workspace.go` — apply TEE runtime class and affinity to workspace job
- `internal/resolver/defaults.go` — CC in ResolvedDefaults
- `internal/webhook/agentrun_webhook.go` — TEE type validation
- `config/crd/bases/agent.xonovex.com_agentruns.yaml` — CC field
- `config/samples/agentrun_sample.yaml` — AKS CC example

## Estimated Duration

Medium — ~3h
