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
    - packages/shared/shared-agent-go/pkg/nix/nix.go
    - packages/agent/agent-operator-go/go.mod
    - packages/agent/agent-operator-go/internal/builder/toolchain_nix.go
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Operator: Consume shared-agent-go Nix Package Definitions

Refactor `NixToolchain` to use shared nix package set definitions and add package set expansion support, so users can specify `nodejs` instead of listing all its constituent packages.

## Objective

The operator's `toolchain_nix.go` currently takes `NixSpec.Packages` literally and appends `nixpkgs#<pkg>` for each one. The CLI's `internal/nixenv` already has a full `PackageSets` map and `ExpandPackageSets()` function (now in `shared-agent-go/pkg/nix/`). Adding expansion to the operator lets users specify shorthand names like `nodejs`, `rust`, `kubernetes` in their `AgentToolchain` resources.

## Current State

**`internal/builder/toolchain_nix.go:86-97`** (`installScript`):
```go
func (n *NixToolchain) installScript() string {
    var pkgRefs []string
    for _, pkg := range n.nix.Packages {
        pkgRefs = append(pkgRefs, "nixpkgs#"+pkg)
    }
    script := "set -e\n"
    script += "cp -a /nix/. /nix-env/\n"
    script += fmt.Sprintf("nix ... profile install ... %s\n", strings.Join(pkgRefs, " "))
    script += "cp -a /nix/. /nix-env/\n"
    return script
}
```

No package set expansion. No defaults. Packages specified verbatim.

## Tasks

### 1. Import shared nix package in `toolchain_nix.go`

**File**: `packages/agent/agent-operator-go/internal/builder/toolchain_nix.go`

Add import (go.mod already has the replace directive; shared-agent-go require added in subplan-03):
```go
import (
    // ... existing imports ...
    nixdefs "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/nix"
)
```

### 2. Add package set expansion to `installScript`

**File**: `packages/agent/agent-operator-go/internal/builder/toolchain_nix.go`

Replace the literal package iteration with expansion:

```go
func (n *NixToolchain) installScript() string {
    expanded := nixdefs.ExpandPackageSets(n.nix.Packages)
    var pkgRefs []string
    for _, pkg := range expanded {
        pkgRefs = append(pkgRefs, "nixpkgs#"+pkg)
    }
    script := "set -e\n"
    script += "cp -a /nix/. /nix-env/\n"
    script += fmt.Sprintf("nix --extra-experimental-features \"nix-command flakes\" profile install --profile /nix/var/nix/profiles/agent %s\n", strings.Join(pkgRefs, " "))
    script += "cp -a /nix/. /nix-env/\n"
    return script
}
```

Individual package names (not in a set) pass through unchanged, so existing AgentToolchain resources work without modification.

### 3. Update toolchain webhook validation

**File**: `packages/agent/agent-operator-go/internal/webhook/agenttoolchain_webhook.go`

Add validation for package names: after expansion, all resulting package names should be valid nixpkgs attribute names (no shell metacharacters). Use `shell.ContainsMetachars` from shared-core-go (available after subplan-01):

```go
import "github.com/xonovex/platform/packages/shared/shared-core-go/pkg/shell"
import nixdefs "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/nix"

// In ValidateCreate/ValidateUpdate:
for _, pkg := range spec.Nix.Packages {
    if shell.ContainsMetachars(pkg) {
        return fmt.Errorf("package name %q contains invalid characters", pkg)
    }
}
```

### 4. Update toolchain builder tests

**File**: `packages/agent/agent-operator-go/internal/builder/toolchain_nix_test.go` (create or update)

Test cases:
- Plain package names pass through as `nixpkgs#<name>`
- Named package sets (e.g., `"nodejs"`) expand to their constituent packages
- Unknown names are preserved as-is (no error)
- Deduplication: if a package appears in both explicit list and a set, it appears once

### 5. Update sample AgentToolchain to demonstrate package sets

**File**: `packages/agent/agent-operator-go/config/samples/` (relevant toolchain sample)

Add a comment showing the available package sets, so operators know what shorthand names work.

## Validation Steps

```bash
cd packages/agent/agent-operator-go
go build ./...
go vet ./...
go test ./...
# Verify install script output for a toolchain with "nodejs" package set
```

## Success Criteria

- [ ] `toolchain_nix.go` imports `pkg/nix` from shared-agent-go
- [ ] `ExpandPackageSets` is called before building the install script
- [ ] Named sets (`nodejs`, `go`, `rust`, etc.) expand correctly
- [ ] Individual package names work identically to before
- [ ] Webhook validates package names for shell metacharacters
- [ ] Toolchain builder tests cover set expansion
- [ ] Operator binary compiles and all tests pass

## Files Modified/Created

- `packages/agent/agent-operator-go/internal/builder/toolchain_nix.go` (modified)
- `packages/agent/agent-operator-go/internal/builder/toolchain_nix_test.go` (new or modified)
- `packages/agent/agent-operator-go/internal/webhook/agenttoolchain_webhook.go` (modified)
- `packages/agent/agent-operator-go/config/samples/` (one sample updated)

## Estimated Duration

Small — ~1 hour
