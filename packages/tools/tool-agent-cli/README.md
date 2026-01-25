# tool-agent-cli

Unified CLI for running AI coding agents with multiple model providers and sandbox options.

## Installation

```bash
npm install
npm run build
```

## Usage

```bash
# Run with default agent (Claude)
node dist/src/agent.js run

# Run with specific agent
node dist/src/agent.js run -a claude
node dist/src/agent.js run -a opencode

# Run with sandbox
node dist/src/agent.js run -s bwrap
node dist/src/agent.js run -s docker
node dist/src/agent.js run -s nix

# Run with worktree
node dist/src/agent.js run --worktree-branch feature/my-feature

# Run with terminal wrapper
node dist/src/agent.js run -t tmux
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
source <(node dist/src/agent.js completion)

# Add to ~/.bashrc for persistence
node dist/src/agent.js completion >> ~/.bashrc
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
node dist/src/agent.js run -c config.yaml
```

## License

MIT
