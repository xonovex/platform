# Xonovex Platform

Monorepo for Xonovex tools and configuration packages.

## Structure

### Subdirectories

- **`packages/config/`**: Shared configuration packages (ESLint, TypeScript, Vitest, Prettier, Vite)
- **`packages/agent/`**: CLI tools (agent-cli, agent-cli-go)
- **`packages/shared/`**: Shared libraries (shared-core, shared-core-go)
- **`packages/plugin/`**: Claude Code plugins (skills)
- **`packages/doc/`**: Documentation packages (action graph diagrams, workflow diagrams)
- **`packages/guide/`**: Guidelines packages
- **`packages/asset/`**: Static assets (images)

### Workflow

- **Setup**: `npm install`
- **Tasks**: `npx moon run <project>:<task>` or `npx moon run :<task>` for all
- **Moon**: Templates in `.moon/tasks/*.yml` auto-inherit by type/language/tags
- **Query**: Filter projects with `moon query projects`

### Code Style

- **Imports**: Direct from source, no re-exports or backwards-compatibility wrappers
- **Design**: Modular functions, explicit context, small focused files
- **Quality**: Strict types, clear naming, explicit error handling
- **Validation**: typecheck, lint, build, test must pass; fix warnings at root cause
- **Deprecation**: Remove unused/deprecated code immediately; do not add @deprecated markers or keep backwards-compatibility shims
- **Paradigm**: Follow the [Functional Programming](packages/guide/guide-general-fp/index.md) guidelines

## Commit Convention

Uses [Conventional Commits](https://www.conventionalcommits.org/). Format: `type(scope): description`

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`, `revert`
