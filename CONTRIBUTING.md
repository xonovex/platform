# Contributing

## Structure

```
packages/
  agent/                # CLI tools
    agent-cli/          # Agent Wrapper (TypeScript)
    agent-cli-go/       # Agent Wrapper (Go)
    agent-cli-go-*/     # Platform-specific Go binaries
  asset/                # Static assets
    asset-images/       # Shared images (diagrams, etc.)
  config/               # Shared configuration packages
    eslint-config-*/    # ESLint configurations
    ts-config-*/        # TypeScript configurations
    vitest-config-*/    # Vitest configurations
    prettier-config/    # Prettier configuration
    vite-config-base/   # Vite configuration
  doc/                  # Documentation packages
    doc-guidelines/     # Source guidelines (markdown)
    doc-agent-workflow/ # Agent workflow diagram
    doc-moon-action-diagrams/ # Moon action graph diagrams
  plugin/               # Claude Code plugins
    plugin-*-skills/    # Guideline skill plugins (35 plugins)
    plugin-workflow-commands/  # Workflow commands (plan, git, insights, code)
    plugin-utility-commands/  # Utility commands (content, instructions, skills, slash commands)
  script/               # Internal build scripts
    script-moon-common/ # Shared moon script utilities
    script-moon-*/      # Moon task scripts (action-graph, npm-check, npm-publish, version-bump, version-detect)
  shared/               # Shared libraries
    shared-core/        # Core TypeScript library (@xonovex/core)
    shared-core-go/     # Core Go library
.claude/commands/       # Claude Code slash commands
.claude-plugin/         # Claude Code plugin marketplace
```

## Development

Uses [moonrepo](https://moonrepo.dev/) for task orchestration.

```bash
npm install                         # Setup
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

## Version Bump

Packages tagged with `npm` in Moon use `moon-version-bump` to bump versions, update workspace dependents, and generate changelog entries.

```bash
npx moon run <project>:version-bump              # patch bump (default)
npx moon run <project>:version-bump -- minor      # minor bump
npx moon run <project>:version-bump -- --dry-run  # preview without writing
```

This will:

1. Bump the version in the target package's `package.json`
2. Update all workspace packages that depend on it
3. Generate a `CHANGELOG.md` entry from conventional commits since the last version change

To detect which projects have changed versions:

```bash
npx moon run moon-version-detect:run
npx moon run moon-version-detect:run -- --base main
```

## Claude Code Plugins

The monorepo hosts a Claude Code plugin marketplace at `.claude-plugin/marketplace.json` containing 37 plugins:

- **35 guideline skill plugins** — each copies guidelines from `packages/doc/doc-guidelines/` into a `skills/` directory during build
- **2 command plugins** — copy slash commands from `.claude/commands/` into a `commands/` directory during build

Install plugins via:

```
/plugin marketplace add <owner>/<repo>
/plugin install <plugin-name>
```

## Code Style

- **Paradigm**: Functional programming (see `general-fp-guidelines`)
- **Imports**: Direct from source, no re-exports
- **Design**: Modular functions, explicit context, small focused files
- **Quality**: Strict types, clear naming, explicit error handling
- **Deprecation**: Remove unused code immediately
