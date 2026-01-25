# Xonovex Platform

Monorepo for Xonovex tools and configuration packages.

## Structure

### Subdirectories

- **`packages/config/`**: Shared configuration packages (ESLint, TypeScript, Vitest)
- **`packages/tools/`**: CLI tools and shared libraries
- **`docs/`**: Technical documentation and workflow diagrams

### Workflow

- **Setup**: `npm install`
- **Tasks**: `npx moon run <project>:<task>` or `npx moon run :<task>` for all
- **Moon**: Templates in `.moon/tasks/*.yml` auto-inherit by type/language/tags
- **Query**: Filter projects with `moon query projects`

### Code Style

- **Paradigm**: Follow `general-fp-guidelines` (functional programming)
- **Imports**: Direct from source, no re-exports or backwards-compatibility wrappers
- **Design**: Modular functions, explicit context, small focused files
- **Quality**: Strict types, clear naming, explicit error handling
- **Validation**: typecheck, lint, build, test must pass; fix warnings at root cause
- **Deprecation**: Remove unused/deprecated code immediately; do not add @deprecated markers or keep backwards-compatibility shims

## Commit Convention

Uses [Conventional Commits](https://www.conventionalcommits.org/). Format: `type(scope): description`

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`, `revert`
