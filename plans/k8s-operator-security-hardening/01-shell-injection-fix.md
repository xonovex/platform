---
type: plan
has_subplans: false
parent_plan: plans/k8s-operator-security-hardening.md
parallel_group: 1
status: complete
dependencies:
  plans: []
  files:
    - packages/agent/agent-operator-go/internal/webhook/agentrun_webhook.go
    - packages/agent/agent-operator-go/internal/webhook/agentworkspace_webhook.go
    - packages/agent/agent-operator-go/internal/builder/container.go
    - packages/agent/agent-operator-go/internal/builder/workspace.go
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Shell Injection Fix

## Objective

Eliminate the shell injection vulnerability in all clone and worktree scripts. Both `container.go` and `workspace.go` concatenate user-provided `URL`, `Branch`, `Commit`, and worktree branch names directly into shell scripts without sanitization. Fix via webhook input validation (primary) plus shell quoting (defense-in-depth).

## Tasks

### 1. Add input validation helpers in a new `validator` package

Create `packages/agent/agent-operator-go/internal/validator/repository.go`:

```go
package validator

import (
    "fmt"
    "regexp"
)

// repoURL allows http/https and git+ssh URLs
var repoURLPattern = regexp.MustCompile(
    `^(https?://[^\s]+|git@[a-zA-Z0-9._-]+:[a-zA-Z0-9/_.-]+\.git)$`,
)

// branch/ref: letters, digits, ., -, _, /  — no shell metacharacters
var refPattern = regexp.MustCompile(`^[a-zA-Z0-9._/\-]+$`)

// commit SHA: hex, 7-40 chars
var commitPattern = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)

func ValidateRepositoryURL(url string) error {
    if url == "" {
        return fmt.Errorf("repository URL is required")
    }
    if !repoURLPattern.MatchString(url) {
        return fmt.Errorf("repository URL %q contains invalid characters or unsupported scheme", url)
    }
    return nil
}

func ValidateBranch(branch string) error {
    if branch == "" {
        return nil // optional field
    }
    if !refPattern.MatchString(branch) {
        return fmt.Errorf("branch %q contains invalid characters", branch)
    }
    return nil
}

func ValidateCommit(commit string) error {
    if commit == "" {
        return nil // optional field
    }
    if !commitPattern.MatchString(commit) {
        return fmt.Errorf("commit %q must be a 7-40 character hex SHA", commit)
    }
    return nil
}
```

Also create `packages/agent/agent-operator-go/internal/validator/repository_test.go` covering:
- Valid https/git@ URLs pass
- URLs with shell metacharacters (`;`, `|`, `$`, `` ` ``, `&&`) are rejected
- Valid branch names pass
- Branches with metacharacters are rejected
- Valid 40-char SHA and short 7-char SHA pass
- Non-hex commit strings are rejected

### 2. Call validators from AgentRun webhook

In `packages/agent/agent-operator-go/internal/webhook/agentrun_webhook.go`, update `validate()` after the existing mutual-exclusivity checks:

```go
import "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"

func (w *AgentRunWebhook) validate(run *agentv1alpha1.AgentRun) (admission.Warnings, error) {
    // ... existing checks ...

    // Validate inline workspace repository fields
    if run.Spec.Workspace != nil {
        repo := run.Spec.Workspace.Repository
        if err := validator.ValidateRepositoryURL(repo.URL); err != nil {
            return nil, err
        }
        if err := validator.ValidateBranch(repo.Branch); err != nil {
            return nil, err
        }
        if err := validator.ValidateCommit(repo.Commit); err != nil {
            return nil, err
        }
    }

    return nil, nil
}
```

### 3. Call validators from AgentWorkspace webhook

In `packages/agent/agent-operator-go/internal/webhook/agentworkspace_webhook.go`, update the existing `validate()` function to also call the repository validators:

```go
import "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"

func (w *AgentWorkspaceWebhook) validate(ws *agentv1alpha1.AgentWorkspace) (admission.Warnings, error) {
    // ... existing checks ...

    if err := validator.ValidateRepositoryURL(ws.Spec.Repository.URL); err != nil {
        return nil, err
    }
    if err := validator.ValidateBranch(ws.Spec.Repository.Branch); err != nil {
        return nil, err
    }
    if err := validator.ValidateCommit(ws.Spec.Repository.Commit); err != nil {
        return nil, err
    }

    return nil, nil
}
```

### 4. Add shell quoting in buildCloneScript (defense-in-depth)

In `packages/agent/agent-operator-go/internal/builder/container.go`, update `buildCloneScript` to quote all user-supplied values:

```go
func buildCloneScript(repo agentv1alpha1.RepositorySpec, wsType agentv1alpha1.WorkspaceType) string {
    script := "set -e\n"
    script += "cd " + workspaceMountPath + "\n"
    script += "git clone"
    if repo.Branch != "" {
        script += " --branch " + shellQuote(repo.Branch)
    }
    script += " --single-branch --depth 1"
    script += " -- " + shellQuote(repo.URL) + " .\n"

    if repo.Commit != "" {
        script += "git fetch origin " + shellQuote(repo.Commit) + "\n"
        script += "git checkout " + shellQuote(repo.Commit) + "\n"
    }

    if vcs, err := GetVCSStrategy(wsType); err == nil {
        script += vcs.PostCloneScript()
    }

    return script
}
```

Add `shellQuote` helper in `packages/agent/agent-operator-go/internal/builder/shell.go`:

```go
package builder

import "strings"

// shellQuote wraps a string in single quotes, escaping any embedded single quotes.
// This is safe for POSIX sh arguments even if the value contains spaces or special chars.
func shellQuote(s string) string {
    return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
```

### 5. Apply the same quoting in buildWorkspaceCloneScript

In `packages/agent/agent-operator-go/internal/builder/workspace.go`, update `buildWorkspaceCloneScript`:

```go
func buildWorkspaceCloneScript(repo *agentv1alpha1.RepositorySpec, wsType agentv1alpha1.WorkspaceType) string {
    script := "set -e\n"
    script += "cd " + workspaceMountPath + "\n"
    script += "git clone"
    if repo.Branch != "" {
        script += " --branch " + shellQuote(repo.Branch)
    }
    script += " --single-branch --depth 1"
    script += " -- " + shellQuote(repo.URL) + " .\n"

    if repo.Commit != "" {
        script += "git fetch origin " + shellQuote(repo.Commit) + "\n"
        script += "git checkout " + shellQuote(repo.Commit) + "\n"
    }

    if strategy, err := GetVCSStrategy(wsType); err == nil {
        script += strategy.PostCloneScript()
    }

    return script
}
```

Also quote `worktreeBranch` and `sourceBranch` in `BuildWorktreeInitContainers` — these come from `AgentWorkspace.Spec.Git.Worktree` fields. The VCS strategy `WorktreeScript` implementations in `workspace_git.go` and `workspace_jj.go` must also apply `shellQuote` to the branch/path arguments passed in.

### 6. Update existing tests

In `packages/agent/agent-operator-go/internal/builder/container_test.go`:
- Add `TestBuildCloneScript_InjectionRejected` — verify `branch: "main; rm -rf /"` produces a safely quoted script (the webhook prevents it reaching the builder, but the builder also defends itself).

In `packages/agent/agent-operator-go/internal/webhook/agentrun_webhook_test.go`:
- Add tests: `TestValidateCreate_MaliciousBranch`, `TestValidateCreate_MaliciousURL`, `TestValidateCreate_MaliciousCommit` — verify webhook returns errors for metacharacter injection.

## Validation Steps

```bash
cd packages/agent/agent-operator-go
go build ./...
go vet ./...
go test ./internal/validator/...
go test ./internal/builder/...
go test ./internal/webhook/...
golangci-lint run ./...
```

## Success Criteria

- [ ] `validator` package exists with URL, branch, commit regex validators
- [ ] AgentRun webhook calls all three validators for inline workspace repos
- [ ] AgentWorkspace webhook calls all three validators for repository spec
- [ ] `shellQuote` helper exists and is used in all clone/worktree scripts
- [ ] Existing clone script tests still pass
- [ ] New injection tests pass (both validator and builder levels)
- [ ] No regressions in webhook unit tests

## Files Modified/Created

- `internal/validator/repository.go` (new)
- `internal/validator/repository_test.go` (new)
- `internal/builder/shell.go` (new)
- `internal/builder/container.go` — `buildCloneScript`
- `internal/builder/workspace.go` — `buildWorkspaceCloneScript`, `BuildWorktreeInitContainers`
- `internal/builder/workspace_git.go` — `WorktreeScript` quoting
- `internal/builder/workspace_jj.go` — `WorktreeScript` quoting
- `internal/webhook/agentrun_webhook.go` — `validate()`
- `internal/webhook/agentworkspace_webhook.go` — `validate()`
- `internal/builder/container_test.go` — new injection test
- `internal/webhook/agentrun_webhook_test.go` — new injection tests

## Estimated Duration

Small — ~2h
