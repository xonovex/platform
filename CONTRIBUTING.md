# Contributing

Install the workspace dependencies, change the package that owns the concern, and run its checks before creating a conventional commit. Submit version changes through the reviewed release workflow.

## Validate a change

Run package checks while editing, then run the repository gate before delivery. Use the Nix development shell when the host lacks the required toolchains.

```bash
npm install
npx moon run <project>:ci-check --force
nix develop --no-update-lock-file --command npx moon run :ci-check --force
```

Integration and acceptance suites use separate task tags. Run the applicable suites for changes that affect process execution or external integration. A skipped or cached check is not new validation evidence.

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

Use [moonrepo](https://moonrepo.dev/) to run a package task, query projects, or run an aggregate task across the workspace.

```bash
npm install                         # Setup
npx moon run <project>:<task>       # Run task for specific project
npx moon run :<task>                # Run task for all projects
npx moon query projects             # List all projects
```

## Commit Convention

Write a [Conventional Commit](https://www.conventionalcommits.org/) that states the resulting behavior or the purpose of the change.

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

Submit version changes through a reviewed `version packages` pull request. Merging that pull request to `main` runs the release workflow; do not publish or tag packages directly.

The repository has three release lines, each versioned in lockstep within itself: the grouped plugin packages, the `npm`-tagged `config` packages together with `shared-core`, and the agent CLI with its platform binaries.

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
