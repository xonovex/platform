# Contributing

## Structure

Every group directory carries an `AGENTS.md` describing the rules that hold across
its packages, paired with a `CLAUDE.md` that points at it.

```
packages/
  agent/                # Agent CLI, operator, and delivery packages
    agent-cli-go/       # Agent sandbox CLI (Go)
    agent-cli-go-*/     # Platform-specific Go binaries and the GitHub Action wrapper
    agent-operator-go/  # Kubernetes operator (Go)
    agent-operator-go-docker/ # Operator image publishing
  asset/                # Static assets, private and unversioned
    asset-diagrams/     # Agent sandbox isolation diagrams (.dot sources, rendered PNGs)
    asset-images/       # Shared images
  config/               # Shared configuration packages
    eslint-config-*/    # ESLint configurations
    ts-config-*/        # TypeScript configurations
    vitest-config-*/    # Vitest configurations
    prettier-config/    # Prettier configuration
    vite-config-base/   # Vite configuration
  skill/                # Coding guidelines and skills
    skill-*/            # Skill packages (instructions, references, scripts, and assets)
  command/              # Workflow and utility commands
    command-utility/    # Utility commands (content, instructions, slash commands)
    command-workflow/   # Explicit workflow operation commands
  script/               # Moon task scripts, each shipping one binary
    script-moon-*-common/ # Shared script code, shipping no binary
    script-moon-npm-*/  # Publishing and dependency checks
    script-moon-version-*/ # Version bump and change detection
    script-moon-skill-eval-*/ # Skill trigger, routing, and output evaluation
    script-moon-skill-validate-*/ # Skill spec, link, drift, and routing validation
    script-moon-*-validate/ # Command and release validation
  moon/                 # Shared Moon toolchains and task extensions
    moon-nix-extension/ # Preferred: lazily composed Nix environments
    moon-nix-runtime/   # Shared runtime for the Nix plugins
    moon-nix-toolchain/ # Flake-pinned task execution, kept for compatibility
  shared/               # Shared libraries
    shared-core/        # Core TypeScript library
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

The repository has three release lines, each versioned in lockstep within itself:
the skill and command plugin packages, the `npm`-tagged `config` packages together
with `shared-core`, and the agent CLI with its platform binaries. Version changes
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
