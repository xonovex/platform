# @xonovex/agent-cli

Unified CLI for running AI coding agents with multiple model providers and sandbox options.

## Installation

```bash
npm install -g @xonovex/agent-cli
```

## Usage

```bash
# Run with default agent (Claude)
agent-cli run

# Run with specific agent
agent-cli run -a claude
agent-cli run -a opencode

# Run with sandbox
agent-cli run -s bwrap
agent-cli run -s docker
agent-cli run -s nix

# Run with worktree
agent-cli run --worktree-branch feature/my-feature

# Run with terminal wrapper
agent-cli run -t tmux
```

## Commands

### run

Run an AI coding agent.

```
Options:
  -a, --agent <type>           Agent: claude, opencode (default: claude)
  -p, --provider <name>        Model provider for the agent
  -s, --sandbox <method>       Sandbox: none, bwrap, docker, compose, nix (default: none)
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
source <(agent-cli completion)

# Add to ~/.bashrc for persistence
agent-cli completion >> ~/.bashrc
```

## Configuration

Create a config file (YAML or JSON):

```yaml
sandbox:
  method: bwrap
  network: true
  bindPaths:
    - /home/user/projects
agent: claude
provider: anthropic
```

Load with:

```bash
agent-cli run -c config.yaml
```

## License

MIT
