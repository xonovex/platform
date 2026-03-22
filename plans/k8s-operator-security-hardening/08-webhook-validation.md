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
    - packages/agent/agent-operator-go/internal/webhook/agentprovider_webhook.go
    - packages/agent/agent-operator-go/internal/webhook/agentworkspace_webhook.go
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pending
---

# Webhook Validation Completion

## Objective

Complete the two incomplete webhook validators: (1) `AgentProviderWebhook` which currently has an empty `validate()` and provides no protection against injection in `environment` keys or `cliArgs`, (2) any remaining URL/branch format validation in `AgentWorkspaceWebhook` not already covered by subplan 01 (the workspace webhook already validates repo URL/branch; this subplan adds the storage size format check and shared volume constraints).

Note: The `AgentWorkspace` URL/branch validation is already addressed in subplan 01 (`shell-injection-fix`). This subplan focuses on remaining gaps: AgentProvider injection vectors and AgentWorkspace storage validation.

## Tasks

### 1. Implement AgentProvider webhook validation

In `packages/agent/agent-operator-go/internal/webhook/agentprovider_webhook.go`, update `validate()`:

```go
import (
    "regexp"
    "strings"
)

// envKeyPattern: env var names must be alphanumeric + underscore, cannot start with digit.
var envKeyPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Blocked env key prefixes that could interfere with agent or system behaviour.
var blockedEnvKeyPrefixes = []string{
    "LD_",       // dynamic linker
    "DYLD_",     // macOS dynamic linker
    "PYTHONPATH",
    "RUBYOPT",
    "NODE_OPTIONS",
    "JAVA_TOOL_OPTIONS",
}

func (w *AgentProviderWebhook) validate(provider *agentv1alpha1.AgentProvider) (admission.Warnings, error) {
    // Validate secretRef format
    if provider.Spec.AuthTokenSecretRef != nil {
        ref := provider.Spec.AuthTokenSecretRef
        if ref.Name == "" {
            return nil, fmt.Errorf("authTokenSecretRef.name is required")
        }
        if ref.Key == "" {
            return nil, fmt.Errorf("authTokenSecretRef.key is required")
        }
        // K8s secret names: lowercase alphanumeric and hyphens, max 253 chars
        if !k8sNamePattern.MatchString(ref.Name) {
            return nil, fmt.Errorf("authTokenSecretRef.name %q is not a valid Kubernetes resource name", ref.Name)
        }
    }

    // Validate environment variable keys
    for key := range provider.Spec.Environment {
        if !envKeyPattern.MatchString(key) {
            return nil, fmt.Errorf("environment key %q is not a valid env var name", key)
        }
        upperKey := strings.ToUpper(key)
        for _, blocked := range blockedEnvKeyPrefixes {
            if strings.HasPrefix(upperKey, blocked) {
                return nil, fmt.Errorf("environment key %q is not allowed (blocked prefix %q)", key, blocked)
            }
        }
    }

    // Validate cliArgs: no shell metacharacters, no empty strings
    for i, arg := range provider.Spec.CliArgs {
        if arg == "" {
            return nil, fmt.Errorf("cliArgs[%d] is empty", i)
        }
        if containsShellMetachars(arg) {
            return nil, fmt.Errorf("cliArgs[%d] %q contains shell metacharacters", i, arg)
        }
    }

    return nil, nil
}

// k8sNamePattern: lowercase alphanumeric and hyphens, max 253 chars.
var k8sNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{0,251}[a-z0-9]$|^[a-z0-9]$`)

// containsShellMetachars returns true if the string contains shell injection characters.
func containsShellMetachars(s string) bool {
    const metachars = ";|&$`\\\"'<>(){}!#~"
    return strings.ContainsAny(s, metachars)
}
```

### 2. Add storage size validation to AgentWorkspace webhook

In `packages/agent/agent-operator-go/internal/webhook/agentworkspace_webhook.go`, the existing `validate()` (already updated in subplan 01 for repo URL/branch) should also validate storage sizes:

```go
import "k8s.io/apimachinery/pkg/api/resource"

func (w *AgentWorkspaceWebhook) validate(ws *agentv1alpha1.AgentWorkspace) (admission.Warnings, error) {
    // ... existing repo URL/branch validation from subplan 01 ...

    // Validate storageSize is a valid Kubernetes resource quantity
    if ws.Spec.StorageSize != "" {
        if _, err := resource.ParseQuantity(ws.Spec.StorageSize); err != nil {
            return nil, fmt.Errorf("storageSize %q is not a valid resource quantity: %v", ws.Spec.StorageSize, err)
        }
    }

    // Validate shared volume storage sizes and mount paths
    mountPaths := make(map[string]bool)
    for _, vol := range ws.Spec.SharedVolumes {
        if vol.StorageSize != "" {
            if _, err := resource.ParseQuantity(vol.StorageSize); err != nil {
                return nil, fmt.Errorf("sharedVolumes[%q].storageSize %q is not a valid resource quantity: %v",
                    vol.Name, vol.StorageSize, err)
            }
        }
        // Check for duplicate mount paths
        if mountPaths[vol.MountPath] {
            return nil, fmt.Errorf("duplicate mountPath %q in sharedVolumes", vol.MountPath)
        }
        mountPaths[vol.MountPath] = true

        // MountPath must be absolute
        if !strings.HasPrefix(vol.MountPath, "/") {
            return nil, fmt.Errorf("sharedVolumes[%q].mountPath %q must be an absolute path",
                vol.Name, vol.MountPath)
        }
    }

    return nil, nil
}
```

### 3. Extract containsShellMetachars to shared validator

Move `containsShellMetachars` to the `internal/validator` package created in subplan 01 to avoid duplication:

In `packages/agent/agent-operator-go/internal/validator/repository.go` (from subplan 01), add:

```go
// ContainsShellMetachars returns true if s contains characters that have special
// meaning in POSIX shell and could enable command injection.
func ContainsShellMetachars(s string) bool {
    const metachars = ";|&$`\\\"'<>(){}!#~\n\r"
    return strings.ContainsAny(s, metachars)
}
```

Then use it in both `agentprovider_webhook.go` and wherever else injection checks are needed.

### 4. Add AgentProvider webhook unit tests

Create `packages/agent/agent-operator-go/internal/webhook/agentprovider_webhook_test.go`:

```go
func TestAgentProviderWebhook_ValidSecretRef(t *testing.T) { ... }
func TestAgentProviderWebhook_EmptySecretName(t *testing.T) { ... }
func TestAgentProviderWebhook_EmptySecretKey(t *testing.T) { ... }
func TestAgentProviderWebhook_InvalidSecretName(t *testing.T) { ... } // e.g. "MySecret" (uppercase)
func TestAgentProviderWebhook_ValidEnvKey(t *testing.T) { ... }
func TestAgentProviderWebhook_InvalidEnvKey(t *testing.T) { ... }     // "1INVALID"
func TestAgentProviderWebhook_BlockedEnvKey(t *testing.T) { ... }     // "LD_PRELOAD"
func TestAgentProviderWebhook_ValidCliArgs(t *testing.T) { ... }
func TestAgentProviderWebhook_EmptyCliArg(t *testing.T) { ... }
func TestAgentProviderWebhook_InjectionCliArg(t *testing.T) { ... }   // "--model; rm -rf /"
func TestAgentProviderWebhook_NilSecretRef(t *testing.T) { ... }      // allowed (no auth ref)
```

### 5. Add AgentWorkspace webhook storage tests

Extend `packages/agent/agent-operator-go/internal/webhook/agentworkspace_webhook_test.go`:

```go
func TestAgentWorkspaceWebhook_ValidStorageSize(t *testing.T) { ... }  // "10Gi" passes
func TestAgentWorkspaceWebhook_InvalidStorageSize(t *testing.T) { ... } // "10gigabytes" rejected
func TestAgentWorkspaceWebhook_DuplicateMountPath(t *testing.T) { ... }
func TestAgentWorkspaceWebhook_RelativeMountPath(t *testing.T) { ... }
func TestAgentWorkspaceWebhook_ValidSharedVolumes(t *testing.T) { ... }
```

### 6. Verify existing webhook tests still pass

After subplan 01 adds URL/branch validation to `AgentWorkspaceWebhook`, the tests in `agentworkspace_webhook_test.go` need to be compatible. Check that the existing workspace webhook tests use valid repo URLs (they use `https://github.com/example/repo.git` which passes the validator).

## Validation Steps

```bash
cd packages/agent/agent-operator-go
go build ./...
go vet ./...
go test ./internal/validator/...
go test ./internal/webhook/...
golangci-lint run ./...
```

## Success Criteria

- [ ] `AgentProviderWebhook.validate()` is no longer empty
- [ ] Provider secretRef name validated against K8s naming rules
- [ ] Provider environment keys validated against env var name pattern
- [ ] Blocked env key prefixes (`LD_`, `DYLD_`, `NODE_OPTIONS`, etc.) rejected
- [ ] Provider cliArgs validated against shell metachar pattern
- [ ] AgentWorkspace webhook validates storage quantity format
- [ ] Duplicate mount paths rejected
- [ ] Relative mount paths rejected
- [ ] All new webhook tests pass
- [ ] `ContainsShellMetachars` lives in `internal/validator` package (no duplication)

## Files Modified/Created

- `internal/webhook/agentprovider_webhook.go` — full `validate()` implementation
- `internal/webhook/agentprovider_webhook_test.go` (new or updated)
- `internal/webhook/agentworkspace_webhook.go` — storage/mount path validation
- `internal/webhook/agentworkspace_webhook_test.go` — new storage tests
- `internal/validator/repository.go` — add `ContainsShellMetachars`
- `internal/validator/repository_test.go` — tests for ContainsShellMetachars

## Estimated Duration

Small — ~2h
