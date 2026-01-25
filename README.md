# Xonovex Platform

Monorepo containing AI agent tooling and a structured workflow for AI-assisted development. Includes an agent wrapper CLI for Claude Code and OpenCode with sandbox support (bubblewrap, Docker, Nix) and custom providers. Also includes slash commands and skills for Claude Code that enable plan-driven development with parallel execution, worktree management, and continuous validation.

## Structure

```
packages/
  config/              # Shared configuration packages
    eslint-config-*/   # ESLint configurations
    ts-config-*/       # TypeScript configurations
    vitest-config-*/   # Vitest configurations
  tools/               # CLI tools
    tool-lib/          # Shared TypeScript utilities
    tool-lib-go/       # Shared Go utilities
    tool-agent-cli/    # Agent CLI (TypeScript)
    tool-agent-cli-go/ # Agent CLI (Go)
docs/                  # Documentation and workflow diagrams
.claude/commands/      # Claude Code slash commands
.claude/skills/        # Claude Code skills (guidelines)
```

## Agent CLI

CLI tool for running AI coding agents in sandboxed environments with provider and wrapper support.

```bash
npx agent run --agent claude --sandbox bwrap --wrapper tmux
```

## Claude Commands

Slash commands for Claude Code located in `.claude/commands/`. Includes commands for planning, code quality, git workflows, and insights extraction.

## Claude Skills

Technology-specific guidelines in `.claude/skills/` covering TypeScript, React, Hono, Docker, Kubernetes, and more.

## Workflow

![Workflow Diagram](https://raw.githubusercontent.com/xonovex/platform/refs/heads/main/docs/workflow-diagram.png)

### Setup

Run the agent wrapper CLI with your preferred configuration:

| Agent | Provider | Sandbox | Example |
|-------|----------|---------|---------|
| Claude Code | Default | bubblewrap | `agent run --agent claude --sandbox bwrap` |
| Claude Code | GLM | Docker | `agent run --agent claude --provider glm --sandbox docker` |
| Claude Code | Gemini | None | `agent run --agent claude --provider gemini` |
| OpenCode | GitHub Copilot | bubblewrap | `agent run --agent opencode --provider copilot --sandbox bwrap` |

### Research & Planning

| Command | Description |
|---------|-------------|
| `plan-research` | Explain what I want, it researches viability (using Explore agents with Haiku or equivalent), suggests alternatives, tells me if the idea is good |
| `plan-create` | Creates `plans/<plan>.md` with frontmatter (status, skills to consult, library versions, parallelization info). Variants like `plan-tdd-create` generate red-green-refactor workflows |
| `plan-subplans-create` | Creates `plans/<plan>/<subplans>.md`. Even subplans of subplans are possible |
| `git-commit` | Commit pending plans to the repo |

### Worktree Setup

| Command | Description |
|---------|-------------|
| `plan-worktree-create` | Creates worktree at `../<repo>-<feature>`, sets `git config branch.<branch>.plan` so other commands know which plan is active |

Then cd into the worktree.

### Development Cycle

Repeat per session until complete:

| Command | Description |
|---------|-------------|
| `plan-continue` | Auto-detects plan from worktree config, finds where it left off |
| *(agent works)* | Agent implements the next eligible subplan |
| `plan-validate` | Validates work against guidelines, plan and test suite |
| `insights-extract` | *(optional)* Saves self-corrections to `insights/` with frontmatter |
| `plan-update` | Updates subplan and parent plan status |

### Code Quality

Optional, separate session:

| Command | Description |
|---------|-------------|
| `code-simplify` | Finds code smells |
| `code-harden` | Improves type safety, validation, error handling |

### Merge

| Command | Description |
|---------|-------------|
| `plan-worktree-merge` | Intelligent conflict resolution (knows the plan), merges to parent branch |
| `plan-validate` | *(optional)* Validates parallel group together on parent |
| `insights-integrate` | *(optional)* Merges insights into guidelines/AGENTS.md |
| `git-commit --push` | Push changes |

### Maintenance

Run as needed:

| Command | Description |
|---------|-------------|
| `code-align` | Check alignment with current guidelines |
| `shared-extract` | Extract duplicated code across packages into shared modules |

---

### Parallel Execution

Multiple agents can work on parallel subplan groups simultaneously, each needs its own worktree associated with its specific subplan.

### Agent Orchestration

An orchestrating agent can run the entire workflow autonomously by spawning agent instances that execute the commands according to a higher level goal. The human only needs to provide the initial goal, then the orchestrator handles research, planning, subplan creation, worktree management and coordinating parallel agents. Each spawned agent runs in its own session/worktree and the orchestrator monitors progress via plan status updates, decides when to merge and handles the full lifecycle. This is something I am still working on.

### Design Decisions

* **Domain-agnostic commands**: the agent figures out what to do based on context (language, platform etc.)
* **No hooks except git hooks** (for now): I give agents freedom to decide when something cannot be fixed in the current session
* **Plans committed in git**: easy to continue from another machine, branch off for alternative implementations, compare approaches
* **`*-simplify` commands** for everything (instructions, skills, slash commands) which I run occasionally to generalize, compress, remove duplication and ensure consistency

---

### Guidelines

Skills in `.claude/skills/` provide technology-specific guidelines. Standard skills cover common best practices, while `-opinionated` variants contain specialized patterns:

* `c99-guidelines` / `c99-opinionated-guidelines`: Standard C99 vs caller-owns-memory, SoA, SIMD patterns
* `lua-guidelines` / `lua-opinionated-guidelines`: Standard Lua vs LuaJIT performance optimization
* `hono-guidelines` / `hono-opinionated-guidelines`: Standard Hono vs inline OpenAPI handlers, router selection
* `general-fp-guidelines` / `general-oop-guidelines`: Functional vs object-oriented paradigms

## Setup

```bash
npm install
```

## Build

```bash
npm run build       # Build all packages
npm run typecheck   # Type check all packages
npm run lint        # Lint all packages
npm run test        # Run all tests
```

## Development

Uses [moonrepo](https://moonrepo.dev/) for task orchestration.

```bash
npx moon run <project>:<task>   # Run task for specific project
npx moon run :<task>            # Run task for all projects
npx moon query projects         # List all projects
```

## Commit Convention

Uses [Conventional Commits](https://www.conventionalcommits.org/).

```
type(scope): description

Types: feat, fix, docs, style, refactor, test, chore, build, ci, perf, revert
```

## License

MIT
