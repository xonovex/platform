# Xonovex Platform

Monorepo for Xonovex tools and configuration packages.

- `packages/config/` — shared configs (ESLint, TypeScript, Vitest, Prettier, Vite)
- `packages/agent/` — CLI tools (agent-cli, agent-cli-go)
- `packages/shared/` — shared libraries (shared-core, shared-core-go)
- `packages/skill/` — coding guidelines and skills
- `packages/command/` — workflow and utility commands
- `packages/diagram/` — diagrams (action graph, workflow)
- `packages/asset/` — static assets

- Setup → `npm install`
- Tasks → `npx moon run <project>:<task>` or `:<task>` for all
- Templates in `.moon/tasks/*.yml` auto-inherit by type/language/tags
- Query → `moon query projects`

- Direct imports from source, no re-exports or shims
- Modular functions, explicit context, small focused files
- Strict types, clear naming, explicit error handling
- typecheck, lint, build, test must pass; fix at root cause
- Remove unused/deprecated code immediately; no @deprecated markers
- Follow [Functional Programming](packages/skill/skill-general-fp/guide/SKILL.md) guidelines
- Conventional commits
