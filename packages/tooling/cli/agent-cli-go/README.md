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
  --isolation <method>         Isolation: none, bwrap, docker (default: none)
  --provision <method>         Provision: none, nix, command (default: none)
  --network <method>           Network egress: host, none (default: host)
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
  -n, --dry-run                Show configuration without executing
```

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

### completion

Use `completion` to generate a completion script for the current shell.

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
