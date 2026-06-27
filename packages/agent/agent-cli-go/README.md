# @xonovex/agent-cli-go

Go implementation of agent-cli. Configures sandboxes, providers, and terminal sessions, then launches the agent.

## Installation

```bash
npm install -g @xonovex/agent-cli-go
```

## Usage

```bash
# Run with default agent (Claude)
agent-cli run

# Run with specific agent
agent-cli run -a claude
agent-cli run -a opencode

# Run with the three sandbox axes (isolation × provision × network)
agent-cli run --isolation bwrap --provision command --init-command 'echo setup'
agent-cli run --isolation docker --network proxy --egress-allow github.com
agent-cli run --isolation bwrap --provision nix --nix-source packages --nix-rev <rev> --nix-packages ripgrep
agent-cli run --isolation bwrap --provision nix --nix-source flake --nix-shell default

# Run with worktree
agent-cli run --worktree-branch feature/my-feature

# Run with terminal wrapper
agent-cli run -t tmux
```

## Commands

### run

Run an AI coding agent. The sandbox is selected by three orthogonal axes — see
`packages/agent/AGENTS.md` for the model and the four-guarantee policy.

```
Options:
  -a, --agent <type>           Agent: claude, opencode (default: claude)
  -p, --provider <name>        Model provider for the agent
  --isolation <method>         Isolation: none, bwrap, docker (default: none)
  --provision <method>         Provision: none, nix, command (default: none)
  --network <method>           Network egress: host, none, proxy (default: host)
  --egress-allow <host>        Extra allowlist host for --network proxy (repeatable)
  --host-passthrough           Expose host/base-image tools (forfeits host-tools-unreachable)
  --init-command <cmd>         Init command for --provision command (repeatable)
  --nix-source <kind>          Nix source: packages, flake (default: packages)
  --nix-rev <rev>              Pinned nixpkgs rev for --nix-source packages
  --nix-packages <pkg>         Package for --nix-source packages (repeatable)
  --nix-shell <name>           devShell for --nix-source flake (default: default)
  --require-pinned-toolchain   Mandate pinned provisioning + host-tools-unreachable
  -w, --work-dir <dir>         Working directory
  --worktree-branch <branch>   Create worktree with branch
  -t, --terminal <wrapper>     Terminal wrapper: tmux
  -c, --config <file>          Load configuration from file
  -n, --dry-run                Show configuration without executing
```

### completion

Generate shell completion script.

```bash
# Bash
source <(agent-cli completion bash)

# Zsh
agent-cli completion zsh > "${fpath[1]}/_agent-cli"

# Fish
agent-cli completion fish | source
```

## Configuration

Create a config file (YAML or JSON):

```yaml
agent: claude
provider: anthropic
bindPaths:
  - /home/user/projects
```

Load with:

```bash
agent-cli run -c config.yaml
```

## Testing

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...
```

## License

MIT
