# Contributing

## Structure

```
packages/
  agent/                # Agent CLI, operator, and delivery packages
    agent-cli-go/       # Agent sandbox CLI (Go)
    agent-cli-go-*/     # Platform-specific Go binaries
    agent-operator-go/  # Kubernetes operator (Go)
    agent-operator-go-docker/ # Operator image publishing
    agent-governance-decision-docker/ # Governance decision image
  asset/                # Static assets
    asset-images/       # Shared images (diagrams, etc.)
  config/               # Shared configuration packages
    eslint-config-*/    # ESLint configurations
    ts-config-*/        # TypeScript configurations
    vitest-config-*/    # Vitest configurations
    prettier-config/    # Prettier configuration
    vite-config-base/   # Vite configuration
  diagram/              # Diagram packages
    diagram-moon-action/ # Moon action graph diagrams
  skill/                # Coding guidelines and skills
    skill-*/            # Skill packages (instructions, references, scripts, and assets)
  command/              # Workflow and utility commands
    command-utility/    # Utility commands (content, instructions, slash commands)
    command-workflow/   # Explicit workflow operation commands
  script/               # Internal build scripts
    script-moon-common/ # Shared moon script utilities
    script-moon-*/      # Moon task scripts (action-graph, npm-check, npm-publish, version-bump, version-detect)
  moon/                 # Shared Moon toolchains and task extensions
    moon-nix-toolchain/ # Flake-pinned task execution
  shared/               # Shared libraries
    shared-core/        # Core TypeScript library (@xonovex/core)
    shared-core-go/     # Core Go library
    shared-agent-go/    # Shared agent, provider, policy, and provisioning types
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

| Type       | Description        |
| ---------- | ------------------ |
| `feat`     | New feature        |
| `fix`      | Bug fix            |
| `docs`     | Documentation      |
| `style`    | Formatting         |
| `refactor` | Code restructuring |
| `test`     | Tests              |
| `chore`    | Maintenance        |
| `build`    | Build system       |
| `ci`       | CI configuration   |
| `perf`     | Performance        |
| `revert`   | Revert commit      |

## Version Bump and Release

Packages tagged with `npm` in Moon are versioned in lockstep. Version changes
must be submitted through a reviewed `version packages` pull request. Merging
that pull request to `main` runs the release workflow; do not publish or tag
packages directly.

The versioning workflow:

1. Bumps the version in the target package's `package.json`
2. Updates all workspace packages that depend on it
3. Generates a `CHANGELOG.md` entry from the conventional commits since the last version change

Changed-version packages are detected by comparing each `package.json` `version` against a base git ref (default the previous commit).

## Agent Skills

Each package in `packages/skill/` contains a harness-neutral `SKILL.md` and any
focused references, scripts, or assets needed by that capability.

## Code Style

- **Paradigm**: Functional programming (see `packages/skill/skill-fp/fp-guide/SKILL.md`)
- **Imports**: Direct from source, no re-exports
- **Design**: Modular functions, explicit context, small focused files
- **Quality**: Strict types, clear naming, explicit error handling
- **Deprecation**: Remove unused code immediately
