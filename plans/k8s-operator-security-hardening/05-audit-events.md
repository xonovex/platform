---
type: plan
has_subplans: false
parent_plan: plans/k8s-operator-security-hardening.md
parallel_group: 3
status: pending
dependencies:
  plans:
    - plans/k8s-operator-security-hardening/03-network-policy.md
    - plans/k8s-operator-security-hardening/04-agent-policy-crd.md
  files:
    - packages/agent/agent-operator-go/internal/controller/agentrun_controller.go
    - packages/agent/agent-operator-go/internal/controller/agentworkspace_controller.go
    - packages/agent/agent-operator-go/internal/controller/agentprovider_controller.go
    - packages/agent/agent-operator-go/cmd/operator/main.go
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Audit Events

## Objective

Emit Kubernetes Events for all security-relevant controller actions so that a security team can audit agent activity via `kubectl get events` or a centralized logging system. Add structured log fields to existing controller logs for correlation.

## Tasks

### 1. Add EventRecorder to all reconcilers

In `packages/agent/agent-operator-go/internal/controller/agentrun_controller.go`, add the recorder field:

```go
import "k8s.io/client-go/tools/record"

type AgentRunReconciler struct {
    client.Client
    Scheme   *runtime.Scheme
    Recorder record.EventRecorder
}
```

Apply the same addition to `AgentWorkspaceReconciler` and `AgentProviderReconciler`.

### 2. Wire EventRecorder in main.go

In `packages/agent/agent-operator-go/cmd/operator/main.go`:

```go
if err = (&controller.AgentRunReconciler{
    Client:   mgr.GetClient(),
    Scheme:   mgr.GetScheme(),
    Recorder: mgr.GetEventRecorderFor("agent-operator"),
}).SetupWithManager(mgr); err != nil { ... }

if err = (&controller.AgentProviderReconciler{
    Client:   mgr.GetClient(),
    Scheme:   mgr.GetScheme(),
    Recorder: mgr.GetEventRecorderFor("agent-operator"),
}).SetupWithManager(mgr); err != nil { ... }

if err = (&controller.AgentWorkspaceReconciler{
    Client:   mgr.GetClient(),
    Scheme:   mgr.GetScheme(),
    Recorder: mgr.GetEventRecorderFor("agent-operator"),
}).SetupWithManager(mgr); err != nil { ... }
```

### 3. Emit audit events from AgentRun controller

In `agentrun_controller.go`, add event recording at key lifecycle points. Use event reason strings that are easy to query/filter.

**On Job creation** (in `reconcileStandalone` and `reconcileWithWorkspace`):

```go
r.Recorder.Eventf(agentRun, corev1.EventTypeNormal, "AgentRunStarted",
    "Created Job %s (agent=%s, provider=%s, runtimeClass=%s)",
    jobName, string(agentType), agentRun.Spec.ProviderRef, ptrOrEmpty(agentRun.Spec.RuntimeClassName))
```

**On Succeeded phase**:

```go
r.Recorder.Event(agentRun, corev1.EventTypeNormal, "AgentRunSucceeded",
    "Agent Job completed successfully")
```

**On Failed phase**:

```go
r.Recorder.Eventf(agentRun, corev1.EventTypeWarning, "AgentRunFailed",
    "Agent Job failed: %s", message)
```

**On TimedOut phase**:

```go
r.Recorder.Eventf(agentRun, corev1.EventTypeWarning, "AgentRunTimedOut",
    "Agent Job exceeded timeout of %v", agentRun.Spec.Timeout.Duration)
```

**On NetworkPolicy creation**:

```go
r.Recorder.Eventf(agentRun, corev1.EventTypeNormal, "NetworkPolicyCreated",
    "Created NetworkPolicy %s", np.Name)
```

**On policy rejection** (from webhook — webhook can't emit events directly, so the controller records a Forbidden event if it sees a policy violation on an admitted run that somehow bypassed the webhook):

No controller-side policy event needed — webhook is the enforcement point.

### 4. Emit audit events from AgentWorkspace controller

In `agentworkspace_controller.go`:

**On init Job creation**:

```go
r.Recorder.Eventf(&ws, corev1.EventTypeNormal, "WorkspaceInitStarted",
    "Created init Job %s to clone %s", initJobName, ws.Spec.Repository.URL)
```

**On Ready phase**:

```go
r.Recorder.Event(&ws, corev1.EventTypeNormal, "WorkspaceReady",
    "Repository cloned successfully, workspace is ready")
```

**On Failed phase**:

```go
r.Recorder.Eventf(&ws, corev1.EventTypeWarning, "WorkspaceFailed",
    "Init Job failed: %s", message)
```

### 5. Emit audit events from AgentProvider controller

In `agentprovider_controller.go`, record when the provider secret is resolved:

```go
r.Recorder.Eventf(&provider, corev1.EventTypeNormal, "ProviderSecretResolved",
    "Secret %s key %s resolved successfully",
    provider.Spec.AuthTokenSecretRef.Name, provider.Spec.AuthTokenSecretRef.Key)
```

When secret is missing or key not found:

```go
r.Recorder.Eventf(&provider, corev1.EventTypeWarning, "ProviderSecretMissing",
    "Secret %s not found or key %s missing",
    provider.Spec.AuthTokenSecretRef.Name, provider.Spec.AuthTokenSecretRef.Key)
```

### 6. Add structured log fields throughout controllers

Replace bare `log.Error(err, "message")` calls with structured fields in `agentrun_controller.go`:

```go
log := log.FromContext(ctx).WithValues(
    "agentRun", agentRun.Name,
    "namespace", agentRun.Namespace,
)

// Then use:
log.Error(err, "failed to resolve provider",
    "providerRef", agentRun.Spec.ProviderRef)

log.Info("creating Job",
    "jobName", jobName,
    "agentType", agentType,
    "runtimeClass", ptrOrEmpty(agentRun.Spec.RuntimeClassName))
```

Add a helper:

```go
func ptrOrEmpty(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}
```

Apply the same structured logging pattern in `agentworkspace_controller.go` and `agentprovider_controller.go`.

### 7. Add RBAC for events

The `events` create/patch permission should already be added by the network-policy subplan (03). Verify `role.yaml` includes:

```yaml
  - apiGroups:
      - ""
    resources:
      - events
    verbs:
      - create
      - patch
```

### 8. Update controller tests

In the existing controller unit test files, update reconciler instantiation to include a fake recorder:

```go
import "k8s.io/client-go/tools/record"

reconciler := &AgentRunReconciler{
    Client:   fakeClient,
    Scheme:   scheme,
    Recorder: record.NewFakeRecorder(100),
}
```

Add assertions that expected events are emitted:

```go
// After reconcile:
events := recorder.Events
select {
case event := <-events:
    assert.Contains(t, event, "AgentRunStarted")
default:
    t.Error("expected AgentRunStarted event")
}
```

## Validation Steps

```bash
cd packages/agent/agent-operator-go
go build ./...
go vet ./...
go test ./internal/controller/...
golangci-lint run ./...
```

## Success Criteria

- [ ] All three controllers have an `EventRecorder` field
- [ ] `main.go` wires `mgr.GetEventRecorderFor("agent-operator")` into all reconcilers
- [ ] AgentRun controller emits events for: Started, Succeeded, Failed, TimedOut, NetworkPolicyCreated
- [ ] AgentWorkspace controller emits events for: InitStarted, Ready, Failed
- [ ] AgentProvider controller emits events for: SecretResolved, SecretMissing
- [ ] All controller log calls include structured `agentRun`/`namespace` fields
- [ ] `ptrOrEmpty` helper added (or equivalent)
- [ ] Controller tests updated with `FakeRecorder` and event assertions
- [ ] RBAC has events create/patch

## Files Modified/Created

- `internal/controller/agentrun_controller.go` — add Recorder, emit events, structured logging
- `internal/controller/agentworkspace_controller.go` — add Recorder, emit events, structured logging
- `internal/controller/agentprovider_controller.go` — add Recorder, emit events, structured logging
- `cmd/operator/main.go` — wire GetEventRecorderFor
- `config/rbac/role.yaml` — events create/patch (may already exist from subplan 03)

## Estimated Duration

Small-Medium — ~3h
