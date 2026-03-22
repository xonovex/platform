---
type: plan
has_subplans: false
parent_plan: plans/agent-go-code-sharing.md
parallel_group: 2
status: complete
dependencies:
  plans:
    - plans/agent-go-code-sharing/subplan-02-shared-agent-go-expand.md
  files:
    - packages/shared/shared-agent-go/pkg/worktree/vcs.go
    - packages/agent/agent-operator-go/internal/builder/harness_claude.go
    - packages/agent/agent-operator-go/internal/builder/harness_opencode.go
    - packages/agent/agent-operator-go/internal/builder/harness.go
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Operator: Consume shared-agent-go for Agent Command Building

Refactor operator harness builders to use `shared-agent-go/pkg/agents` for agent binary names and base arg construction, eliminating the duplicate command-building logic.

## Objective

The operator's `ClaudeCommandBuilder` and `OpencodeCommandBuilder` duplicate what `shared-agent-go/pkg/agents` already does. The operator always runs agents in a K8s Job (sandbox=true equivalent), so it can use `agents.BuildClaudeArgs()` and `agents.BuildOpencodeArgs()` for base args and append its prompt-specific args on top.

## Current State

**`internal/builder/harness_claude.go`** (13 lines):
```go
func (c *ClaudeCommandBuilder) Command(run *AgentRun) ([]string, []string) {
    args := []string{"--permission-mode", "bypassPermissions"}
    if run.Spec.Prompt != "" {
        args = append(args, "--print", "--prompt", run.Spec.Prompt)
    }
    return []string{"claude"}, args
}
```

**`internal/builder/harness_opencode.go`** (12 lines):
```go
func (o *OpencodeCommandBuilder) Command(run *AgentRun) ([]string, []string) {
    var args []string
    if run.Spec.Provider != nil && len(run.Spec.Provider.CliArgs) > 0 {
        args = append(args, run.Spec.Provider.CliArgs...)
    }
    return []string{"opencode"}, args
}
```

**`api/v1alpha1/agentharness_types.go`**: defines `AgentType` as `string` with constants `AgentTypeClaude = "claude"` and `AgentTypeOpencode = "opencode"`.

## Tasks

### 1. Activate shared-agent-go import in operator go.mod

**File**: `packages/agent/agent-operator-go/go.mod`

The `replace` directive already points to the local shared-agent-go. Add the actual import requirement:

```
require (
    ...
    github.com/xonovex/platform/packages/shared/shared-agent-go v0.0.0
)
```

Run `go mod tidy` to verify the replace directive resolves correctly.

### 2. Refactor `harness_claude.go` to use shared agents

**File**: `packages/agent/agent-operator-go/internal/builder/harness_claude.go`

```go
package builder

import (
    "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
    "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
    agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// ClaudeCommandBuilder builds command/args for Claude Code
type ClaudeCommandBuilder struct{}

func (c *ClaudeCommandBuilder) Command(run *agentv1alpha1.AgentRun) ([]string, []string) {
    agent, _ := agents.GetAgent(types.AgentClaude)
    // Operator always runs in sandbox mode (K8s Job = isolated environment)
    baseArgs := agents.BuildClaudeArgs(nil, types.AgentExecOptions{Sandbox: true})
    args := baseArgs
    if run.Spec.Prompt != "" {
        args = append(args, "--print", "--prompt", run.Spec.Prompt)
    }
    return []string{agent.Binary}, args
}
```

### 3. Refactor `harness_opencode.go` to use shared agents

**File**: `packages/agent/agent-operator-go/internal/builder/harness_opencode.go`

```go
package builder

import (
    "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
    "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
    agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// OpencodeCommandBuilder builds command/args for Opencode
type OpencodeCommandBuilder struct{}

func (o *OpencodeCommandBuilder) Command(run *agentv1alpha1.AgentRun) ([]string, []string) {
    agent, _ := agents.GetAgent(types.AgentOpencode)
    var providerCliArgs []string
    if run.Spec.Provider != nil {
        providerCliArgs = run.Spec.Provider.CliArgs
    }
    args := agents.BuildOpencodeArgs(nil, types.AgentExecOptions{
        Sandbox:         true,
        ProviderCliArgs: providerCliArgs,
    })
    return []string{agent.Binary}, args
}
```

### 4. Keep `harness.go` registry unchanged

**File**: `packages/agent/agent-operator-go/internal/builder/harness.go`

No changes needed — the `HarnessCommandBuilder` interface and registry remain. The `AgentType` keys still use `agentv1alpha1.AgentTypeClaude` / `AgentTypeOpencode` (string `"claude"` / `"opencode"`), which match `types.AgentClaude` / `types.AgentOpencode` in shared-agent-go. No type aliasing needed.

### 5. Verify harness builder unit tests still pass

**File**: `packages/agent/agent-operator-go/internal/builder/` (test files)

Run existing builder tests. The output of `ClaudeCommandBuilder.Command()` must still be:
- command: `["claude"]`
- args include `--permission-mode bypassPermissions`
- args include `--print --prompt <prompt>` when prompt is set

Check `agents.BuildClaudeArgs(nil, AgentExecOptions{Sandbox: true})` returns `["--permission-mode", "bypassPermissions"]` by reading `packages/shared/shared-agent-go/pkg/agents/claude.go` before writing. If the signature differs, adapt accordingly.

## Validation Steps

```bash
# Check shared-agent-go compiles (no changes there, just verifying)
cd packages/shared/shared-agent-go && go build ./...

# Operator build and test
cd packages/agent/agent-operator-go
go mod tidy
go build ./...
go vet ./...
go test ./...

# Integration tests (requires KUBEBUILDER_ASSETS)
# go test -tags=integration ./test/integration/
```

## Success Criteria

- [ ] `harness_claude.go` uses `agents.GetAgent(types.AgentClaude)` for binary name
- [ ] `harness_claude.go` uses `agents.BuildClaudeArgs()` for base arguments
- [ ] `harness_opencode.go` uses `agents.GetAgent(types.AgentOpencode)` for binary name
- [ ] `harness_opencode.go` uses `agents.BuildOpencodeArgs()` for arguments
- [ ] No hardcoded `"claude"` or `"opencode"` binary strings in harness builders
- [ ] All existing builder tests pass
- [ ] Operator binary compiles cleanly

## Files Modified/Created

- `packages/agent/agent-operator-go/go.mod` (modified — add shared-agent-go require)
- `packages/agent/agent-operator-go/internal/builder/harness_claude.go` (modified)
- `packages/agent/agent-operator-go/internal/builder/harness_opencode.go` (modified)

## Estimated Duration

Small — ~1 hour
