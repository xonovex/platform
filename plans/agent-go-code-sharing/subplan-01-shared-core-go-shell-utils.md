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

# shared-core-go: Add Shell Utilities

Add `pkg/shell/` to shared-core-go with `ShellQuote` and `ContainsShellMetachars`, expose via scriptlib, then migrate both consumers off their local copies.

## Objective

Both the operator (`internal/builder/shell.go`) and CLI (`internal/wrapper/tmux/tmux.go`) implement local unexported versions of POSIX shell quoting. Extract into shared-core-go so both can import a single tested implementation.

## Tasks

### 1. Create `pkg/shell/shell.go` in shared-core-go

**File**: `packages/shared/shared-core-go/pkg/shell/shell.go`

```go
package shell

import "strings"

// Quote wraps s in single quotes, escaping any embedded single quotes.
// Safe for POSIX sh arguments even if the value contains spaces or special chars.
func Quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// ContainsMetachars returns true if s contains POSIX shell metacharacters
// that could enable command injection.
func ContainsMetachars(s string) bool {
	const metachars = ";|&$`\\\"'<>(){}!#~\n\r"
	return strings.ContainsAny(s, metachars)
}
```

### 2. Create `pkg/shell/shell_test.go` in shared-core-go

**File**: `packages/shared/shared-core-go/pkg/shell/shell_test.go`

Cover: plain string, string with spaces, single quotes, double quotes, semicolons, empty string. Mirror the pattern in `pkg/colors/colors_test.go`.

### 3. Add shell re-exports to scriptlib

**File**: `packages/shared/shared-core-go/pkg/scriptlib/scriptlib.go`

Add to the var block:
```go
var (
    // ... existing exports ...
    ShellQuote         = shell.Quote
    ShellContainsMetachars = shell.ContainsMetachars
)
```

Add import: `"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/shell"`

### 4. Migrate operator: replace local `builder/shell.go`

**File**: `packages/agent/agent-operator-go/internal/builder/shell.go`

Replace file contents with an import alias for ergonomic internal use:
```go
package builder

import "github.com/xonovex/platform/packages/shared/shared-core-go/pkg/shell"

// shellQuote is a package-local alias for shell.Quote.
var shellQuote = shell.Quote
```

This keeps all call sites in builder (`workspace_git.go`, `workspace_jj.go`) unchanged.

### 5. Migrate operator: replace local `validator/repository.go` `ContainsShellMetachars`

**File**: `packages/agent/agent-operator-go/internal/validator/repository.go`

Remove the local `ContainsShellMetachars` function and its `strings` import.
Replace call sites with `shell.ContainsMetachars` from the shared package.

Add import: `"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/shell"`

### 6. Migrate CLI: replace local `shellQuote` in tmux

**File**: `packages/agent/agent-cli-go/internal/wrapper/tmux/tmux.go`

Remove the local unexported `shellQuote` function (currently defined inline in the file).
Replace all call sites with `shell.Quote` from shared-core-go.

Add import: `"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/shell"`

## Validation Steps

```bash
# shared-core-go
cd packages/shared/shared-core-go
go test ./...
go vet ./...

# operator
cd packages/agent/agent-operator-go
go build ./...
go test ./...
go vet ./...

# CLI
cd packages/agent/agent-cli-go
go build ./...
go test ./...
go vet ./...
```

## Success Criteria

- [ ] `pkg/shell/shell.go` exists in shared-core-go with `Quote` and `ContainsMetachars`
- [ ] `pkg/shell/shell_test.go` tests both functions
- [ ] scriptlib re-exports `ShellQuote` and `ShellContainsMetachars`
- [ ] Operator `builder/shell.go` is a thin alias, not a reimplementation
- [ ] `ContainsShellMetachars` removed from `validator/repository.go`
- [ ] CLI `tmux.go` uses `shell.Quote` from shared-core-go
- [ ] All existing tests pass

## Files Modified/Created

- `packages/shared/shared-core-go/pkg/shell/shell.go` (new)
- `packages/shared/shared-core-go/pkg/shell/shell_test.go` (new)
- `packages/shared/shared-core-go/pkg/scriptlib/scriptlib.go` (modified)
- `packages/agent/agent-operator-go/internal/builder/shell.go` (modified)
- `packages/agent/agent-operator-go/internal/validator/repository.go` (modified)
- `packages/agent/agent-cli-go/internal/wrapper/tmux/tmux.go` (modified)

## Estimated Duration

Small — ~1 hour
