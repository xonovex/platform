# Shared Agent Go

Shared Go library for agent infrastructure, used by [agent-cli-go](../../agent/agent-cli-go/) and [agent-operator-go](../../agent/agent-operator-go/).

## Packages

| Package          | Description                                        |
| ---------------- | -------------------------------------------------- |
| `pkg/agents`     | Agent type definitions and command building         |
| `pkg/config`     | Configuration loading and validation                |
| `pkg/nix`        | Nix package sets, defaults, pins, and expansion     |
| `pkg/providers`  | Model provider registry and environment building    |
| `pkg/types`      | Shared type definitions                             |
| `pkg/validation` | Repository URL, branch, and commit validation       |
| `pkg/worktree`   | Worktree naming and VCS type constants (git, jj)    |
