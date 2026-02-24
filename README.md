# Xonovex Platform

![License](https://img.shields.io/badge/license-MIT-blue)
![Node](https://img.shields.io/badge/node-20%2B-green)
![Go](https://img.shields.io/badge/go-1.21%2B-00ADD8)

> Execution context manager for AI coding agents

- **Unified agent CLI** for Claude Code and OpenCode
- **Sandbox support**: bubblewrap, Docker, Nix
- **Custom providers**: GLM (direct), Gemini, GPT-5 Codex (via CLIProxyAPI)
- **Plan-driven workflow** with worktrees and parallel execution
- **Kubernetes orchestration** for scalable, cloud-native agent execution

## Why agent-cli?

AI coding agents handle prompts, tools, and code changes — they *are* the agent. What they don't control is the environment they run in: where the process executes, which model backs it, how the terminal session is managed, and whether the workspace is reproducible.

agent-cli manages everything *around* the agent. Some agents have overlapping built-in features (Claude Code supports tmux and worktrees natively), but agent-cli operates at a different layer — it configures the execution context *before* the agent launches, rather than being controlled by the agent at runtime:

- **Sandbox isolation** — run agents inside bwrap, Docker, or Nix containers so they can't touch your host system
- **Model provider routing** — proxy agent API calls to Gemini, GLM, or other models transparently
- **Terminal session management** — create and attach tmux sessions with git-aware naming before the agent starts, rather than relying on the agent to manage its own terminal
- **Reproducible environments** — Nix-based package sets for consistent dev tooling across machines
- **Multi-agent support** — unified interface across Claude Code, OpenCode, and future agents, with the same sandbox and session management regardless of which agent runs inside
- **Kubernetes orchestration** — run agents as Kubernetes pods for scalable, cloud-native execution

agent-cli is not an agent itself. It doesn't read your code, make decisions, or generate patches. It sets up the sandbox, wires the provider, launches the session and gets out of the way.

<table><tr><td style="border: 2px solid gray; padding: 0;">
<img src="packages/asset/asset-images/multiple-agents.png" alt="Multiple Agents">
</td></tr></table>

## Philosophy

**Foundation:** Built from years of experience using coding agents like AutoGPT, Aider and Claude Code, this project provides a shared skill library and execution model that can be extended and improved collaboratively. Designed to integrate with existing systems rather than replace them.

**Vision:** Autonomous agents coordinating across sandboxed and containerized environments, with support for orchestration platforms like Kubernetes and lifecycle management.

## Installation

### npm packages

```bash
# Agent CLI (TypeScript)
npm install -g @xonovex/agent-cli

# Agent CLI (Go) - auto-downloads platform binary
npm install -g @xonovex/agent-cli-go
```

## Running Agents

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

### CLI

```bash
agent-cli run --agent claude --sandbox bwrap
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

<details>
<summary><strong>Examples</strong></summary>

```bash
# Run Claude Code with bubblewrap sandbox
agent-cli run --agent claude --sandbox bwrap

# Run with Gemini provider
agent-cli run --agent claude --provider gemini

# Run in tmux session
agent-cli run --agent claude --sandbox bwrap --terminal tmux

# Run OpenCode with Docker sandbox
agent-cli run --agent opencode --sandbox docker
```

</details>

### Docker

Docker compose with custom provider support via [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI).

```bash
docker compose -f packages/docker/docker-agent/compose.yaml run --rm ai-agent
```

<details>
<summary><strong>Services</strong></summary>

| Service | Provider | Description |
|---------|----------|-------------|
| `ai-agent` | Default | Pass-through Anthropic API |
| `ai-agent-glm` | GLM | Zhipu AI GLM-4 models |
| `ai-agent-gemini` | Gemini | Google Gemini 3.x models |
| `ai-agent-gemini-claude` | Gemini-Claude | Hybrid thinking models |
| `ai-agent-gpt5-codex` | GPT-5 Codex | OpenAI models |

</details>

<details>
<summary><strong>Environment Variables</strong></summary>

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_AUTH_TOKEN` | Anthropic API token (for default provider) |
| `ZAI_AUTH_TOKEN` | Z.AI API token (for GLM provider) |
| `CLI_PROXY_API_KEY` | CLI Proxy API key (for Gemini/GPT providers) |
| `AGENT_WORK_DIR` | Working directory to mount (defaults to `$PWD`) |

</details>

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

[View workflow diagram](packages/diagram/diagram-agent-workflow/workflow-diagram.png)

<details>
<summary><strong>Commands</strong></summary>

| Command | Description |
|---------|-------------|
| `plan-research` | Research viability, suggest alternatives |
| `plan-create` | Create plan with frontmatter and parallelization info |
| `plan-subplans-create` | Create subplans for parallel execution |
| `plan-worktree-create` | Create worktree at `../<repo>-<feature>` |
| `plan-continue` | Auto-detect plan and resume work |
| `plan-validate` | Validate against guidelines and tests |
| `plan-update` | Update plan status |
| `plan-worktree-merge` | Merge with intelligent conflict resolution |
| `code-simplify` | Find code smells |
| `code-harden` | Improve type safety and error handling |
| `insights-extract` | Save self-corrections to `insights/` |
| `insights-integrate` | Merge insights into guidelines |
| `git-commit` | Commit changes (use `--push` to push) |

</details>

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

## Development Setup

| Requirement | Purpose |
|-------------|---------|
| Node.js 20+ | TypeScript CLI and development |
| Go 1.21+ | Go CLI (optional) |
| Docker | Docker sandbox |
| bubblewrap | bwrap sandbox (Linux only) |
| Nix | Nix sandbox (optional) |
| tmux | Terminal wrapper (optional) |

```bash
git clone https://github.com/xonovex/platform.git
cd platform && npm install && npm run build
```

<details>
<summary><strong>Build Go CLI (optional)</strong></summary>

```bash
npm run build -w @xonovex/agent-cli-go
```

</details>

## License

MIT

---

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.
