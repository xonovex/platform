# Shared Agent Go

Use this shared Go library to build agent infrastructure for [agent-cli-go](../../tooling/cli/agent-cli-go) and [agent-operator-go](../../sandbox/operator/agent-operator-go).

## Packages

Import the package that owns the required agent concern.

| Package | Description |
| --- | --- |
| `pkg/agentcmd` | The environment plan that a provider route gives an agent, with no credential value |
| `pkg/agents` | Harness facts as embedded JSON, one `harnesses/<name>.json` per harness under the schema `https://schemas.xonovex.com/agent/harness.schema.json`; overrides from `harnesses.d/<name>.json`, and the command builders that read the facts |
| `pkg/artifact` | The neutral model of the skills, agents, commands and instructions that a plugin or an overlay source gives an agent |
| `pkg/egress` | The deny-by-default egress allowlist of a managed run |
| `pkg/isolation` | Isolation method names and the check of a digest-pinned image |
| `pkg/network` | Network method parsing and the checks of the managed method |
| `pkg/overlay` | The overlay dialect of each agent |
| `pkg/plugin` | Plugin manifests, the capabilities that need a grant, the two marketplace formats and the lock file |
| `pkg/policy` | Sandbox guarantees and `EnforcePolicy`, which checks them fail-closed |
| `pkg/providers` | Model provider files, catalog with `extends` and model slot overrides, and environment building; ships no provider and no model id |
| `pkg/provision` | Provision method names, and in `pkg/provision/nix` the Nix package sets, defaults, pins and expansion |
| `pkg/types` | Shared type definitions |
| `pkg/validation` | Repository URL, branch, and commit validation |
| `pkg/workspace` | VCS type constants (git, jj) |
