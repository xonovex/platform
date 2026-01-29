# Xonovex Platform

Monorepo for Xonovex tools and configuration packages.

## Structure

### Subdirectories

- **`packages/config/`**: Shared configuration packages (ESLint, TypeScript, Vitest, Prettier, Vite)
- **`packages/cli/`**: CLI tools (agent-cli, agent-cli-go)
- **`packages/lib/`**: Shared libraries (core, core-go)
- **`packages/plugins/`**: Claude Code plugins (skills)
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

## Skills

- **Invoke matching skill first** - Use Skill tool when context matches available skills
- **Read relevant detail only** - Load `details/*.md` file matching the specific context, not all details

## Commit Convention

Uses [Conventional Commits](https://www.conventionalcommits.org/). Format: `type(scope): description`

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`, `revert`
