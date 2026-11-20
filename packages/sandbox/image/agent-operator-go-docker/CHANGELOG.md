# @xonovex/agent-operator-go-docker

## 0.2.0

### Breaking changes

These changes need an action when you upgrade the operator from the previous release.

- A Claude Code run with a prompt starts `claude --permission-mode bypassPermissions -p <prompt>`. The previous release passed `--print --prompt <prompt>`.
- Admission refuses a prompt for an agent that declares no non-interactive run (`opencode`, `pi`, `copilot` and `kiro`), and a prompt that starts with `-`.
- Admission refuses a provider that sets CLI arguments for `claude`, a provider that sets environment for `opencode`, and any provider for `copilot` or `kiro`, which use their own login.
- `--workspace-init-image` is required. The operator ships no default init image and refuses to start without the flag.
- `presetRef` names a provider file `<name>.json` of the directory that `--providers-dir` sets. Without the flag no `presetRef` resolves, and no provider ships.

### Added

- An `AgentHarness` accepts six agent types: `claude`, `codex`, `copilot`, `kiro`, `opencode` and `pi`.

### Minor Changes

- [`f75b993`](https://github.com/xonovex/platform/commit/f75b993f6e6a6e4d27eb3a79c5337e83cccab214) [@Deorder](https://github.com/Deorder)! - define providers and models in user data with a digest approval gate for project files
- [`891a19f`](https://github.com/xonovex/platform/commit/891a19f05bbcc348716d573a86fe7d3b07f554c7) [@Deorder](https://github.com/Deorder)! - organize packages and group agent plugins
