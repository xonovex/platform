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
agent-cli run -a codex -p <name>
agent-cli run -a kiro

# Run with the three sandbox axes: isolation, provision, and network
agent-cli run --isolation bwrap --provision command --isolation-bwrap-passthrough --init-command 'command -v claude'
agent-cli run --isolation docker --image <digest-pinned image> --provision nix --nix-source packages --nix-rev <rev> --network none
agent-cli run --isolation bwrap --provision nix --nix-source packages --nix-rev <rev> --nix-packages ripgrep
agent-cli run --isolation bwrap --provision nix --nix-source flake --nix-shell default

# Reach only the hosts you name, and inject a credential from a store
agent-cli run --isolation bwrap --network managed --allow-host api.anthropic.com --allow-host '*.example.com:8443' --credential ANTHROPIC_API_KEY=keychain:anthropic/work

# Run with worktree
agent-cli run --worktree-branch feature/my-feature

# Run with terminal wrapper
agent-cli run -t tmux
agent-cli run -t herdr-tui --terminal-detach
agent-cli run -t herdr-core
agent-cli run -t auto

# Check that a run could start, without starting it
agent-cli preflight --isolation bwrap --require-kernel-isolation

# Read the record of a past run
agent-cli runs list
agent-cli runs show latest --json

# Define and inspect model providers
agent-cli providers list
agent-cli run -p <name> --model sonnet=<model-id>

# Show the merged configuration and where each value came from
agent-cli config show --sources

# Write agentic.json from the configuration files of the previous release
agent-cli config migrate

# Print a JSON Schema
agent-cli schema agentic
agent-cli schema provider
```

## Commands

### run

Use `run` to start an AI coding agent. Select the sandbox with the three independent isolation, provision, and network axes. See `packages/tooling/cli/AGENTS.md` for the model and the four-guarantee policy.

```
Options:
  -a, --agent <type>           Agent: claude, codex, copilot, kiro, opencode, pi (default: claude)
  -p, --provider <name>        Model provider, the name of a provider file
      --model <slot>=<id>      Override a model slot of the provider (repeatable)
      --project <file>         Project agentic.json (default: the nearest agentic.json from the work directory up to the repository root)
      --approve-config-digest <digest>  Approve the project file with this digest
  --isolation <method>         Isolation: none, bwrap, docker, seatbelt, docker-sandboxes, openshell (default: none)
  --provision <method>         Provision: none, nix, command (default: none)
  --network <method>           Network egress: host, none, managed (default: host)
  --allow-host <host>          Host a managed run may reach: host, host:port or *.domain (repeatable; all others are denied)
  --credential <binding>       Inject NAME=scheme:locator as an environment variable (repeatable)
  --image <image>              Container image for docker and openshell (required by both; none ships)
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
  -t, --terminal <wrapper>     Terminal wrapper: tmux, herdr-tui, herdr-core, auto
  --plugin <name>              Overlay an installed plugin on the run (repeatable)
  --overlay-dir <dir>          Overlay a directory of skills, agents, commands and AGENTS.md (repeatable)
  --store <dir>                Plugin store (default: $XDG_DATA_HOME/agentic/plugins)
  -n, --dry-run                Show configuration without executing
```

### Isolations

Each isolation declares what it can guarantee, and `--require-*` stops a run that needs more.

| Isolation | Confines | Network `none` | Network `managed` | Notes |
| --- | --- | --- | --- | --- |
| `none` | nothing | no | no | Runs on the host. |
| `bwrap` | Linux namespaces | yes | unix socket | Attack-surface reduction, not a kernel boundary. |
| `docker` | container | yes | unix socket | Needs an image. `--isolation-docker-runtime runsc` adds a kernel boundary. |
| `seatbelt` | macOS `sandbox-exec` profile | yes | loopback port | macOS only. Denies file writes outside the workspace and hides the home directory. |
| `docker-sandboxes` | `sbx` sandbox | no | no | Runs the agent in its own sandbox image. Fails closed for `none` and `managed`, because `sbx` has deny rules only. |
| `openshell` | OpenShell gateway sandbox | yes | policy | Needs an image. Uploads the work directory, downloads it after the run, and refuses credentials and binds. |

`docker` and `openshell` start the agent from a container image, and agent-cli ships no image: the image is your setting. Set `image` in the `run` section of a user or managed `agentic.json` (`run.image`), or pass `--image`. A run on either isolation without an image stops before it starts. The image must hold the agent unless a `nix` or `command` provision supplies it, and only a digest-pinned image (`name@sha256:<digest>`) counts toward `--require-pinned-provision` and `--require-host-tools-unreachable`.

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

A terminal wrapper (`--terminal tmux`, `--terminal herdr-tui` or `--terminal herdr-core`) installs the environment through a private launch script that deletes itself, so a credential reaches the terminal session through that script.

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
- `agent-cli` ships no provider and no model id. You define each provider in a file, see [Providers](#providers). `pi` takes its Azure OpenAI variables through `--env`.
- `--dry-run` prints the protocol, the model access and the version probe of the agent. `--agent` and `--provider` complete in the shell, and `--provider` offers only the providers you defined.
- A harness override file `harnesses.d/<agent>.json` beside the managed, the user or the project `agentic.json` changes the facts of one agent, such as its binary. It names only the members it changes, and `agent-cli schema harness-override` prints its schema. The overrides apply user first, then project, then managed, so the managed one wins. A project override takes effect only after you approve its digest, see [Providers](#providers).

### Providers

A provider is a JSON document that names the agent it serves, the credential and the route the agent uses, and the models of that route as named slots. `agent-cli` ships no provider and no model id, so a new model never needs a new release: you change a file. Define a provider as an entry of the `providers` section of an `agentic.json` file, keyed by its name, or as the file `providers.d/<name>.json` beside an `agentic.json`. `-p <name>` selects the provider, and a name that no layer defines fails with `provider <name> is not defined; add <name>.json to <directory>`.

| Layer | Where | Role |
| --- | --- | --- |
| managed | the `providers` section of the managed `agentic.json` and `providers.d` beside it, see [Configuration](#configuration) | An administrator defines providers that no other layer replaces. |
| user | the `providers` section of `$XDG_CONFIG_HOME/agentic/agentic.json` and `$XDG_CONFIG_HOME/agentic/providers.d` | Your providers. |
| project | the `providers` section of the project `agentic.json` and `providers.d` beside it | The providers of a repository. They take effect only after you approve the digest of the file. |

A later layer replaces an earlier provider of the same name, managed last, so a project file never changes a managed provider. One layer may not define a name both in its `providers` section and in `providers.d`.

| Member | Meaning |
| --- | --- |
| `name`, `agent` | The name of the provider, which is also the file name, and the agent it serves. |
| `displayName`, `portable` | A label, and whether the route is safe outside the local host environment. |
| `extends` | A provider of the same agent to refine. Models and environment merge entry by entry, and `args` replace the inherited ones when set. |
| `credential` | `sourceEnv` names the host variable that holds the credential and `targetEnv` the variable the agent reads. The file holds names, never a value. |
| `models` | Named slots, each with an `id` and an optional `effort`. |
| `environment`, `args` | The route. A value may hold `{model:<slot>}` for the id of a slot and `{model:<slot>:effort}` for the id with its effort in parentheses, or the id alone when the slot has no effort. |

A route must use what its agent reads: `claude` reads `environment`, and `codex` reads `args` and `environment`.

Override a slot for one run with `--model <slot>=<id>`, or for every run with the `models` list of the `run` section of an `agentic.json` file. A later layer wins per slot (user, project, command line, managed), and a command line value that contradicts a managed one is an error. `--dry-run` and `preflight` show the id of each slot and the file or layer that decided it. `agent-cli providers list` shows each defined provider and its layer, and `agent-cli schema provider` prints the schema of a provider file for an editor.

Project content takes effect only after you approve the digest of its file: the project `agentic.json` when it sets the `run` or the `providers` section, each file in the project `providers.d` and each harness override in the project `harnesses.d`. `agent-cli providers approve` shows each such file and records its digest after you answer yes to `Approve this digest? [y/N]` on a terminal, or when you pass `--approve-config-digest <digest>`. A run, `preflight` and `config show` ask the same question on a terminal. Without a terminal they stop and name the digest, and `--approve-config-digest <digest>` approves it. The approval ledger is `$XDG_STATE_HOME/agentic/trust.json`, and any change to a file asks again.

The three files below are examples, not shipped code. Replace each `<model-id>` with an id your service accepts.

A Claude Code route through an Anthropic-compatible endpoint, `zai.json`:

```json
{
  "$schema": "https://schemas.xonovex.com/agent/provider.schema.json",
  "name": "zai",
  "agent": "claude",
  "portable": true,
  "credential": {"sourceEnv": "ZAI_AUTH_TOKEN", "targetEnv": "ANTHROPIC_AUTH_TOKEN"},
  "models": {
    "opus": {"id": "<model-id>"},
    "sonnet": {"id": "<model-id>"},
    "haiku": {"id": "<model-id>"}
  },
  "environment": {
    "ANTHROPIC_BASE_URL": "https://api.z.ai/api/anthropic",
    "API_TIMEOUT_MS": "3000000",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "{model:opus}",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "{model:sonnet}",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "{model:haiku}"
  }
}
```

A Claude Code route through a proxy on the local host, `proxy.json`, with a reasoning effort on one slot:

```json
{
  "name": "proxy",
  "agent": "claude",
  "credential": {"sourceEnv": "CLI_PROXY_API_KEY", "targetEnv": "ANTHROPIC_AUTH_TOKEN"},
  "models": {
    "opus": {"id": "<model-id>", "effort": "high"},
    "sonnet": {"id": "<model-id>"}
  },
  "environment": {
    "ANTHROPIC_BASE_URL": "http://127.0.0.1:8317",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "{model:opus:effort}",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "{model:sonnet}",
    "CLAUDE_CODE_SUBAGENT_MODEL": "{model:sonnet}"
  }
}
```

A Codex route through the same proxy, `codex-proxy.json`:

```json
{
  "name": "codex-proxy",
  "agent": "codex",
  "credential": {"sourceEnv": "CLI_PROXY_API_KEY", "targetEnv": "CLI_PROXY_API_KEY"},
  "models": {"default": {"id": "<model-id>"}},
  "args": [
    "-c", "model_provider=\"proxy\"",
    "-c", "model_providers.proxy.name=\"CLI proxy\"",
    "-c", "model_providers.proxy.base_url=\"http://127.0.0.1:8317/v1\"",
    "-c", "model_providers.proxy.wire_api=\"responses\"",
    "-c", "model_providers.proxy.env_key=\"CLI_PROXY_API_KEY\"",
    "-c", "model=\"{model:default}\""
  ]
}
```

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
- The overlay directory is removed when the run ends. A `herdr-core` run that waits for the agent removes it when the agent ends. A `tmux`, `herdr-tui` or detached run keeps it for the terminal session, because the agent runs on after the CLI returns, and the CLI prints its path.
- The overlay changes none of the four `--require-*` guarantees: the mount holds only the files of the plugins.

### Terminals

| Terminal | Runs the agent in | Notes |
| --- | --- | --- |
| `tmux` | a tmux session and window | `--terminal-session` and `--terminal-window` name them. The run attaches unless `--terminal-detach` is set. |
| `herdr-tui` | a pane beside the current pane of the herdr terminal user interface | Needs the CLI to run in a herdr-tui pane (`HERDR_ENV=1`). It never attaches to a herdr session the CLI does not run in. `--terminal-detach` keeps the focus where it is. |
| `herdr-core` | a new terminal of the herdr-core runtime, a child of the terminal the CLI runs in | Needs the CLI to run in a herdr-core terminal (`HC_TERMINAL_ID` set) and the `hc` command. It talks to the runtime through `hc` only. Each terminal carries the label `agent-cli.run=<run id>`, so `hc ls --label agent-cli.run=<run id>` finds it. An interrupt of the CLI ends the program and removes its terminal. With `--terminal-detach` the terminal outlives the CLI and removes itself when the program exits. |
| `auto` | the wrapper whose session the CLI runs in | Chooses `herdr-core` when `HC_TERMINAL_ID` is set and `herdr-tui` when `HERDR_ENV=1`. It fails when neither is set, and when both are set it chooses `herdr-core`, because every program in a `herdr-core` terminal inherits `HC_TERMINAL_ID`, a `herdr-tui` session included. It never chooses tmux. |

The exit code of a terminal run differs by wrapper:

- `herdr-tui` returns when herdr has accepted the command, not when the agent ends, so the CLI exits 0 and the run record shows exit code 0.
- `herdr-core` waits for the program and returns its exit code, which the run record keeps. A program that a signal ended reports 128 plus the signal, and a run you interrupt reports 130. With `--terminal-detach` it returns when the terminal has started, with exit code 0.

### preflight

Use `preflight` to resolve a run and report whether it could start. It takes the flags of `run` that describe a run and the same configuration layers, and it starts nothing, reads no credential and writes no overlay.

```
agent-cli preflight [run flags] [--json]
```

The report lists each axis choice with the layer that made it, each guarantee with the reason it holds or fails, and each check of the host: the isolation tool, the terminal, the version control, the credential stores and the plugins of the overlay. Exit code 0 means the run could start, 1 means it could not, and 2 means the command line was wrong.

### runs

Every run that is not a dry run leaves a JSON record in `$XDG_STATE_HOME/agentic/runs` (default `~/.local/state/agentic/runs`), readable by its owner only. A record holds the agent and its version probe, the choice on each axis, each guarantee with its outcome, the digest of each plugin and overlay directory, the names of the credentials, the work directory, the exit code and the layers that decided each configuration key. It never holds a credential value. A run that fails before the agent starts leaves a record with the error and no exit code.

```
agent-cli runs list [--json]
agent-cli runs show <id|latest> [--json]
```

A store that cannot be written costs the run its record and never its result: the CLI prints a warning.

### providers

Use `providers list` to show every defined provider with its agent, the layer and the file it came from, and, for a project file, whether its digest is approved. Use `providers show <name>` to show one provider: its file, the hosts its route reaches, its model slots and its approval. Use `providers approve` to review and approve the project files, see [Providers](#providers), and `providers revoke [digest]` to remove the approval of the project files, or of one digest, from the ledger. Each takes `--project` and `--work-dir`.

### config

Use `config show` to print the merged `run` and `providers` sections as JSON, and `config show --sources` to print each value with the layer and the file that set it. It reads the layers as a run does, so project content passes the trust gate first. It takes `--project`, `--work-dir` and `--approve-config-digest`.

Use `config migrate` to write the legacy configuration files as `agentic.json` and `providers.d` files, see [Legacy configuration](#legacy-configuration). It takes `--work-dir` for the project and `--managed` to also write the managed file.

### schema

Use `schema` to print a JSON Schema (draft 2020-12): `agentic` for an `agentic.json` file, `agentic-run` and `agentic-providers` for its `run` and `providers` sections, `provider` for a provider file, `harness-override` for a harness file and `run-record` for a run record. `agent-cli schema` lists the names.

### plugin

Use `plugin` to manage the plugins that a run can overlay. A plugin is a directory with a manifest (`.claude-plugin/plugin.json`, `.codex-plugin/plugin.json` or `plugin.json`) and any of `skills/`, `agents/`, `commands/` and `AGENTS.md`.

```
agent-cli plugin install <source> [--approve-config-digest <digest>] [--allow <capability>]
agent-cli plugin list
agent-cli plugin remove <name>
```

`<source>` is a directory, a repository URL (`https://`, `ssh://`, `file://` or `user@host:path`, with an optional `#branch-or-tag`), or `name@marketplace`.

- An install fails closed. It prints the name, the source, the revision and the file count of the plugin, then passes the same trust gate as project content. On a terminal it shows the digest, asks `Approve this digest? [y/N]` and installs only on `y` or `yes`: an empty answer, any other answer, the end of the input or an interrupt installs nothing and exits 1. Without a terminal it stops and names the digest. Review the files, then repeat the call with `--approve-config-digest <digest>`, which never asks. A digest of other content fails. The approval ledger `$XDG_STATE_HOME/agentic/trust.json` keeps the digest, so the same content installs again without a question, and `providers revoke <digest>` removes it. The answer never grants a capability: pass `--allow` for each.
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

Use `completion` to generate a completion script for the current shell. It completes agents, providers, isolations, network modes, terminals, `--credential` schemes, schema names and run ids.

```bash
# Bash
source <(agent-cli completion bash)

# Zsh
agent-cli completion zsh > "${fpath[1]}/_agent-cli"

# Fish
agent-cli completion fish | source
```

## Configuration

Put reusable run defaults in the `run` section of an `agentic.json` file. The CLI merges three files and the command line, and `--dry-run`, `preflight` and `config show --sources` show which layer decided each value.

| Layer | Where | Role |
| --- | --- | --- |
| managed | `/etc/agentic/agentic.json` (`/Library/Application Support/agentic` on macOS, `%ProgramData%\agentic` on Windows) | An administrator sets a value that no other layer can change. |
| user | `$XDG_CONFIG_HOME/agentic/agentic.json` (default `~/.config/agentic/agentic.json`) | The defaults of the user. |
| project | the nearest `agentic.json` from the work directory up to the repository root, or the file of `--project` | The defaults of a repository. It can set only keys that cannot widen access: `agent`, `provider`, the four `require*` keys and `models`. Its `run` and `providers` sections take effect only after you approve its digest, see [Providers](#providers). |
| flags | the command line | The choice of one run. |

Each key has one merge rule, listed by `agent-cli schema agentic-run`:

- `replace` (`agent`, `provider`, `isolation`, `provision`, `network`, `image`, `homeDir`): the managed value wins, then the flag, then the project, then the user. A flag that contradicts a managed value is an error.
- `append` (`models`, `bindPaths`, `roBindPaths`, `customEnv`, `allowHosts`, `credentials`): the items of every layer join, user first and managed last, without duplicates.
- `require` (`requirePinnedProvision`, `requireHostToolsUnreachable`, `requireEgressRestricted`, `requireKernelIsolation`): a guarantee that any layer demands stays demanded.

```json
{
  "$schema": "https://schemas.xonovex.com/agent/agentic.schema.json",
  "version": 1,
  "run": {
    "agent": "claude",
    "provider": "zai",
    "models": ["sonnet=<model-id>"],
    "isolation": "bwrap",
    "network": "managed",
    "image": "registry.example/agent@sha256:<digest>",
    "homeDir": "/home/user",
    "bindPaths": ["/home/user/projects"],
    "roBindPaths": ["/home/user/reference"],
    "customEnv": ["FEATURE_FLAG=true"],
    "allowHosts": ["api.anthropic.com", "*.example.com:8443"],
    "credentials": ["ANTHROPIC_API_KEY=keychain:anthropic/work"],
    "requireEgressRestricted": true
  }
}
```

### Legacy configuration

This release still reads the configuration files of the previous release, and the next release removes them. Each legacy file prints a warning, and `--dry-run`, `preflight` and `config show --sources` name it with the suffix `(legacy)`.

| Layer | Legacy file | Legacy provider directory |
| --- | --- | --- |
| managed | `/etc/agent-cli/managed.yaml`, `managed.yml` or `managed.toml` (`/Library/Application Support/agent-cli` on macOS, `%ProgramData%\agent-cli` on Windows) | `/etc/agent-cli/providers` |
| user | `~/.config/sandboxed-claude/config` (YAML, TOML or `key=value`) | `~/.config/sandboxed-claude/providers` |
| project | `.agent-cli.yaml`, `.agent-cli.yml` or `.agent-cli.toml` in the work directory | `.agent-cli/providers` in the work directory |

A layer that has both a legacy file and an `agentic.json` is an error that names both files. The legacy project providers and model overrides take effect only after you approve their digest, and an approval in the previous ledger `$XDG_DATA_HOME/agent-cli/approvals.json` still holds.

`agent-cli config migrate` writes the user file, the project file of the work directory and, with `--managed`, the managed file as `agentic.json`, and each legacy provider file as `providers.d/<name>.json` beside it. It checks each file against the schema and never replaces a file. Check the new files, then remove the legacy files. A migrated project file needs a new approval.

The run records moved from `$XDG_STATE_HOME/agent-cli/runs` to `$XDG_STATE_HOME/agentic/runs`, and the plugin store moved from `$XDG_DATA_HOME/agent-cli/plugins` to `$XDG_DATA_HOME/agentic/plugins`. The store holds only relative paths, so move the old directory to the new place to keep the installed plugins and marketplaces.

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
