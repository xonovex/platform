# Shared Agent Go

Use this shared Go library to build agent infrastructure for [agent-cli-go](../../tooling/cli/agent-cli-go) and [agent-operator-go](../../sandbox/operator/agent-operator-go).

## Packages

Import the package that owns the required agent concern.

| Package | Description |
| --- | --- |
| `pkg/agents` | Harness facts as embedded JSON, one `harnesses/<name>.json` per harness under the schema `https://schemas.xonovex.com/agent/harness.schema.json`; overrides from `harnesses.d/<name>.json`, and the command builders that read the facts |
| `pkg/network` | Network method parsing and the checks of the managed method |
| `pkg/config` | Configuration loading and validation |
| `pkg/nix` | Nix package sets, defaults, pins, and expansion |
| `pkg/providers` | Model provider files, catalog with `extends` and model slot overrides, and environment building; ships no provider and no model id |
| `pkg/types` | Shared type definitions |
| `pkg/validation` | Repository URL, branch, and commit validation |
| `pkg/worktree` | Worktree naming and VCS type constants (git, jj) |
