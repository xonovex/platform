---
type: plan
has_subplans: false
parent_plan: plans/agent-go-code-sharing.md
parallel_group: 3
status: complete
dependencies:
  plans:
    - plans/agent-go-code-sharing/subplan-02-shared-agent-go-expand.md
    - plans/agent-go-code-sharing/subplan-06-cli-consume-shared-validation.md
  files:
    - packages/shared/shared-agent-go/pkg/worktree/vcs.go
    - packages/agent/agent-cli-go/internal/worktree/worktree.go
    - packages/agent/agent-cli-go/internal/cmd/run.go
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# CLI: Add Jujutsu (jj) VCS Support

Add Jujutsu workspace support to the CLI's worktree management, using the `VCSType` constants from shared-agent-go and mirroring the operator's `JujutsuStrategy` for the local execution context.

## Objective

The operator supports both git worktrees and Jujutsu (`jj`) workspaces via its VCS strategy pattern. The CLI currently only supports git worktrees. Adding `--vcs jj` to the CLI's `run` command enables users to use jj workspaces locally, consistent with how the operator handles jj in K8s.

**Key jj concepts** (from operator's `workspace_jj.go`):
- After cloning, run `jj git init --colocate` to colocate jj on top of git
- Create workspaces with `jj workspace add <path> --revision <source>`
- jj workspaces are analogous to git worktrees but use jj's change-based model

## Current State

**`internal/worktree/worktree.go`**: `Setup()` uses `git worktree add` exclusively. No VCS abstraction.

**`internal/cmd/run.go`**: No `--vcs` flag.

**`shared-agent-go/pkg/worktree/vcs.go`** (created in subplan-02):
```go
type VCSType string
const (
    VCSGit     VCSType = "git"
    VCSJujutsu VCSType = "jj"
    VCSDefault VCSType = VCSGit
)
```

**Operator `builder/workspace_jj.go`** (reference implementation):
```go
func (j *JujutsuStrategy) PostCloneScript() string {
    return "jj git init --colocate\n"
}
func (j *JujutsuStrategy) WorktreeScript(path, _, sourceBranch string) string {
    return fmt.Sprintf("jj workspace add %s --revision %s\n", shellQuote(path), shellQuote(sourceBranch))
}
```

## Tasks

### 1. Add `--vcs` flag to `run.go`

**File**: `packages/agent/agent-cli-go/internal/cmd/run.go`

```go
var vcs string
cmd.Flags().StringVar(&vcs, "vcs", "git", "VCS type for worktree (git, jj)")
```

Pass the VCS type through to the worktree setup call.

### 2. Refactor `worktree/worktree.go` to support VCS strategies

**File**: `packages/agent/agent-cli-go/internal/worktree/worktree.go`

Add `VCSType` to `Config`:

```go
import "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/worktree"

type Config struct {
    SourceBranch string
    Branch       string
    Dir          string
    VCS          worktree.VCSType // default: VCSGit
}
```

### 3. Add `IsJJAvailable()` helper

**File**: `packages/agent/agent-cli-go/internal/worktree/worktree.go`

```go
// IsJJAvailable returns true if the jj binary is on PATH
func IsJJAvailable() bool {
    _, err := exec.LookPath("jj")
    return err == nil
}
```

### 4. Add `SetupJJ()` function

**File**: `packages/agent/agent-cli-go/internal/worktree/worktree.go`

Mirror the operator's jj strategy for local execution:

```go
// SetupJJ creates or reuses a jj workspace.
// Assumes the current directory is a git repo colocated with jj
// (i.e., `jj git init --colocate` has already been run).
func SetupJJ(config Config, repoDir string, verbose bool) (string, error) {
    resolvedDir := config.Dir
    if !filepath.IsAbs(config.Dir) {
        resolvedDir = filepath.Join(repoDir, config.Dir)
    }

    // Check if workspace already exists
    if _, err := os.Stat(resolvedDir); err == nil {
        if verbose {
            scriptlib.LogInfo(fmt.Sprintf("Reusing existing jj workspace at %s", config.Dir))
        }
        return resolvedDir, nil
    }

    // Ensure jj is available
    if !IsJJAvailable() {
        return "", fmt.Errorf("jj is not installed or not on PATH; install from https://martinvonz.github.io/jj/")
    }

    sourceBranch := config.SourceBranch
    if sourceBranch == "" {
        sourceBranch = GetCurrentBranchSync(repoDir)
        if sourceBranch == "" {
            return "", fmt.Errorf("failed to determine source revision")
        }
    }

    if verbose {
        scriptlib.LogInfo(fmt.Sprintf("Creating jj workspace at %s from %s", config.Dir, sourceBranch))
    }

    cmd := exec.Command("jj", "workspace", "add", resolvedDir, "--revision", sourceBranch)
    cmd.Dir = repoDir
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("jj workspace add failed: %w", err)
    }

    if verbose {
        scriptlib.LogSuccess("jj workspace created successfully")
    }
    return resolvedDir, nil
}
```

### 5. Update `Setup()` to dispatch by VCS type

**File**: `packages/agent/agent-cli-go/internal/worktree/worktree.go`

```go
func Setup(config Config, repoDir string, verbose bool) (string, error) {
    vcs := config.VCS
    if vcs == "" {
        vcs = worktree.VCSGit
    }
    switch vcs {
    case worktree.VCSJujutsu:
        return SetupJJ(config, repoDir, verbose)
    default:
        return setupGit(config, repoDir, verbose)
    }
}
```

Rename the existing `Setup()` body to `setupGit()` (unexported).

### 6. Update `run.go` to pass VCS type

**File**: `packages/agent/agent-cli-go/internal/cmd/run.go`

Validate `--vcs` value using `worktree.VCSType`:

```go
import sharedworktree "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/worktree"

vcsType := sharedworktree.VCSType(vcs)
if !vcsType.IsValid() {
    return fmt.Errorf("unknown --vcs %q; valid values: git, jj", vcs)
}

// Pass to worktree config:
wtConfig := worktree.Config{
    Branch:       worktreeBranch,
    Dir:          worktreeDir,
    SourceBranch: worktreeSourceBranch,
    VCS:          vcsType,
}
```

### 7. Add worktree tests for jj path

**File**: `packages/agent/agent-cli-go/internal/worktree/` (new test file or extend existing)

Test cases:
- `SetupJJ` returns error when jj is not available
- `Setup` with `VCSJujutsu` dispatches to `SetupJJ`
- `Setup` with `VCSGit` (default) dispatches to `setupGit` (existing behaviour)
- Invalid `VCSType` from CLI produces clear error

### 8. Update integration tests

**File**: `packages/agent/agent-cli-go/test/integration/run_test.go`

Add test: `--vcs unknown` exits with non-zero and helpful message.

## Validation Steps

```bash
cd packages/agent/agent-cli-go
go build ./...
go vet ./...
go test ./...

# Manual smoke test (if jj is installed):
# ./dist/agent-cli run --vcs jj --worktree my-feature claude
```

## Success Criteria

- [ ] `--vcs` flag accepted by `run` command, defaults to `git`
- [ ] Unknown `--vcs` value produces clear error before execution
- [ ] `VCSType` from `shared-agent-go/pkg/worktree` used for constants
- [ ] `Setup()` dispatches to `SetupJJ()` for `jj` VCS type
- [ ] `SetupJJ()` uses `jj workspace add` and checks for jj availability
- [ ] Existing git worktree behaviour unchanged
- [ ] Tests cover VCS dispatch and jj-unavailable error
- [ ] All existing tests pass

## Files Modified/Created

- `packages/agent/agent-cli-go/internal/cmd/run.go` (modified)
- `packages/agent/agent-cli-go/internal/worktree/worktree.go` (modified)
- `packages/agent/agent-cli-go/test/integration/run_test.go` (modified)

## Estimated Duration

Small-medium — ~2 hours
