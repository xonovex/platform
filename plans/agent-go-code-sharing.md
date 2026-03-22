---
type: plan
has_subplans: true
status: complete
proposed_subplans:
  - shared-agent-go-expand
  - shared-core-go-shell-utils
  - operator-consume-shared-agents
  - operator-consume-shared-providers
  - operator-consume-shared-nix
  - cli-consume-shared-validation
  - cli-add-jj-support
dependencies:
  plans: []
  subplans:
    - plans/agent-go-code-sharing/subplan-01-shared-core-go-shell-utils.md
    - plans/agent-go-code-sharing/subplan-02-shared-agent-go-expand.md
    - plans/agent-go-code-sharing/subplan-03-operator-consume-shared-agents.md
    - plans/agent-go-code-sharing/subplan-04-operator-consume-shared-providers.md
    - plans/agent-go-code-sharing/subplan-05-operator-consume-shared-nix.md
    - plans/agent-go-code-sharing/subplan-06-cli-consume-shared-validation.md
    - plans/agent-go-code-sharing/subplan-07-cli-add-jj-support.md
parallel_groups:
  - group: 1
    parallel: true
    subplans:
      - plans/agent-go-code-sharing/subplan-01-shared-core-go-shell-utils.md
      - plans/agent-go-code-sharing/subplan-02-shared-agent-go-expand.md
  - group: 2
    parallel: true
    depends_on: [1]
    subplans:
      - plans/agent-go-code-sharing/subplan-03-operator-consume-shared-agents.md
      - plans/agent-go-code-sharing/subplan-04-operator-consume-shared-providers.md
      - plans/agent-go-code-sharing/subplan-05-operator-consume-shared-nix.md
  - group: 3
    parallel: true
    depends_on: [1]
    subplans:
      - plans/agent-go-code-sharing/subplan-06-cli-consume-shared-validation.md
      - plans/agent-go-code-sharing/subplan-07-cli-add-jj-support.md
skills_to_consult:
  - skill-general-fp
research_sources:
  documentation: []
  versions:
    go: "1.25.5"
    controller-runtime: "0.20.4"
    cobra: "1.8.1"
---

# Agent Go Code Sharing

Unify agent-cli-go and agent-operator-go by extracting shared logic into shared-agent-go and shared-core-go, so both consumers benefit from each other's capabilities.

## Goals

- Operator consumes shared-agent-go for agent command building, provider definitions, and nix package sets instead of maintaining its own implementations
- CLI gains input validation, jj (Jujutsu) VCS support, and policy awareness from operator concepts
- Shared libraries expand to cover all common agent infrastructure (types, commands, providers, nix, validation, shell utils)
- No behavioral regressions in either consumer — existing tests continue to pass

## Current State

- **agent-cli-go**: Heavily uses shared-agent-go (agents, providers, config, worktree) and shared-core-go (logging). Has full nix environment system, 5 sandbox methods, tmux wrapper, git worktree support.
- **agent-operator-go**: Entirely self-contained. Declares shared-agent-go and shared-core-go in go.mod with `replace` directives but imports neither. Has its own agent command builders, provider resolution, nix toolchain, workspace VCS strategies (git + jj), validation, and shell utilities.
- **shared-agent-go**: Agent types, command building (claude/opencode), provider registry, config loading, worktree naming.
- **shared-core-go**: Terminal colors, logging, scriptlib facade. No external dependencies.

## Research Findings

### Duplicated Code

| Concept | CLI Location | Operator Location | Divergence |
|---------|-------------|-------------------|------------|
| Agent command args | `shared-agent-go/pkg/agents/` | `builder/harness_claude.go`, `harness_opencode.go` | Operator adds K8s-specific flags |
| Agent type constants | `shared-agent-go/pkg/types/agent.go` | Hardcoded strings in CRD types | Identical values |
| Provider env vars | `shared-agent-go/pkg/providers/` | `resolver/provider.go` | Different resolution (registry vs K8s Secret) |
| Nix packages | `internal/nixenv/types.go` (package sets, defaults) | `builder/toolchain_nix.go` (simple install) | CLI has full build system; operator has init container |
| Shell quoting | `internal/wrapper/tmux/tmux.go` (`shellQuote`) | `builder/shell.go` (`ShellQuote`) | Nearly identical |
| Worktree naming | `shared-agent-go/pkg/worktree/naming.go` | `builder/workspace*.go` (inline) | Operator has git + jj strategies |
| Repo validation | None | `validator/repository.go` | CLI lacks validation entirely |

### Design Principle

**Share definitions, not mechanisms.** The *what* (agent types, provider specs, nix package sets, validation rules) is shared. The *how* (K8s Secrets vs file config, init containers vs local nix-build, RWX PVCs vs local worktrees) stays in each consumer.

### Alternatives Considered

1. **Merge into single binary** — Rejected. CLI and operator have fundamentally different deployment models (local binary vs K8s controller).
2. **Keep fully separate** — Rejected. Too much drift; bug fixes in command building must be applied twice.
3. **Shared library expansion (chosen)** — Best balance: shared definitions with consumer-specific adapters.

## Proposed Approach

### Phase 1: Expand Shared Libraries

1. **shared-core-go**: Add `pkg/shell/` with `ShellQuote()` and related utilities
2. **shared-agent-go**: Add `pkg/nix/` with package set definitions, default packages, validation
3. **shared-agent-go**: Add `pkg/validation/` with repository URL, branch, commit validation (from operator)
4. **shared-agent-go**: Expand `pkg/worktree/` with VCS strategy types (git + jj)

### Phase 2: Operator Consumes Shared Libraries

5. **Operator harness builders**: Refactor `harness_claude.go` / `harness_opencode.go` to use `shared-agent-go/agents` for base command construction, wrapping with K8s-specific additions
6. **Operator provider resolution**: Use `shared-agent-go/providers` for provider definitions; keep K8s Secret resolution in operator's resolver
7. **Operator nix toolchain**: Use `shared-agent-go/nix` for package set definitions; keep init container logic in operator

### Phase 3: CLI Gains Operator Capabilities

8. **CLI validation**: Use `shared-agent-go/validation` for worktree branch names, repo URLs
9. **CLI jj support**: Add Jujutsu worktree/workspace support using shared VCS strategy types

## Risk Assessment

- **Operator test breakage**: The operator has extensive e2e tests (Kind, gVisor, Kata, CoCo). Refactoring builders to use shared code could introduce subtle differences. Mitigation: run full test suite after each change; keep K8s-specific wrappers thin.
- **API surface growth in shared-agent-go**: Adding nix, validation, and expanded worktree packages increases the shared library's surface area. Mitigation: keep packages focused and independently testable.
- **Versioning coordination**: Both consumers use `replace` directives pointing to local paths, so there's no semver concern within the monorepo. Risk is low.
- **controller-gen incompatibility**: Operator README notes controller-gen is broken with Go 1.25+. Shared code must not require code generation. Mitigation: shared packages are pure Go libraries, no CRD types.

## Proposed Child Plans

### Group 1 — Shared Library Expansion (parallel)

- **shared-core-go-shell-utils**: Add `pkg/shell/` with `ShellQuote` and shell utility functions
- **shared-agent-go-expand**: Add `pkg/nix/`, `pkg/validation/`, expand `pkg/worktree/` with VCS strategy types

### Group 2 — Operator Integration (parallel, depends on Group 1)

- **operator-consume-shared-agents**: Refactor operator harness builders to use shared-agent-go agents package
- **operator-consume-shared-providers**: Refactor operator provider resolution to use shared provider definitions
- **operator-consume-shared-nix**: Refactor operator nix toolchain to use shared nix package definitions

### Group 3 — CLI Enhancement (parallel, depends on Group 1)

- **cli-consume-shared-validation**: Add input validation to CLI using shared validation package
- **cli-add-jj-support**: Add Jujutsu VCS support to CLI worktree management

## Success Criteria

- [ ] Operator imports and uses shared-agent-go for agent command building
- [ ] Operator imports and uses shared-agent-go for provider definitions
- [ ] Operator imports and uses shared-agent-go for nix package sets
- [ ] CLI validates worktree inputs using shared validation
- [ ] Shell quoting lives in shared-core-go, consumed by both
- [ ] All existing tests pass in both agent-cli-go and agent-operator-go
- [ ] No duplicate agent type constants, command building logic, or nix package definitions

## Estimated Effort

- Group 1 (shared libraries): Small-medium — extracting and organizing existing code
- Group 2 (operator integration): Medium — refactoring builders/resolvers with careful testing
- Group 3 (CLI enhancement): Small — consuming new shared packages
