# Xonovex Platform

Monorepo for Xonovex tools and configuration packages.

## Structure

- `packages/config/` — shared configs (ESLint, TypeScript, Vitest, Prettier, Vite)
- `packages/agent/` — CLI tools (agent-cli, agent-cli-go)
- `packages/shared/` — shared libraries (shared-core, shared-core-go)
- `packages/guide/` — Claude Code guides
- `packages/prompt/` — Claude Code prompts
- `packages/diagram/` — diagrams (action graph, workflow)
- `packages/asset/` — static assets

## Workflow

- Setup → `npm install`
- Tasks → `npx moon run <project>:<task>` or `:<task>` for all
- Templates in `.moon/tasks/*.yml` auto-inherit by type/language/tags
- Query → `moon query projects`

## Code Style

- Direct imports from source, no re-exports or shims
- Modular functions, explicit context, small focused files
- Strict types, clear naming, explicit error handling
- typecheck, lint, build, test must pass; fix at root cause
- Remove unused/deprecated code immediately; no @deprecated markers
- Follow [Functional Programming](packages/guide/guide-general-fp/index.md) guidelines

## Commit Convention

[Conventional Commits](https://www.conventionalcommits.org/): `type(scope): description`

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`, `revert`
