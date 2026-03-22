---
type: plan
has_subplans: false
parent_plan: plans/agent-go-code-sharing.md
parallel_group: 1
status: complete
dependencies:
  plans: []
  files: []
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# shared-agent-go: Expand with Nix, Validation, VCS Types

Add three new packages to shared-agent-go: `pkg/nix/` (package set definitions extracted from CLI), `pkg/validation/` (repository validation extracted from operator), and expand `pkg/worktree/` with VCS type constants.

## Objective

Centralise definitions that both consumers need:
- Nix package sets, default packages, and nixpkgs pins live in shared-agent-go so CLI and operator reference the same source of truth
- Repository URL/branch/commit validation patterns are shared so the CLI can validate inputs too
- VCS type constants (git, jj) live in shared-agent-go types to avoid string literals in both consumers

## Tasks

### 1. Add `pkg/nix/nix.go` to shared-agent-go

**File**: `packages/shared/shared-agent-go/pkg/nix/nix.go`

Move from `packages/agent/agent-cli-go/internal/nixenv/types.go` — the pure definitions only (no build/resolve/render logic):

```go
package nix

// Pin represents a nixpkgs channel pin
type Pin struct {
    Name string
    Ref  string
}

// Pins maps friendly names to their channel refs
var Pins = map[string]Pin{
    "nixos-24.11":     {Name: "nixos-24.11", Ref: "nixos-24.11"},
    "nixos-unstable":  {Name: "nixos-unstable", Ref: "nixos-unstable"},
    "nixpkgs-unstable":{Name: "nixpkgs-unstable", Ref: "nixpkgs-unstable"},
}

// DefaultPin is the default nixpkgs pin
const DefaultPin = "nixos-unstable"

// DefaultPackages are the default packages for agent environments
var DefaultPackages = []string{
    "nodejs_24", "git", "ripgrep", "fd", "fzf", "jq", "curl", "coreutils", "bash",
}

// PackageSets are predefined collections of packages for common use cases
var PackageSets = map[string][]string{
    "nodejs":     {"nodejs_24", "python312", "gnumake", "gcc", "gnused", "gawk", "binutils"},
    "python":     {"python312", "python312Packages.pip"},
    "go":         {"go"},
    "rust":       {"rustc", "cargo"},
    "kubernetes": {"kubectl", "kubernetes-helm", "k9s"},
    "terraform":  {"terraform", "terragrunt"},
    "docker":     {"docker-client"},
    "aws":        {"awscli2"},
    "gcp":        {"google-cloud-sdk"},
}

// ExpandPackageSets expands any named package sets in the input list,
// returning a deduplicated slice of individual package names.
func ExpandPackageSets(packages []string) []string {
    seen := make(map[string]bool)
    var result []string
    for _, pkg := range packages {
        if set, ok := PackageSets[pkg]; ok {
            for _, p := range set {
                if !seen[p] {
                    seen[p] = true
                    result = append(result, p)
                }
            }
        } else {
            if !seen[pkg] {
                seen[pkg] = true
                result = append(result, pkg)
            }
        }
    }
    return result
}

// ValidatePin returns an error if the pin name is not recognised
func ValidatePin(pin string) error {
    if pin == "" {
        return nil // empty means default
    }
    if _, ok := Pins[pin]; !ok {
        return fmt.Errorf("unknown nixpkgs pin %q; known pins: nixos-24.11, nixos-unstable, nixpkgs-unstable", pin)
    }
    return nil
}
```

Add `"fmt"` import.

### 2. Add `pkg/nix/nix_test.go`

Cover `ExpandPackageSets` (plain packages pass through, named sets expand, unknown names preserved, deduplication) and `ValidatePin` (known pins pass, unknown pin errors, empty string passes).

### 3. Add `pkg/validation/repository.go` to shared-agent-go

**File**: `packages/shared/shared-agent-go/pkg/validation/repository.go`

Move from `packages/agent/agent-operator-go/internal/validator/repository.go` verbatim, changing the package name to `validation`:

```go
package validation
// ... same content as operator's validator/repository.go ...
// ValidateRepositoryURL, ValidateBranch, ValidateCommit, ContainsShellMetachars
```

The operator's `validator` package will then import this and re-export or call through (see subplan-06 for CLI usage).

### 4. Add `pkg/validation/repository_test.go`

Move tests from `packages/agent/agent-operator-go/internal/validator/repository_test.go` into the new shared package. The operator's test file can then be removed (the operator will rely on the shared tests).

### 5. Expand `pkg/worktree/` with VCS type constants

**File**: `packages/shared/shared-agent-go/pkg/worktree/vcs.go` (new)

```go
package worktree

// VCSType represents the type of version control system
type VCSType string

const (
    VCSGit      VCSType = "git"
    VCSJujutsu  VCSType = "jj"
    VCSDefault  VCSType = VCSGit
)

// IsValid returns true if vt is a recognised VCS type
func (vt VCSType) IsValid() bool {
    return vt == VCSGit || vt == VCSJujutsu
}
```

### 6. Update operator's validator package to delegate to shared

**File**: `packages/agent/agent-operator-go/internal/validator/repository.go`

After shared `pkg/validation` exists, replace the operator's validator functions with thin wrappers that call the shared ones, eliminating duplication:

```go
package validator

import "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"

var (
    ValidateRepositoryURL  = validation.ValidateRepositoryURL
    ValidateBranch         = validation.ValidateBranch
    ValidateCommit         = validation.ValidateCommit
    ContainsShellMetachars = validation.ContainsShellMetachars
)
```

All existing call sites in the operator remain unchanged.

### 7. Update CLI's nixenv/types.go to reference shared definitions

**File**: `packages/agent/agent-cli-go/internal/nixenv/types.go`

Replace the local `NixpkgsPins`, `DefaultNixpkgsPin`, `DefaultBasePackages`, `PackageSets` with aliases pointing to the shared package:

```go
package nixenv

import "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/nix"

// NixpkgsPins, DefaultNixpkgsPin, DefaultBasePackages, PackageSets re-exported
// from shared-agent-go for backward compatibility within this package.
var NixpkgsPins     = nix.Pins
const DefaultNixpkgsPin = nix.DefaultPin
var DefaultBasePackages = nix.DefaultPackages
var PackageSets         = nix.PackageSets
```

Keep the CLI-specific types (`EnvSpec`, `ResolvedEnv`, `BuildResult`, `NixSandboxConfig`) in `types.go`.

Update any calls to `nixenv.ExpandPackageSets()` in `build.go` to use `nix.ExpandPackageSets()`.

## Validation Steps

```bash
# shared-agent-go
cd packages/shared/shared-agent-go
go test ./...
go vet ./...

# operator (validator re-export should compile cleanly)
cd packages/agent/agent-operator-go
go build ./...
go test ./...

# CLI (nixenv/types.go alias should compile cleanly)
cd packages/agent/agent-cli-go
go build ./...
go test ./...
```

## Success Criteria

- [ ] `pkg/nix/nix.go` in shared-agent-go with Pins, DefaultPackages, PackageSets, ExpandPackageSets, ValidatePin
- [ ] `pkg/nix/nix_test.go` covers ExpandPackageSets and ValidatePin
- [ ] `pkg/validation/repository.go` in shared-agent-go with all validation functions
- [ ] `pkg/validation/repository_test.go` moved from operator
- [ ] `pkg/worktree/vcs.go` with VCSType, VCSGit, VCSJujutsu constants
- [ ] Operator's `validator/repository.go` delegates to shared (no duplication)
- [ ] CLI's `nixenv/types.go` aliases shared nix definitions
- [ ] All existing tests pass in both consumers

## Files Modified/Created

- `packages/shared/shared-agent-go/pkg/nix/nix.go` (new)
- `packages/shared/shared-agent-go/pkg/nix/nix_test.go` (new)
- `packages/shared/shared-agent-go/pkg/validation/repository.go` (new)
- `packages/shared/shared-agent-go/pkg/validation/repository_test.go` (new)
- `packages/shared/shared-agent-go/pkg/worktree/vcs.go` (new)
- `packages/agent/agent-operator-go/internal/validator/repository.go` (modified)
- `packages/agent/agent-operator-go/internal/validator/repository_test.go` (deleted or emptied)
- `packages/agent/agent-cli-go/internal/nixenv/types.go` (modified)

## Estimated Duration

Medium — ~2 hours
