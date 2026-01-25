# Xonovex Platform

![License](https://img.shields.io/badge/license-MIT-blue)
![Node](https://img.shields.io/badge/node-20%2B-green)
![Go](https://img.shields.io/badge/go-1.21%2B-00ADD8)

> Run AI coding agents in sandboxed environments with custom providers

- **Agent wrapper CLI** for Claude Code and OpenCode
- **Sandbox support**: bubblewrap, Docker, Nix
- **Custom providers**: GLM (direct), Gemini, GPT-5 Codex (via CLIProxyAPI)
- **Claude Code Tasks support** for persistent, parallel sub-agent coordination
- **Plan-driven workflow** with worktrees and parallel execution

<table><tr><td style="border: 2px solid gray; padding: 0;">
<img src="https://raw.githubusercontent.com/xonovex/platform/refs/heads/main/docs/multiple-agents.png" alt="Multiple Agents">
</td></tr></table>

## Philosophy

**Foundation:** Born from 2.5 years of working with coding agents (AutoGPT, Aider, Claude Code), this project is not just another framework. Instead it builds on existing tools by introducing composable skills and shared guidelines. Fork it, adapt it and evolve it. Through collective evolution and LLM-powered assimilation, improvements can flow in both directions. Sandboxed execution, multi-agent coordination and assimilable skills.

**Vision:** self-organizing agents operating across sandboxes, containers and orchestration platforms (Kubernetes, etc.) spawning their own environments, coordinating across boundaries and managing lifecycles autonomously.

## Index

- [Philosophy](#philosophy)
- [Quick Start](#quick-start)
- [Requirements](#requirements)
- [Installation](#installation)
- [Agent Wrapper](#agent-wrapper)
- [Workflow](#workflow)
- [Docker Sandbox](#docker-sandbox)
- [Claude Commands](#claude-commands)
- [Claude Skills](#claude-skills)
- [License](#license)

## Quick Start

```bash
# Clone and build
git clone https://github.com/xonovex/platform.git
cd platform && npm install && npm run build

# Run Claude Code in a sandbox
npx agent-cli run --agent claude --sandbox bwrap

# Or use Docker
docker compose -f packages/docker/docker-agent/compose.yaml run --rm ai-agent
```

## Requirements

| Requirement | Purpose |
|-------------|---------|
| Node.js 20+ | TypeScript CLI and development |
| Go 1.21+ | Go CLI (optional) |
| Docker | Docker sandbox |
| bubblewrap | bwrap sandbox (Linux only) |
| Nix | Nix sandbox (optional) |
| tmux | Terminal wrapper (optional) |

## Installation

```bash
# Clone and install dependencies
git clone https://github.com/xonovex/platform.git
cd platform
npm install

# Build all packages (TypeScript)
npm run build

# Build Go CLI (optional, requires Go 1.21+)
npm run build -w @xonovex/agent-cli-go
```

## Agent Wrapper

CLI tool for running AI coding agents in sandboxed environments with provider and wrapper support.

```
+-------------+     +-------------+     +-------------+
|  agent-cli  |---->|   sandbox   |---->| claude/open |
|             |     | bwrap/docker|     |    code     |
+-------------+     +-------------+     +-------------+
       |                                       |
       v                                       v
+-------------+                         +-------------+
|  provider   |                         |    your     |
| gemini/glm  |                         |    code     |
+-------------+                         +-------------+
```

### Usage

```bash
# TypeScript version
npx agent-cli run [options]

# Go version
npx agent-cli-go run [options]
```

<details>
<summary><strong>Options</strong></summary>

| Option | Description |
|--------|-------------|
| `-a, --agent <type>` | Agent to run: `claude`, `opencode` (default: `claude`) |
| `-p, --provider <name>` | Model provider: `gemini`, `gemini-claude`, `glm`, `gpt5-codex` |
| `-s, --sandbox <method>` | Sandbox: `none`, `bwrap`, `docker`, `compose`, `nix` (default: `none`) |
| `-t, --terminal <wrapper>` | Terminal wrapper: `tmux` |
| `-w, --work-dir <dir>` | Working directory |
| `-n, --dry-run` | Show command without executing |
| `-v, --verbose` | Enable verbose output |

</details>

### Examples

```bash
# Run Claude Code with bubblewrap sandbox
npx agent-cli run --agent claude --sandbox bwrap

# Run with Gemini provider
npx agent-cli run --agent claude --provider gemini

# Run in tmux session
npx agent-cli run --agent claude --sandbox bwrap --terminal tmux

# Run OpenCode with Docker sandbox
npx agent-cli run --agent opencode --sandbox docker
```

## Workflow

```
+---------------------+     +---------------------+     +---------------------+
|      Research       |     |      Planning       |     |   Worktree Setup    |
+---------------------+     +---------------------+     +---------------------+
| 1. plan-research    |---->| 1. plan-create      |---->| 1. plan-worktree-   |
|    - viability      |     | 2. plan-subplans    |     |      create         |
|    - alternatives   |     | 3. git-commit       |     | 2. cd <worktree>    |
+---------------------+     +---------------------+     +---------------------+
                                                                  |
            +-----------------------------------------------------+
            |
            v
+---------------------+     +---------------------+     +---------------------+
|  Development Loop   |     |    Code Quality     |     |        Merge        |
+---------------------+     +---------------------+     +---------------------+
| 1. plan-continue    |---->| 1. code-simplify    |---->| 1. plan-worktree-   |
| 2. (implement)      |     | 2. code-harden      |     |      merge          |
| 3. plan-validate    |     |                     |     | 2. git-commit       |
| 4. insights-extract |     +---------------------+     |      --push         |
| 5. plan-update      |            |                    +---------------------+
+---------------------+            |                              |
            ^                      |                              |
            |                      |                              v
            +--- more subplans? ---+                    +---------------------+
                                                        |        Done         |
                                                        +---------------------+

Parallel: Multiple agents work on parallel subplan groups in separate worktrees
Learning: insights-integrate merges learnings into guidelines for future sessions
```

[View full workflow diagram](https://raw.githubusercontent.com/xonovex/platform/refs/heads/main/docs/workflow-diagram.png)

### Research & Planning

| Command | Description |
|---------|-------------|
| `plan-research` | Research viability, suggest alternatives |
| `plan-create` | Create plan with frontmatter and parallelization info |
| `plan-subplans-create` | Create subplans for parallel execution |
| `git-commit` | Commit pending plans |

### Worktree Setup

| Command | Description |
|---------|-------------|
| `plan-worktree-create` | Create worktree at `../<repo>-<feature>` |

### Development Cycle

| Command | Description |
|---------|-------------|
| `plan-continue` | Auto-detect plan and resume work |
| `plan-validate` | Validate against guidelines and tests |
| `insights-extract` | Save self-corrections to `insights/` |
| `plan-update` | Update plan status |

### Code Quality

| Command | Description |
|---------|-------------|
| `code-simplify` | Find code smells |
| `code-harden` | Improve type safety and error handling |

### Merge

| Command | Description |
|---------|-------------|
| `plan-worktree-merge` | Merge with intelligent conflict resolution |
| `insights-integrate` | Merge insights into guidelines |
| `git-commit --push` | Push changes |

<details>
<summary><strong>Parallel Execution</strong></summary>

Multiple agents can work on parallel subplan groups simultaneously, each in its own worktree.

</details>

<details>
<summary><strong>Agent Orchestration</strong></summary>

An orchestrating agent can run the entire workflow autonomously by spawning agent instances that execute commands according to a higher level goal. The orchestrator handles research, planning, subplan creation, worktree management and coordinating parallel agents. This is still a work in progress.

</details>

<details>
<summary><strong>Design Decisions</strong></summary>

* **Domain-agnostic commands**: the agent figures out what to do based on context
* **No hooks except git hooks**: agents decide when something cannot be fixed
* **Plans committed in git**: continue from another machine, branch off for alternatives
* **`*-simplify` commands**: generalize, compress, remove duplication

</details>

## Docker Sandbox

Docker compose setup for running agents in isolated containers with custom provider support via [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI).

| Service | Provider | Description |
|---------|----------|-------------|
| `ai-agent` | Default | Pass-through Anthropic API |
| `ai-agent-glm` | GLM | Zhipu AI GLM-4 models |
| `ai-agent-gemini` | Gemini | Google Gemini 3.x models |
| `ai-agent-gemini-claude` | Gemini-Claude | Hybrid thinking models |
| `ai-agent-gpt5-codex` | GPT-5 Codex | OpenAI models |

### Usage

```bash
# Build the Docker image
docker build -t ai-agent -f packages/docker/docker-agent/Dockerfile .

# Run with default provider
docker compose -f packages/docker/docker-agent/compose.yaml run --rm ai-agent

# Run with Gemini provider
docker compose -f packages/docker/docker-agent/compose.yaml run --rm ai-agent-gemini
```

<details>
<summary><strong>Environment Variables</strong></summary>

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_AUTH_TOKEN` | Anthropic API token (for default provider) |
| `ZAI_AUTH_TOKEN` | Z.AI API token (for GLM provider) |
| `CLI_PROXY_API_KEY` | CLI Proxy API key (for Gemini/GPT providers) |
| `AGENT_WORK_DIR` | Working directory to mount (defaults to `$PWD`) |

</details>

## Claude Commands

Slash commands for Claude Code located in `.claude/commands/`:

- **Planning**: `plan-research`, `plan-create`, `plan-continue`, `plan-validate`
- **Code quality**: `code-simplify`, `code-harden`, `code-align`
- **Git workflows**: `git-commit`, `plan-worktree-create`, `plan-worktree-merge`
- **Insights**: `insights-extract`, `insights-integrate`

## Claude Skills

Technology-specific guidelines in `.claude/skills/` covering TypeScript, React, Hono, Docker, Kubernetes, and more.

<details>
<summary><strong>Standard vs Opinionated Skills</strong></summary>

Standard skills cover common best practices, while `-opinionated` variants contain specialized patterns:

* `c99-guidelines` / `c99-opinionated-guidelines`: Standard C99 vs caller-owns-memory, SoA, SIMD
* `lua-guidelines` / `lua-opinionated-guidelines`: Standard Lua vs LuaJIT optimization
* `hono-guidelines` / `hono-opinionated-guidelines`: Standard Hono vs inline OpenAPI handlers
* `general-fp-guidelines` / `general-oop-guidelines`: Functional vs object-oriented paradigms

</details>

## License

MIT

---

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.
