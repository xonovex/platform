# moon-version-bump

Bumps a package version (patch, minor, or major), updates all dependent packages in the workspace, and generates a changelog entry from git commit history.

## Usage

```bash
npx moon-version-bump              # patch bump (default)
npx moon-version-bump minor        # minor bump
npx moon-version-bump --type major # major bump
npx moon-version-bump --dry-run    # preview without writing
npx moon-version-bump --exact 2.0.0          # set exact version
npx moon-version-bump --preid beta           # prerelease: 1.2.4-beta.0
npx moon-version-bump --no-changelog         # skip changelog generation
npx moon-version-bump --no-dependents        # skip updating dependents
npx moon-version-bump --changelog-path CHANGES.md  # custom changelog file
npx moon-version-bump --git-base abc1234     # override git base ref
npx moon-version-bump --include-types feat,fix,chore  # custom included types
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `--type, -t` | string | Bump type: patch, minor, or major (default: patch) |
| `--dry-run, -d` | boolean | Preview changes without writing files |
| `--no-changelog` | boolean | Skip changelog generation |
| `--no-dependents` | boolean | Skip updating dependent packages |
| `--changelog-path <path>` | string | Custom changelog filename (default: `CHANGELOG.md`) |
| `--preid <tag>` | string | Prerelease identifier (e.g. `beta` → `1.2.4-beta.0`) |
| `--exact <version>` | string | Set exact version instead of bumping |
| `--git-base <ref>` | string | Override git ref for changelog commit range |
| `--include-types <types>` | string | Comma-separated conventional commit types to include (default: `feat,fix,refactor,perf,docs`) |

## Behavior

1. Bumps the version in `package.json`
2. Updates all workspace packages that depend on this package
3. Generates a `CHANGELOG.md` entry from conventional commits since the last version change
4. Skips changelog generation when the version was already bumped (idempotency)
