---
type: plan
has_subplans: false
parent_plan: plans/agent-go-code-sharing.md
parallel_group: 2
status: complete
dependencies:
  plans:
    - plans/agent-go-code-sharing/subplan-02-shared-agent-go-expand.md
    - plans/agent-go-code-sharing/subplan-03-operator-consume-shared-agents.md
  files:
    - packages/agent/agent-operator-go/go.mod
    - packages/agent/agent-operator-go/internal/resolver/provider.go
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Operator: Consume shared-agent-go Provider Definitions

Enable the operator to reference pre-defined provider presets from shared-agent-go via a `presetRef` field on `AgentProvider`, reducing manual env var configuration for known providers.

## Objective

The operator's `AgentProvider` is a K8s resource where users manually specify all env vars and secrets. The shared-agent-go provider registry already defines complete configurations for Gemini, GLM, GPT-5.2, etc. (base URLs, model mappings, timeout env vars). Adding a `presetRef` to `AgentProvider` lets users reference a named preset and only override what differs — e.g., point to their own secret without re-specifying the full Anthropic-compatible URL config.

Design: **share definitions, not mechanisms**. Preset env vars become defaults; the CRD's `environment` map overrides them. Secret resolution stays in the operator's resolver.

## Current State

**`api/v1alpha1/agentprovider_types.go`**: `AgentProviderSpec` has `Type`, `DisplayName`, `AuthTokenSecretRef`, `Environment`, `CliArgs`.

**`internal/resolver/provider.go`**: Copies `provider.Spec.Environment` into the env map, then injects `ANTHROPIC_AUTH_TOKEN` if `ANTHROPIC_BASE_URL` is present.

**`shared-agent-go/pkg/providers`**: Full registry with Gemini, Gemini-Claude, GLM, GPT-5.2 (claude), and Gemini (opencode) presets.

## Tasks

### 1. Add `PresetRef` field to `AgentProviderSpec`

**File**: `packages/agent/agent-operator-go/api/v1alpha1/agentprovider_types.go`

Add an optional `PresetRef` field:

```go
type AgentProviderSpec struct {
    // PresetRef references a named provider preset from shared-agent-go
    // (e.g., "gemini", "glm", "gpt5-codex" for agent type "claude").
    // Preset env vars are applied as defaults; Environment overrides them.
    // +optional
    PresetRef string `json:"presetRef,omitempty"`

    // AgentType is required when PresetRef is set to disambiguate presets
    // that exist for multiple agent types. Defaults to "claude".
    // +optional
    AgentType string `json:"agentType,omitempty"`

    // ... existing fields unchanged ...
    Type                string            `json:"type,omitempty"`
    DisplayName         string            `json:"displayName,omitempty"`
    AuthTokenSecretRef  *SecretKeyRef     `json:"authTokenSecretRef,omitempty"`
    Environment         map[string]string `json:"environment,omitempty"`
    CliArgs             []string          `json:"cliArgs,omitempty"`
}
```

Since controller-gen is broken with Go 1.25+, also update the CRD YAML manually:
**File**: `packages/agent/agent-operator-go/config/crd/bases/agent.xonovex.io_agentproviders.yaml`

Add `presetRef` and `agentType` under `spec.properties`.

Also update `zz_generated.deepcopy.go` for `AgentProviderSpec` to handle the new string fields (strings are value types, no pointer copying needed — just ensure the deepcopy method is regenerated or manually verified).

### 2. Update `resolver/provider.go` to merge preset env vars

**File**: `packages/agent/agent-operator-go/internal/resolver/provider.go`

Add import for shared providers:
```go
import (
    sharedproviders "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/providers"
    sharedtypes "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)
```

In `ResolveProvider`, before copying `provider.Spec.Environment`, load preset env vars if `PresetRef` is set:

```go
// Load preset env vars as defaults
if provider.Spec.PresetRef != "" {
    agentType := sharedtypes.AgentType(provider.Spec.AgentType)
    if agentType == "" {
        agentType = sharedtypes.AgentClaude
    }
    if preset, err := sharedproviders.GetProvider(provider.Spec.PresetRef, agentType); err == nil {
        for k, v := range sharedproviders.BuildProviderEnv(preset) {
            env[k] = v
        }
    }
}
// CRD environment overrides preset
for k, v := range provider.Spec.Environment {
    env[k] = v
}
```

Apply the same pattern in `resolveInlineProvider` for `ProviderSpec` (which should also get `PresetRef`).

### 3. Add `PresetRef` to `ProviderSpec` (inline provider)

**File**: `packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go`

`ProviderSpec` (used for inline providers in `AgentRun`) should also support preset references:

```go
type ProviderSpec struct {
    // +optional
    PresetRef string `json:"presetRef,omitempty"`
    // +optional
    AgentType string `json:"agentType,omitempty"`
    // ... existing fields ...
}
```

Update `resolveInlineProvider` in `resolver/provider.go` to apply the same preset merging.

### 4. Update webhook validation

**File**: `packages/agent/agent-operator-go/internal/webhook/agentprovider_webhook.go`

Add validation: if `PresetRef` is set, verify it resolves to a known provider. Emit a warning (not an error) if the preset is unknown, to allow forward compatibility.

### 5. Add tests for preset resolution

**File**: `packages/agent/agent-operator-go/internal/resolver/` (test file)

Test cases:
- Provider with valid `PresetRef` gets preset env vars
- Provider `Environment` overrides preset env vars (not the other way)
- Provider without `PresetRef` behaves identically to before
- Unknown `PresetRef` does not cause an error (soft failure)

## Validation Steps

```bash
cd packages/agent/agent-operator-go
go mod tidy
go build ./...
go vet ./...
go test ./...
# go test -tags=integration ./test/integration/
```

## Success Criteria

- [ ] `AgentProviderSpec` has `PresetRef` and `AgentType` fields
- [ ] `ProviderSpec` (inline) also has `PresetRef`
- [ ] Resolver loads preset env vars as defaults when `PresetRef` is set
- [ ] CRD `environment` always overrides preset env vars
- [ ] Unknown `PresetRef` produces a warning, not an error
- [ ] CRD YAML updated for new fields
- [ ] All existing resolver tests pass
- [ ] New preset tests cover override semantics

## Files Modified/Created

- `packages/agent/agent-operator-go/api/v1alpha1/agentprovider_types.go` (modified)
- `packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go` (modified — ProviderSpec)
- `packages/agent/agent-operator-go/api/v1alpha1/zz_generated.deepcopy.go` (modified)
- `packages/agent/agent-operator-go/internal/resolver/provider.go` (modified)
- `packages/agent/agent-operator-go/internal/webhook/agentprovider_webhook.go` (modified)
- `packages/agent/agent-operator-go/config/crd/bases/agent.xonovex.io_agentproviders.yaml` (modified)
- `packages/agent/agent-operator-go/config/crd/bases/agent.xonovex.io_agentruns.yaml` (modified)

## Estimated Duration

Medium — ~2 hours
