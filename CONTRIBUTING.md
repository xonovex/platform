# Contributing

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
    tool-agent-cli/    # Agent Wrapper (TypeScript)
    tool-agent-cli-go/ # Agent Wrapper (Go)
  docker/              # Docker configurations
    docker-agent/      # Agent Docker image, compose, and OTEL config
docs/                  # Documentation and workflow diagrams
.claude/commands/      # Claude Code slash commands
.claude/skills/        # Claude Code skills (guidelines)
```

## Development

Uses [moonrepo](https://moonrepo.dev/) for task orchestration.

```bash
npm install                         # Setup
npm run build                       # Build all packages
npm run typecheck                   # Type check all packages
npm run lint                        # Lint all packages
npm run test                        # Run all tests
npx moon run <project>:<task>       # Run task for specific project
npx moon run :<task>                # Run task for all projects
npx moon query projects             # List all projects
```

## Commit Convention

Uses [Conventional Commits](https://www.conventionalcommits.org/).

```
type(scope): description
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation |
| `style` | Formatting |
| `refactor` | Code restructuring |
| `test` | Tests |
| `chore` | Maintenance |
| `build` | Build system |
| `ci` | CI configuration |
| `perf` | Performance |
| `revert` | Revert commit |

## Code Style

- **Paradigm**: Functional programming (see `general-fp-guidelines`)
- **Imports**: Direct from source, no re-exports
- **Design**: Modular functions, explicit context, small focused files
- **Quality**: Strict types, clear naming, explicit error handling
- **Deprecation**: Remove unused code immediately
