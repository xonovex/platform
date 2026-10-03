# Contributing

## Structure

Every group directory carries an `AGENTS.md` describing the rules that hold across its packages, paired with a `CLAUDE.md` that points at it.

```
packages/
  runtime/
    plugin/
      plugin-*/               # Installable groups of related skills and commands
        skills/*-guide/       # Skill instructions, references, scripts, and evals
        commands/             # Commands that load skills from the owning plugin
  tooling/
    cli/                      # Agent CLI, platform binaries, and GitHub Action
    script/                   # Moon task binaries and shared script code
    moon/                     # Nix toolchains, extensions, and their runtime
  sandbox/
    operator/                 # Kubernetes agent operator
    image/                    # Operator image build and publishing
  library/                    # Shared TypeScript and Go libraries
  config/                     # Shared configuration packages
  asset/                      # Private diagrams and images

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

The repository has three release lines, each versioned in lockstep within itself: the grouped plugin packages, the `npm`-tagged `config` packages together with `shared-core`, and the agent CLI with its platform binaries. Version changes must be submitted through a reviewed `version packages` pull request. Merging that pull request to `main` runs the release workflow; do not publish or tag packages directly.

The versioning workflow:

1. Bumps the version in the target package's `package.json`
2. Updates all workspace packages that depend on it
3. Generates a `CHANGELOG.md` entry from the conventional commits since the last version change

Changed-version packages are detected by comparing each `package.json` `version` against a base git ref (default the previous commit).

## Agent Skills

Each plugin in `packages/runtime/plugin/` owns related skills under `skills/` and commands under `commands/`. Each skill keeps its harness-neutral `SKILL.md`, references, scripts, assets, and evaluations together. Skill Moon project identifiers remain `skill-<topic>`; plugin project identifiers are `plugin-<group>`.

## Code Style

- **Paradigm**: Functional programming (see `packages/runtime/plugin/plugin-core/skills/fp-guide/SKILL.md`)
- **Imports**: Direct from source, no re-exports
- **Design**: Modular functions, explicit context, small focused files
- **Quality**: Strict types, clear naming, explicit error handling
- **Deprecation**: Remove unused code immediately
