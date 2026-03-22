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
    - packages/agent/agent-operator-go/api/v1alpha1/
    - packages/agent/agent-operator-go/internal/webhook/agentrun_webhook.go
    - packages/agent/agent-operator-go/cmd/operator/main.go
    - packages/agent/agent-operator-go/config/rbac/role.yaml
    - packages/agent/agent-operator-go/config/crd/bases/
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Central Policy CRD (AgentPolicy)

## Objective

Add a namespace-scoped `AgentPolicy` CRD with `enforced` constraints (cannot be overridden by AgentRun) and `defaults` (can be overridden). The `AgentRun` admission webhook looks up the policy in the same namespace and rejects runs that violate enforced constraints. If no policy exists in the namespace, all AgentRuns are allowed.

## Tasks

### 1. Define AgentPolicy types

Create `packages/agent/agent-operator-go/api/v1alpha1/agentpolicy_types.go`:

```go
package v1alpha1

import (
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgentPolicyEnforced defines constraints that AgentRuns in this namespace cannot override.
type AgentPolicyEnforced struct {
    // RuntimeClassName, if set, requires all AgentRuns to use this runtimeClassName.
    RuntimeClassName *string `json:"runtimeClassName,omitempty"`

    // RequireSecurityContext, if true, prevents AgentRuns from setting
    // SecurityContext fields that weaken the hardened defaults
    // (e.g. RunAsNonRoot=false, AllowPrivilegeEscalation=true).
    RequireSecurityContext bool `json:"requireSecurityContext,omitempty"`

    // RequireNetworkPolicy, if true, requires AgentRuns to have a NetworkPolicy
    // (i.e. spec.networkPolicy must not be Disabled).
    RequireNetworkPolicy bool `json:"requireNetworkPolicy,omitempty"`

    // MaxTimeout is the maximum allowed timeout for AgentRuns.
    MaxTimeout *metav1.Duration `json:"maxTimeout,omitempty"`

    // MaxResources defines the upper bound for any single container's resource limits.
    MaxResources *corev1.ResourceList `json:"maxResources,omitempty"`

    // AllowedImages is a list of allowed container image prefixes.
    // If set, AgentRun.Spec.Image must match one of these prefixes.
    AllowedImages []string `json:"allowedImages,omitempty"`

    // AllowedRuntimeClassNames lists permitted runtimeClassNames.
    // If non-empty, AgentRun.Spec.RuntimeClassName must be in this list.
    AllowedRuntimeClassNames []string `json:"allowedRuntimeClassNames,omitempty"`
}

// AgentPolicyDefaults defines overridable defaults applied when AgentRun fields are absent.
type AgentPolicyDefaults struct {
    // Image is the default container image when AgentRun.Spec.Image is not set.
    Image string `json:"image,omitempty"`

    // Timeout is the default timeout when AgentRun.Spec.Timeout is not set.
    Timeout *metav1.Duration `json:"timeout,omitempty"`

    // RuntimeClassName is the default runtimeClassName when not set on AgentRun or AgentHarness.
    RuntimeClassName *string `json:"runtimeClassName,omitempty"`
}

// AgentPolicySpec defines the desired state of AgentPolicy.
type AgentPolicySpec struct {
    // Enforced constraints — AgentRuns that violate these are rejected by the webhook.
    Enforced AgentPolicyEnforced `json:"enforced,omitempty"`

    // Defaults — applied when AgentRun fields are not set.
    Defaults AgentPolicyDefaults `json:"defaults,omitempty"`
}

// AgentPolicyStatus defines the observed state of AgentPolicy.
type AgentPolicyStatus struct {
    // Conditions of the AgentPolicy.
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AgentPolicy defines enforced security constraints and defaults for AgentRuns in a namespace.
type AgentPolicy struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   AgentPolicySpec   `json:"spec,omitempty"`
    Status AgentPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentPolicyList contains a list of AgentPolicy.
type AgentPolicyList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []AgentPolicy `json:"items"`
}

func init() {
    SchemeBuilder.Register(&AgentPolicy{}, &AgentPolicyList{})
}
```

Also manually add `DeepCopyInto` and `DeepCopyObject` methods to `zz_generated.deepcopy.go` for `AgentPolicy`, `AgentPolicyList`, `AgentPolicySpec`, `AgentPolicyEnforced`, `AgentPolicyDefaults`, `AgentPolicyStatus`.

### 2. Add policy enforcement to AgentRun webhook

In `packages/agent/agent-operator-go/internal/webhook/agentrun_webhook.go`, the webhook needs access to the K8s client to look up the namespace's AgentPolicy. Change `AgentRunWebhook` to hold a client:

```go
type AgentRunWebhook struct {
    Client client.Client
}
```

Update `SetupWebhookWithManager` to receive the manager client:

```go
func (w *AgentRunWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
    w.Client = mgr.GetClient()
    return ctrl.NewWebhookManagedBy(mgr).
        For(&agentv1alpha1.AgentRun{}).
        WithDefaulter(w).
        WithValidator(w).
        Complete()
}
```

Add policy lookup and enforcement in `validate()`:

```go
func (w *AgentRunWebhook) validate(run *agentv1alpha1.AgentRun) (admission.Warnings, error) {
    // ... existing checks ...

    // Look up AgentPolicy in the namespace (use first policy found)
    ctx := context.Background()
    var policyList agentv1alpha1.AgentPolicyList
    if err := w.Client.List(ctx, &policyList, client.InNamespace(run.Namespace)); err != nil {
        // If we can't read policies, fail open (log but allow)
        return admission.Warnings{"AgentPolicy lookup failed: " + err.Error()}, nil
    }
    if len(policyList.Items) > 0 {
        policy := &policyList.Items[0]
        if err := enforcePolicy(run, policy); err != nil {
            return nil, err
        }
    }

    return nil, nil
}

func enforcePolicy(run *agentv1alpha1.AgentRun, policy *agentv1alpha1.AgentPolicy) error {
    e := policy.Spec.Enforced

    // Enforce runtimeClassName
    if e.RuntimeClassName != nil {
        rc := run.Spec.RuntimeClassName
        if rc == nil || *rc != *e.RuntimeClassName {
            return fmt.Errorf("policy requires runtimeClassName %q", *e.RuntimeClassName)
        }
    }

    // Enforce allowed runtime class names
    if len(e.AllowedRuntimeClassNames) > 0 {
        allowed := false
        for _, name := range e.AllowedRuntimeClassNames {
            if run.Spec.RuntimeClassName != nil && *run.Spec.RuntimeClassName == name {
                allowed = true
                break
            }
        }
        if !allowed {
            return fmt.Errorf("runtimeClassName must be one of %v", e.AllowedRuntimeClassNames)
        }
    }

    // Enforce no privilege escalation weakening
    if e.RequireSecurityContext && run.Spec.SecurityContext != nil {
        sc := run.Spec.SecurityContext
        if sc.AllowPrivilegeEscalation != nil && *sc.AllowPrivilegeEscalation {
            return fmt.Errorf("policy prohibits AllowPrivilegeEscalation=true")
        }
        if sc.RunAsNonRoot != nil && !*sc.RunAsNonRoot {
            return fmt.Errorf("policy requires RunAsNonRoot=true")
        }
    }

    // Enforce network policy required
    if e.RequireNetworkPolicy {
        if run.Spec.NetworkPolicy != nil && run.Spec.NetworkPolicy.Disabled {
            return fmt.Errorf("policy requires NetworkPolicy to be enabled")
        }
    }

    // Enforce max timeout
    if e.MaxTimeout != nil && run.Spec.Timeout != nil {
        if run.Spec.Timeout.Duration > e.MaxTimeout.Duration {
            return fmt.Errorf("timeout %v exceeds policy maximum %v", run.Spec.Timeout.Duration, e.MaxTimeout.Duration)
        }
    }

    // Enforce allowed images
    if len(e.AllowedImages) > 0 && run.Spec.Image != "" {
        allowed := false
        for _, prefix := range e.AllowedImages {
            if strings.HasPrefix(run.Spec.Image, prefix) {
                allowed = true
                break
            }
        }
        if !allowed {
            return fmt.Errorf("image %q is not in the allowed images list", run.Spec.Image)
        }
    }

    return nil
}
```

### 3. Register AgentPolicy in scheme

In `packages/agent/agent-operator-go/cmd/operator/main.go`:

```go
// init() already calls agentv1alpha1.AddToScheme(scheme)
// AgentPolicy is registered in its init() call — no change needed to main.go
// But register the webhook:

import "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/webhook"

// In main(), after controllers, set up webhooks:
if err = (&webhook.AgentRunWebhook{}).SetupWebhookWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to set up webhook", "webhook", "AgentRun")
    os.Exit(1)
}
```

Check existing webhook setup in `main.go` — if webhooks aren't registered yet, add the webhook server setup block:

```go
if err = mgr.GetWebhookServer().Register("/validate-agent-xonovex-com-v1alpha1-agentrun",
    &webhook.Admission{Handler: ...}); err != nil { ... }
```

If webhooks are already registered via `SetupWebhookWithManager`, just ensure `AgentRunWebhook` is instantiated with the client.

### 4. Update RBAC to allow reading AgentPolicy

In `config/rbac/role.yaml`, add:

```yaml
  - apiGroups:
      - agent.xonovex.com
    resources:
      - agentpolicies
    verbs:
      - get
      - list
      - watch
```

Also add `agentpolicies/status` with `get;update;patch`.

### 5. Create CRD YAML manually

Create `packages/agent/agent-operator-go/config/crd/bases/agent.xonovex.com_agentpolicies.yaml`:

Minimal CRD with `spec.properties` for `enforced` and `defaults`. Use `x-kubernetes-preserve-unknown-fields: true` for complex nested types. Follow the pattern from existing CRD files.

Update `config/crd/kustomization.yaml` to include the new CRD file.

### 6. Add sample AgentPolicy

Create `packages/agent/agent-operator-go/config/samples/agentpolicy_sample.yaml`:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentPolicy
metadata:
  name: enterprise-policy
  namespace: ai-agents
spec:
  enforced:
    runtimeClassName: kata
    requireSecurityContext: true
    requireNetworkPolicy: true
    maxTimeout: 2h0m0s
    allowedImages:
      - ghcr.io/your-org/
      - ghcr.io/xonovex/
    allowedRuntimeClassNames:
      - kata
      - gvisor
  defaults:
    image: ghcr.io/xonovex/agent-image:latest
    timeout: 30m0s
```

### 7. Add webhook unit tests

Create `packages/agent/agent-operator-go/internal/webhook/agentpolicy_test.go`:

- `TestEnforcePolicy_AllowsCompliantRun` — compliant AgentRun passes
- `TestEnforcePolicy_RejectsWrongRuntimeClass` — wrong runtimeClassName rejected
- `TestEnforcePolicy_RejectsPrivEsc` — AllowPrivilegeEscalation=true rejected when requireSecurityContext
- `TestEnforcePolicy_RejectsDisabledNetworkPolicy` — NetworkPolicy.Disabled=true rejected when required
- `TestEnforcePolicy_RejectsExceededTimeout` — timeout over max rejected
- `TestEnforcePolicy_RejectsDisallowedImage` — image not in prefix list rejected
- `TestEnforcePolicy_NoPolicy_Allows` — no policy in namespace → allowed

### 8. Add AgentPolicy webhook

Create `packages/agent/agent-operator-go/internal/webhook/agentpolicy_webhook.go`:

```go
package webhook

// AgentPolicyWebhook validates AgentPolicy resources.
// Validates: maxTimeout is positive, maxResources quantities are valid, no duplicate runtimeClassNames.
type AgentPolicyWebhook struct{}
```

## Validation Steps

```bash
cd packages/agent/agent-operator-go
go build ./...
go vet ./...
go test ./api/...
go test ./internal/webhook/...
golangci-lint run ./...
```

## Success Criteria

- [ ] `AgentPolicy` CRD type defined with `enforced` and `defaults` sections
- [ ] `zz_generated.deepcopy.go` has correct DeepCopy methods
- [ ] AgentRun webhook enforces all `enforced` constraints when a policy exists
- [ ] No policy in namespace → all runs allowed (fail open)
- [ ] RBAC includes read permissions for `agentpolicies`
- [ ] CRD YAML created and added to `kustomization.yaml`
- [ ] Sample policy YAML added
- [ ] All policy enforcement unit tests pass

## Files Modified/Created

- `api/v1alpha1/agentpolicy_types.go` (new)
- `api/v1alpha1/zz_generated.deepcopy.go` — manual additions for AgentPolicy types
- `internal/webhook/agentrun_webhook.go` — add Client field, policy lookup, enforcePolicy
- `internal/webhook/agentpolicy_webhook.go` (new)
- `internal/webhook/agentpolicy_test.go` (new)
- `cmd/operator/main.go` — instantiate AgentRunWebhook with client
- `config/rbac/role.yaml` — add agentpolicies
- `config/crd/bases/agent.xonovex.com_agentpolicies.yaml` (new)
- `config/crd/kustomization.yaml` — add new CRD
- `config/samples/agentpolicy_sample.yaml` (new)

## Estimated Duration

Medium — ~5h
