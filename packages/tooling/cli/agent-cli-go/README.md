# @xonovex/agent-cli-go

Use `agent-cli` to configure a sandbox, model provider, workspace, and terminal session before it launches Claude Code, Codex, GitHub Copilot CLI, Kiro CLI, OpenCode, or Pi.

## Installation

Install the command globally from npm.

```bash
npm install -g @xonovex/agent-cli-go
```

## Usage

Select only the agent, sandbox axes, workspace, and terminal behavior that the run needs.

```bash
# Run with default agent (Claude)
agent-cli run

# Run with specific agent
agent-cli run -a claude
agent-cli run -a codex -p gpt5-codex
agent-cli run -a kiro

# Run with the three sandbox axes: isolation, provision, and network
agent-cli run --isolation bwrap --provision command --isolation-bwrap-passthrough --init-command 'command -v claude'
agent-cli run --isolation docker --provision nix --nix-source packages --nix-rev <rev> --network none
agent-cli run --isolation bwrap --provision nix --nix-source packages --nix-rev <rev> --nix-packages ripgrep
agent-cli run --isolation bwrap --provision nix --nix-source flake --nix-shell default

# Reach only the hosts you name, and inject a credential from a store
agent-cli run --isolation bwrap --network managed --allow-host api.anthropic.com --allow-host '*.example.com:8443' --credential ANTHROPIC_API_KEY=keychain:anthropic/work

# Run with worktree
agent-cli run --worktree-branch feature/my-feature

# Run with terminal wrapper
agent-cli run -t tmux
```

## Commands

### run

Use `run` to start an AI coding agent. Select the sandbox with the three independent isolation, provision, and network axes. See `packages/tooling/cli/AGENTS.md` for the model and the four-guarantee policy.

```
Options:
  -a, --agent <type>           Agent: claude, codex, copilot, kiro, opencode, pi (default: claude)
  -p, --provider <name>        Model provider for the agent
  --isolation <method>         Isolation: none, bwrap, docker, seatbelt, docker-sandboxes, openshell (default: none)
  --provision <method>         Provision: none, nix, command (default: none)
  --network <method>           Network egress: host, none, managed (default: host)
  --allow-host <host>          Host a managed run may reach: host, host:port or *.domain (repeatable; all others are denied)
  --credential <binding>       Inject NAME=scheme:locator as an environment variable (repeatable)
  --image <image>              Container image for docker and openshell
  --isolation-bwrap-passthrough
                               Expose host tools to bwrap (forfeits host-tools-unreachable)
  --init-command <cmd>         Init command for --provision command (repeatable)
  --nix-source <kind>          Nix source: packages, flake (default: packages)
  --nix-rev <rev>              Pinned nixpkgs rev for --nix-source packages
  --nix-packages <pkg>         Additional package for --nix-source packages (repeatable; selected agent is automatic)
  --nix-shell <name>           devShell for --nix-source flake (default: default)
  --require-pinned-provision   Require provisioning from a pinned source
  --require-host-tools-unreachable
                               Require host tools to be unreachable
  --require-egress-restricted  Require disabled or enforceably restricted egress
  --require-kernel-isolation   Require a kernel-isolating runtime such as runsc
  -w, --work-dir <dir>         Working directory
  --worktree-branch <branch>   Create worktree with branch
  -t, --terminal <wrapper>     Terminal wrapper: tmux
  -c, --config <file>          Load configuration from file
  --plugin <name>              Overlay an installed plugin on the run (repeatable)
  --overlay-dir <dir>          Overlay a directory of skills, agents, commands and AGENTS.md (repeatable)
  --store <dir>                Plugin store (default: $XDG_DATA_HOME/agent-cli/plugins)
  -n, --dry-run                Show configuration without executing
```

### Isolations

Each isolation declares what it can guarantee, and `--require-*` stops a run that needs more.

| Isolation | Confines | Network `none` | Network `managed` | Notes |
| --- | --- | --- | --- | --- |
| `none` | nothing | no | no | Runs on the host. |
| `bwrap` | Linux namespaces | yes | unix socket | Attack-surface reduction, not a kernel boundary. |
| `docker` | container | yes | unix socket | `--isolation-docker-runtime runsc` adds a kernel boundary. |
| `seatbelt` | macOS `sandbox-exec` profile | yes | loopback port | macOS only. Denies file writes outside the workspace and hides the home directory. |
| `docker-sandboxes` | `sbx` sandbox | no | no | Runs the agent in its own sandbox image. Fails closed for `none` and `managed`, because `sbx` has deny rules only. |
| `openshell` | OpenShell gateway sandbox | yes | policy | Needs `--image`. Uploads the work directory, downloads it after the run, and refuses credentials and binds. |

### Managed network egress

`--network managed` lets the agent reach only the hosts of `--allow-host` (or `allowHosts` in the configuration file) and denies every other host. The allowlist denies by default, and an empty one is an error. Rules are `host`, `host:port` and `*.domain[:port]`. A rule without a port allows 443.

The run starts the allowlist proxy and stops it when the run ends. How traffic reaches the proxy depends on the isolation:

- `bwrap` and `docker` give the sandbox no network. The sandbox runs this same binary as `agent-cli egress-bridge`, which listens on `127.0.0.1:3128` inside the sandbox and relays to a unix socket of the proxy. A direct socket has no interface to use.
- `seatbelt` allows outbound connections only to the loopback port of the proxy.
- `openshell` hands the gateway a policy that lists the allowed endpoints.

The proxy refuses a host name that resolves to a loopback, private or link-local address, so an allowed name cannot reach a service of this machine. A rule that names an address allows that address.

`--require-egress-restricted` passes for `managed` only on an isolation that forces the traffic through the allowlist. An isolation without a transport (`none`, `docker-sandboxes`) fails closed for `managed` whatever the policy.

### Credentials

`--credential NAME=scheme:locator` (or `credentials` in the configuration file) reads a credential when the run starts and passes it to the agent as the environment variable `NAME`. A value is never written to an argument, an overlay or a log, and a dry run shows only where it lives. A missing, unavailable or unreadable credential stops the run before the agent starts.

| Scheme | Locator | Reads from |
| --- | --- | --- |
| `env` | variable name | An environment variable of the CLI. |
| `file` | absolute path | A regular file that only its owner can read. |
| `keychain` | `service` or `service/account` | The macOS keychain, through `security`. |
| `secret-service` | `service` or `service/account` | A Linux secret service, through `secret-tool`. |
| `pass` | entry name | The first line of a `pass` entry. |

A terminal wrapper (`--terminal tmux`) installs the environment through a private launch script that deletes itself, so a credential reaches the terminal session through that script.

### Agents

Each agent declares its facts, and the planner and the isolators compare those facts: no code names an agent. The sandbox axes and the `--require-*` guarantees stay the same for every agent.

| Agent | Binary | Nix package | Protocol | Model access | Added in an isolation |
| --- | --- | --- | --- | --- | --- |
| `claude` | `claude` | `claude-code` | Messages | provider environment | `--permission-mode bypassPermissions` |
| `codex` | `codex` | `codex` | Responses | provider arguments and environment | `--sandbox danger-full-access` |
| `copilot` | `copilot` | `github-copilot-cli` | none | the login of the user with GitHub | nothing |
| `kiro` | `kiro-cli-chat` | `kiro-cli-unwrapped` | none | the login of the user with the vendor | nothing |
| `opencode` | `opencode` | `opencode` | none | provider arguments | nothing |
| `pi` | `pi` | `pi-coding-agent` | Responses | provider arguments and environment | nothing |

- An agent that uses its own login (`copilot`, `kiro`) takes no provider: `--provider` or a provider in the configuration file fails before the run. Log in on the host before you run it with `--isolation none`.
- `kiro` starts `kiro-cli-chat chat --agent-engine v3` directly, because the `kiro-cli` router starts a program at a path under the real home directory.
- Codex sandboxes its own commands, and that fails inside another sandbox, so an isolation adds `--sandbox danger-full-access`. Host execution keeps the Codex sandbox.
- A provider that sets a route the agent does not read, such as environment for OpenCode, fails before the run, because the agent would drop it silently.
- Provider presets: `claude` has `gemini`, `gemini-claude`, `glm` and `gpt5-codex`; `codex` has `gpt5-codex`; `opencode` has `gemini`. `pi` takes its Azure OpenAI variables through `--env`.
- `--dry-run` prints the protocol, the model access and the version probe of the agent. `--agent` and `--provider` complete in the shell.

### Overlay

`--plugin` and `--overlay-dir` lay skills, agents, commands and instructions over one run without a change to the project or to the home directory of the user. The CLI gathers the artifacts, lets the dialect of the agent place them in a private temporary directory, and mounts that directory in the sandbox: read-only when the agent only reads it, read-write when the agent uses it as its configuration directory. Two sources that define one artifact name fail the run.

| Agent | Placement | Hands the directory to the agent with |
| --- | --- | --- |
| `claude` | `.claude/skills`, `.claude/agents`, `.claude/commands`, `CLAUDE.md` | `--add-dir` and `CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1` |
| `codex` | `skills`, `agents/*.toml`, `agentic.config.toml`; a command becomes a skill that the model does not start by itself | `CODEX_HOME` and `-p agentic` |
| `copilot` | a plugin directory and an instruction directory | `--plugin-dir` and `COPILOT_CUSTOM_INSTRUCTIONS_DIRS` |
| `kiro` | `skills`, `agents/*.json`, `prompts`, `steering`, `settings/cli.json` | `KIRO_HOME` |
| `opencode` | `skills`, `agents`, `commands`, `opencode.json` | `OPENCODE_CONFIG_DIR` |
| `pi` | `skills`, `prompts`, `AGENTS.md` | `--skill`, `--prompt-template`, `--append-system-prompt` |

- A project instruction file of the agent wins over an overlay instruction. The run reports each skipped artifact with its reason in `--dry-run`.
- An overlay never carries hooks, MCP servers, language servers or settings of a plugin.
- `kiro` keeps its login in the directory that its overlay replaces, so `--plugin` and `--overlay-dir` fail for `kiro`.
- `--dry-run` loads and verifies the plugins and reports the placement. It writes nothing.
- The overlay directory is removed when the run ends. A run with `--terminal` keeps it for the terminal session, and the CLI prints its path.
- The overlay changes none of the four `--require-*` guarantees: the mount holds only the files of the plugins.

### plugin

Use `plugin` to manage the plugins that a run can overlay. A plugin is a directory with a manifest (`.claude-plugin/plugin.json`, `.codex-plugin/plugin.json` or `plugin.json`) and any of `skills/`, `agents/`, `commands/` and `AGENTS.md`.

```
agent-cli plugin install <source> [--approve-digest <digest>] [--allow <capability>]
agent-cli plugin list
agent-cli plugin remove <name>
```

`<source>` is a directory, a repository URL (`https://`, `ssh://`, `file://` or `user@host:path`, with an optional `#branch-or-tag`), or `name@marketplace`.

- An install fails closed. The first call prints the name, the source, the revision, the file count and the digest of the plugin, and stops. Review the files, then repeat the call with `--approve-digest <digest>`. A digest of other content fails.
- A plugin that carries hooks, MCP servers, language servers or settings needs one `--allow hooks|mcp|lsp|settings` for each. The grant goes into the lock file, and a run fails when the stored plugin carries a capability that the lock does not grant.
- A plugin lists its dependencies in the manifest. Install a dependency first. `plugin remove` fails while another plugin needs the plugin.
- The store holds `lock.json` (name, version, source, revision, digest, grants, dependencies), `marketplaces.json` and `content/<digest>/`. Every run checks the stored files against the digest in the lock. A symlink, a file over 8 MiB, a source over 4096 files or 64 MiB, and an `http://` URL are refused.

### marketplace

Use `marketplace` to register the listings that `name@marketplace` resolves through. A marketplace is a directory or a repository with `.claude-plugin/marketplace.json`, `.agents/plugins/marketplace.json`, or both. Both files must agree on the name and on the source of each plugin they share, and a plugin source is a relative path that starts with `./` and stays inside the marketplace.

```
agent-cli marketplace add <source>
agent-cli marketplace list [name]
agent-cli marketplace remove <name>
```

`remove` fails while an installed plugin came from the marketplace.

### completion

Use `completion` to generate a completion script for the current shell. It completes agents, providers, isolations, network modes and `--credential` schemes.

```bash
# Bash
source <(agent-cli completion bash)

# Zsh
agent-cli completion zsh > "${fpath[1]}/_agent-cli"

# Fish
agent-cli completion fish | source
```

## Configuration

Create a YAML, JSON, or TOML file when run defaults must be reusable. File values provide defaults. Repeated command-line bind and environment flags append to their file equivalents.

```yaml
provider: gemini
homeDir: /home/user
bindPaths:
  - /home/user/projects
roBindPaths:
  - /home/user/reference
customEnv:
  - FEATURE_FLAG=true
allowHosts:
  - api.anthropic.com
  - '*.example.com:8443'
credentials:
  - ANTHROPIC_API_KEY=keychain:anthropic/work
```

Load the file with `--config`.

```bash
agent-cli run -c config.yaml
```

## Testing

Run unit tests for local logic and integration tests for complete command behavior.

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...
```

## License

MIT
