---
type: plan
has_subplans: false
parent_plan: plans/agent-go-code-sharing.md
parallel_group: 3
status: complete
dependencies:
  plans:
    - plans/agent-go-code-sharing/subplan-02-shared-agent-go-expand.md
  files:
    - packages/shared/shared-agent-go/pkg/validation/repository.go
    - packages/agent/agent-cli-go/internal/cmd/run.go
    - packages/agent/agent-cli-go/internal/worktree/worktree.go
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# CLI: Add Input Validation Using shared-agent-go Validation

Add validation to the CLI's `run` command for worktree branch names and repository URLs, using the validation functions extracted to `shared-agent-go/pkg/validation`.

## Objective

The CLI currently accepts any string for `--worktree-branch`, `--worktree-dir`, and related inputs with no validation. A malformed branch name can cause confusing git errors deep in execution. Using the shared validation package from subplan-02 adds early, clear error messages.

## Current State

**`internal/cmd/run.go`**: Parses `--worktree` flag (branch name), `--worktree-dir`, `--bind`, `--ro-bind` paths. No input validation beyond cobra flag parsing.

**`internal/worktree/worktree.go`**: `Setup()` calls git directly; errors from git are surfaced but not pre-validated.

**`shared-agent-go/pkg/validation/repository.go`** (created in subplan-02):
- `ValidateBranch(branch string) error`
- `ValidateRepositoryURL(url string) error`
- `ValidateCommit(commit string) error`
- `ContainsShellMetachars(s string) bool`

## Tasks

### 1. Validate worktree branch name in `run.go`

**File**: `packages/agent/agent-cli-go/internal/cmd/run.go`

After parsing the `--worktree` flag value (branch name), add validation before the worktree setup call:

```go
import "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"

// In the run command, after parsing worktreeBranch:
if worktreeBranch != "" {
    if err := validation.ValidateBranch(worktreeBranch); err != nil {
        return fmt.Errorf("invalid --worktree branch: %w", err)
    }
}
```

### 2. Validate source branch in `run.go`

**File**: `packages/agent/agent-cli-go/internal/cmd/run.go`

If `--worktree-source-branch` (or equivalent flag) is accepted, apply the same validation:

```go
if sourceBranch != "" {
    if err := validation.ValidateBranch(sourceBranch); err != nil {
        return fmt.Errorf("invalid source branch: %w", err)
    }
}
```

### 3. Validate custom env var values for shell metacharacters

**File**: `packages/agent/agent-cli-go/internal/sandboxutil/utils.go`

In `ParseCustomEnv`, validate each value does not contain unescaped shell metacharacters that could cause injection:

```go
import "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"

// After splitting key=value:
if validation.ContainsShellMetachars(value) {
    // Log a warning but do not reject (env values legitimately may contain
    // special chars like $ or " in controlled contexts)
    scriptlib.LogWarning(fmt.Sprintf("custom env value for %q contains shell metacharacters", key))
}
```

Note: This is a warning, not an error, since env values in controlled sandbox contexts can legitimately contain special characters.

### 4. Add validation tests to CLI integration tests

**File**: `packages/agent/agent-cli-go/test/integration/run_test.go`

Add test cases:
- `--worktree` with invalid branch name (e.g., `branch;rm -rf /`) exits with non-zero and prints a clear error
- `--worktree` with valid branch name (e.g., `feature/my-work`) proceeds normally

## Validation Steps

```bash
cd packages/agent/agent-cli-go
go build ./...
go vet ./...
go test ./...

# Manual smoke test:
# ./dist/agent-cli run --worktree 'bad;branch' claude
# Should exit with: "invalid --worktree branch: branch..."
```

## Success Criteria

- [ ] Worktree branch validated with `validation.ValidateBranch()` before git operations
- [ ] Source branch validated similarly
- [ ] Invalid branch name produces clear error message before any git calls
- [ ] Custom env var values log a warning if shell metacharacters are present
- [ ] Integration tests cover invalid branch name rejection
- [ ] All existing CLI tests pass

## Files Modified/Created

- `packages/agent/agent-cli-go/internal/cmd/run.go` (modified)
- `packages/agent/agent-cli-go/internal/sandboxutil/utils.go` (modified)
- `packages/agent/agent-cli-go/test/integration/run_test.go` (modified)

## Estimated Duration

Small — ~1 hour
