# Xonovex Platform

Monorepo for Xonovex tools and configuration packages.

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
docs/                  # Documentation
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
