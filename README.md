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

**Vision:** Self-organizing agents operating across sandboxes, containers and orchestration platforms (Kubernetes, etc.) spawning their own environments, coordinating across boundaries and managing lifecycles autonomously.

## Installation

### npm packages

```bash
# Agent CLI (TypeScript)
npm install -g @xonovex/agent-cli

# Agent CLI (Go) - auto-downloads platform binary
npm install -g @xonovex/agent-cli-go

# Claude Code skills plugin
npm install @xonovex/skills
```

### Claude Code Skills

Install the skills plugin to get coding guidelines and workflow commands:

```bash
# Add the Xonovex marketplace (from within Claude Code)
/plugin marketplace add xonovex/platform

# Install the skills plugin
/plugin install xonovex@xonovex-platform
```

Or test locally during development:

```bash
claude --plugin-dir ./packages/plugins/skills
```

Skills are namespaced as `/xonovex:<skill-name>` (e.g., `/xonovex:typescript-guidelines`).

<details>
<summary><strong>Available Skills</strong></summary>

| Category | Skills |
|----------|--------|
| Languages | `typescript-guidelines`, `python-guidelines`, `c99-guidelines`, `lua-guidelines` |
| Frameworks | `react-guidelines`, `hono-guidelines`, `express.js-guidelines`, `astro-guidelines` |
| Infrastructure | `docker-guidelines`, `kubernetes-guidelines`, `terraform-guidelines` |
| Testing | `vitest-guidelines`, `zod-guidelines` |
| Workflow | `git-guidelines`, `plan-guidelines`, `skill-guidelines` |

</details>

### Development Setup

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

[View workflow diagram](https://raw.githubusercontent.com/xonovex/platform/refs/heads/main/docs/workflow-diagram.png)

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

## Claude Integration

Slash commands in `.claude/commands/` and technology-specific guidelines in `.claude/skills/`.

<details>
<summary><strong>Commands</strong></summary>

- **Planning**: `plan-research`, `plan-create`, `plan-continue`, `plan-validate`
- **Code quality**: `code-simplify`, `code-harden`, `code-align`
- **Git workflows**: `git-commit`, `plan-worktree-create`, `plan-worktree-merge`
- **Insights**: `insights-extract`, `insights-integrate`

</details>

<details>
<summary><strong>Skills</strong></summary>

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
