---
type: plan
has_subplans: false
parent_plan: plans/agent-orthogonal-axis-reorg.md
parallel_group: 2
status: pending
dependencies:
  plans: [01-shared-per-axis-split.md]
  files:
    - packages/agent/agent-cli-go/internal/worktree/
    - packages/agent/agent-cli-go/internal/wrapper/
    - packages/agent/agent-cli-go/internal/workspace/
    - packages/agent/agent-cli-go/internal/terminal/
skills_to_consult: [orthogonal-pattern-guide, hexagonal-pattern-guide, general-fp-guide, moon-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# CLI Workspace + Terminal Axes

## Objective

Relocate the CLI's VCS-checkout and terminal-output mechanisms into two new orthogonal axes — `internal/workspace/{shared,git}` and `internal/terminal/{shared,none,tmux}` — each a `shared/` port-plus-registry core with bare per-type leaves. The `TerminalExecutor` interface and the CLI-only filesystem naming helpers (deleted from shared in subplan 01) land in their owning axis, and `VCSType` is consumed from shared `pkg/workspace` rather than redefined. This is a pure relocation/renaming job under two disjoint directory trees; flag wiring stays in subplan 04.

## Tasks

1. **Create the workspace axis shared core + git leaf.** Split `internal/worktree/worktree.go` into `internal/workspace/shared/workspace.go` (the VCS-agnostic `Setup`/`BuildBindPaths` plus the `VCS` port) and `internal/workspace/git/git.go` (the git-worktree leaf implementing the port). Do NOT fabricate a `jj` leaf — there is no `jj` CLI implementation yet (Decision 5). The port lives in `shared/` per hexagonal; the leaf depends on `shared/`, never the reverse.

   ```go
   // internal/workspace/shared/workspace.go
   package shared

   // VCS is the workspace port; one leaf (git) realizes it today.
   type VCS interface {
       Checkout(ctx context.Context, opts Options) (Checkout, error)
   }

   // Setup and BuildBindPaths stay VCS-agnostic: they take a VCS value.
   func Setup(ctx context.Context, vcs VCS, opts Options) (Checkout, error) { /* ... */ }
   func BuildBindPaths(c Checkout) []string { /* ... */ }
   ```

2. **Move the CLI-only filesystem naming helpers into the workspace axis.** `SanitizeBranchName` / `GetDefaultDir` (formerly shared `pkg/worktree/naming.go`, evicted in subplan 01) generate on-disk directory names — a CLI concern — so they belong in the workspace axis. Place them in `internal/workspace/shared/naming.go` (VCS-agnostic, used by `Setup`) or in the git leaf if git-specific.

   ```go
   // internal/workspace/shared/naming.go
   package shared

   func SanitizeBranchName(branch string) string { /* unchanged logic */ }
   func GetDefaultDir(repo, branch string) string { /* unchanged logic */ }
   ```

3. **Rename the terminal-output axis directory.** Move `internal/wrapper/registry.go` and `internal/wrapper/tmux/` into `internal/terminal/{shared,none,tmux}/`: the registry + port go to `internal/terminal/shared/`, the tmux leaf to `internal/terminal/tmux/`, and add a bare `internal/terminal/none/` direct-exec leaf as the registry default. Keep the lazy-factory registry shape (`map[Method]Factory`, no global state) per microkernel.

   ```go
   // internal/terminal/shared/registry.go
   package shared

   type Factory func() TerminalExecutor
   type Registry map[Method]Factory
   ```

4. **Relocate the `TerminalExecutor` interface out of shared into the terminal axis.** It was shared `pkg/types/terminal.go` (deleted in subplan 01) and has exactly one consumer — the CLI; the operator never implements it. Define it in `internal/terminal/shared/terminal.go` as the axis port; the `none` and `tmux` leaves implement it.

   ```go
   // internal/terminal/shared/terminal.go
   package shared

   type TerminalExecutor interface {
       Run(ctx context.Context, command []string) int
   }
   ```

5. **Consume `VCSType` from shared `pkg/workspace`, not a local redefinition.** Import the `VCSType` enum created in subplan 01 (`shared-agent-go/pkg/workspace`) and map the CLI's `WorkspaceType` onto it in `internal/workspace/shared`; delete any leftover local `VCSType` declaration so the enum has a single owner (connascence: one definition site).

   ```go
   import wsp "github.com/.../shared-agent-go/pkg/workspace"

   // WorkspaceType maps the CLI selector onto the shared enum.
   func (t WorkspaceType) toVCS() wsp.VCSType { /* git -> wsp.VCSGit */ }
   ```

6. **Repoint imports in `cmd/` and tests from the old paths to the new axes.** Replace every `internal/worktree` import with `internal/workspace/{shared,git}` and every `internal/wrapper` import with `internal/terminal/{shared,none,tmux}`; delete the now-empty `internal/worktree/` and `internal/wrapper/` directories. Leave flag declarations to subplan 04 — only fix the import paths and constructor call sites here.

7. **Confirm moon project wiring still resolves the renamed dirs.** Run the `agent-cli-go` moon tasks so the new `internal/workspace` and `internal/terminal` trees are typechecked/built; no `moon.yml` source-glob edits are expected (internal Go packages are covered by the module-wide globs), but verify rather than assume per moon-guide.

## Validation Steps

- `npx moon run agent-cli-go:typecheck`
- `npx moon run agent-cli-go:lint`
- `npx moon run agent-cli-go:build`
- `npx moon run agent-cli-go:test`
- In `packages/agent/agent-cli-go`: `go build ./...` and `go test ./...` (confirms no dangling `internal/worktree` / `internal/wrapper` imports and that the workspace/terminal packages compile).
- `go vet ./internal/workspace/... ./internal/terminal/...`
- Integration help-text tests are NOT expected to change in this subplan (flags move in subplan 04); if any `test/integration` reference points at the old package paths, repoint the import only — do not alter expected help output here.

## Success Criteria

- [ ] `internal/workspace/shared/` holds the `VCS` port + `Setup`/`BuildBindPaths` + naming helpers; `internal/workspace/git/` holds the sole leaf; no `jj` leaf fabricated.
- [ ] `internal/terminal/{shared,none,tmux}/` exist; `shared/` holds the `TerminalExecutor` port + lazy-factory registry; `none` and `tmux` are bare leaves.
- [ ] `internal/worktree/` and `internal/wrapper/` directories are deleted (no backwards-compat shims).
- [ ] `VCSType` is consumed from shared `pkg/workspace`; no local redefinition remains.
- [ ] `shared/` never imports its sibling leaves; leaves import only `shared/` (one-way dependency).
- [ ] All `agent-cli-go` moon tasks (typecheck/lint/build/test) and `go build/test ./...` pass.

## Files Modified/Created

- Created: `internal/workspace/shared/workspace.go`, `internal/workspace/shared/naming.go`, `internal/workspace/git/git.go`
- Created: `internal/terminal/shared/terminal.go`, `internal/terminal/shared/registry.go`, `internal/terminal/none/none.go`, `internal/terminal/tmux/tmux.go`
- Deleted: `internal/worktree/worktree.go` (and the `internal/worktree/` dir), `internal/wrapper/registry.go`, `internal/wrapper/tmux/` (and the `internal/wrapper/` dir)
- Modified: `cmd/` call sites importing the old `internal/worktree` / `internal/wrapper` paths; any `test/` files importing those paths
- All paths under `packages/agent/agent-cli-go/`.

## Dependencies

- **01-shared-per-axis-split** — must land first: it creates shared `pkg/workspace` (the `VCSType` enum this subplan consumes in Task 5) and evicts the CLI-only `TerminalExecutor` (was `pkg/types/terminal.go`) and worktree-naming helpers from shared, which this subplan rehomes into the new CLI axes (Tasks 2 and 4). Runs in parallel with `02-cli-isolation-provision-network` (parallel_group 2) — the two operate on disjoint directory trees.

## Estimated Duration

Small-to-medium — mechanical relocation across two disjoint trees plus import repointing; no new logic. Roughly 0.5 day including running the full CLI validation suite.
