---
type: plan
has_subplans: false
parent_plan: plans/k8s-operator-security-hardening.md
parallel_group: 2
status: complete
dependencies:
  plans:
    - plans/k8s-operator-security-hardening/01-shell-injection-fix.md
    - plans/k8s-operator-security-hardening/02-pod-security-context.md
  files:
    - packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go
    - packages/agent/agent-operator-go/internal/controller/agentrun_controller.go
    - packages/agent/agent-operator-go/config/rbac/role.yaml
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

# NetworkPolicy Generation

## Objective

The operator must create a `NetworkPolicy` owned by each `AgentRun`, targeting its agent pods by label. Default behaviour is deny-all egress when no rules are specified. Users opt in to specific egress destinations via `spec.networkPolicy.egress`.

## Tasks

### 1. Add NetworkPolicy spec to AgentRunSpec

In `packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go`, add types:

```go
import networkingv1 "k8s.io/api/networking/v1"

// AgentNetworkPolicy configures the NetworkPolicy created for an AgentRun's pods.
type AgentNetworkPolicy struct {
    // Disabled skips NetworkPolicy creation entirely.
    // Use only when a cluster-level policy already covers these pods.
    Disabled bool `json:"disabled,omitempty"`

    // Egress rules. If empty and Disabled is false, all egress is denied.
    // Specify rules to allow specific destinations (e.g. the provider API endpoint).
    Egress []networkingv1.NetworkPolicyEgressRule `json:"egress,omitempty"`
}

// AgentRunSpec — add field:
type AgentRunSpec struct {
    // ... existing fields ...

    // NetworkPolicy configures the NetworkPolicy applied to agent pods.
    // Defaults to deny-all egress. Set Disabled:true to skip creation.
    NetworkPolicy *AgentNetworkPolicy `json:"networkPolicy,omitempty"`
}
```

Also add to `AgentSpec` (harness-level default):

```go
type AgentSpec struct {
    // ... existing fields ...
    DefaultNetworkPolicy *AgentNetworkPolicy `json:"defaultNetworkPolicy,omitempty"`
}
```

### 2. Add BuildNetworkPolicy in a new builder file

Create `packages/agent/agent-operator-go/internal/builder/networkpolicy.go`:

```go
package builder

import (
    networkingv1 "k8s.io/api/networking/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// BuildNetworkPolicy creates a NetworkPolicy for an AgentRun's pods.
// It targets pods with the label agent.xonovex.com/agent-run=<run.Name>.
// Ingress is always denied. Egress uses the rules from the spec; empty rules = deny all.
func BuildNetworkPolicy(run *agentv1alpha1.AgentRun, np *agentv1alpha1.AgentNetworkPolicy) *networkingv1.NetworkPolicy {
    var egress []networkingv1.NetworkPolicyEgressRule
    if np != nil {
        egress = np.Egress
    }
    // nil egress slice in NetworkPolicySpec means deny-all egress

    return &networkingv1.NetworkPolicy{
        ObjectMeta: metav1.ObjectMeta{
            Name:      run.Name + "-netpol",
            Namespace: run.Namespace,
            Labels: map[string]string{
                "app.kubernetes.io/name":      "agent-operator",
                "app.kubernetes.io/instance":  run.Name,
                "app.kubernetes.io/component": "agent-network-policy",
            },
        },
        Spec: networkingv1.NetworkPolicySpec{
            PodSelector: metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "app.kubernetes.io/instance": run.Name,
                },
            },
            PolicyTypes: []networkingv1.PolicyType{
                networkingv1.PolicyTypeIngress,
                networkingv1.PolicyTypeEgress,
            },
            Ingress: []networkingv1.NetworkPolicyIngressRule{}, // deny all ingress
            Egress:  egress,
        },
    }
}
```

Also create `packages/agent/agent-operator-go/internal/builder/networkpolicy_test.go`:
- `TestBuildNetworkPolicy_DenyAll` — nil egress rules → `Egress` is nil in spec
- `TestBuildNetworkPolicy_WithEgressRules` — rules are preserved
- `TestBuildNetworkPolicy_Labels` — correct labels and pod selector
- `TestBuildNetworkPolicy_PolicyTypes` — both Ingress and Egress types set

### 3. Create and own NetworkPolicy in AgentRun controller

In `packages/agent/agent-operator-go/internal/controller/agentrun_controller.go`, add RBAC annotation and create logic.

Add RBAC annotation near the top of the file:

```go
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
```

In `reconcileStandalone`, after PVC creation and before Job creation:

```go
// Resolve NetworkPolicy from harness or spec
netpol := run.Spec.NetworkPolicy
if netpol == nil && harness != nil {
    netpol = harness.Spec.DefaultNetworkPolicy
}

// Create NetworkPolicy (unless explicitly disabled)
if netpol == nil || !netpol.Disabled {
    np := builder.BuildNetworkPolicy(agentRun, netpol)
    if err := ctrl.SetControllerReference(agentRun, np, r.Scheme); err != nil {
        return ctrl.Result{}, err
    }
    if err := r.Create(ctx, np); err != nil && !errors.IsAlreadyExists(err) {
        log.Error(err, "failed to create network policy")
        return ctrl.Result{}, err
    }
}
```

Apply the same pattern in `reconcileWithWorkspace`.

Add `networkingv1 "k8s.io/api/networking/v1"` to the controller's imports and register ownership:

```go
func (r *AgentRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&agentv1alpha1.AgentRun{}).
        Owns(&batchv1.Job{}).
        Owns(&networkingv1.NetworkPolicy{}).
        Complete(r)
}
```

### 4. Update RBAC role.yaml

In `packages/agent/agent-operator-go/config/rbac/role.yaml`, add:

```yaml
  - apiGroups:
      - networking.k8s.io
    resources:
      - networkpolicies
    verbs:
      - get
      - list
      - watch
      - create
      - update
      - patch
      - delete
```

Also add events permission for the audit events subplan:

```yaml
  - apiGroups:
      - ""
    resources:
      - events
    verbs:
      - create
      - patch
```

### 5. Update resolver defaults

In `packages/agent/agent-operator-go/internal/resolver/defaults.go`, propagate `DefaultNetworkPolicy` from AgentHarness to the resolved defaults struct:

```go
type ResolvedDefaults struct {
    // ... existing fields ...
    NetworkPolicy *agentv1alpha1.AgentNetworkPolicy
}

func ApplyHarnessDefaults(run *agentv1alpha1.AgentRun, harness *agentv1alpha1.AgentHarness) ResolvedDefaults {
    // ...
    defaults.NetworkPolicy = run.Spec.NetworkPolicy
    if defaults.NetworkPolicy == nil && harness != nil {
        defaults.NetworkPolicy = harness.Spec.DefaultNetworkPolicy
    }
    return defaults
}
```

### 6. Update AgentRun webhook to validate egress rules

In `packages/agent/agent-operator-go/internal/webhook/agentrun_webhook.go`, add validation that egress rules don't contain wildcard/open-ended selectors without explicit approval (advisory warning):

```go
if run.Spec.NetworkPolicy != nil && !run.Spec.NetworkPolicy.Disabled {
    for _, rule := range run.Spec.NetworkPolicy.Egress {
        if len(rule.To) == 0 {
            // Empty `to` means allow egress to all destinations — warn but don't block
            return admission.Warnings{"NetworkPolicy egress rule with empty 'to' allows all destinations"}, nil
        }
    }
}
```

### 7. Manually update CRD YAML

In `config/crd/bases/agent.xonovex.com_agentruns.yaml`, add under `spec.properties`:

```yaml
networkPolicy:
  description: NetworkPolicy configures the NetworkPolicy applied to agent pods.
  type: object
  properties:
    disabled:
      type: boolean
    egress:
      type: array
      items:
        type: object
        x-kubernetes-preserve-unknown-fields: true
```

### 8. Add sample

Add a `networkPolicy` example to `config/samples/agentrun_sample.yaml`:

```yaml
spec:
  # ...
  networkPolicy:
    egress:
      - to:
          - ipBlock:
              cidr: 0.0.0.0/0
              except:
                - 10.0.0.0/8
                - 172.16.0.0/12
                - 192.168.0.0/16
        ports:
          - port: 443
            protocol: TCP
```

## Validation Steps

```bash
cd packages/agent/agent-operator-go
go build ./...
go vet ./...
go test ./internal/builder/...
go test ./internal/controller/...
golangci-lint run ./...
```

## Success Criteria

- [ ] `BuildNetworkPolicy` builder exists and is unit tested
- [ ] `AgentRunSpec` has optional `networkPolicy` field
- [ ] Controller creates a NetworkPolicy before the Job, owned by AgentRun
- [ ] Deny-all is the default when no egress rules specified
- [ ] `Disabled: true` skips NetworkPolicy creation
- [ ] RBAC `role.yaml` includes `networking.k8s.io/networkpolicies`
- [ ] CRD YAML updated
- [ ] NetworkPolicy is deleted when AgentRun is deleted (via ownership/GC)
- [ ] New unit tests pass

## Files Modified/Created

- `api/v1alpha1/agentrun_types.go` — `AgentNetworkPolicy`, `NetworkPolicy` field in `AgentRunSpec`/`AgentSpec`
- `api/v1alpha1/zz_generated.deepcopy.go` — manual DeepCopyInto for `AgentNetworkPolicy`
- `internal/builder/networkpolicy.go` (new)
- `internal/builder/networkpolicy_test.go` (new)
- `internal/controller/agentrun_controller.go` — RBAC annotation, create NetworkPolicy
- `internal/resolver/defaults.go` — propagate NetworkPolicy from harness
- `internal/webhook/agentrun_webhook.go` — egress rule warning
- `config/rbac/role.yaml` — add networking.k8s.io and events
- `config/crd/bases/agent.xonovex.com_agentruns.yaml` — manual update
- `config/samples/agentrun_sample.yaml` — example

## Estimated Duration

Medium — ~4h
