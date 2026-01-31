# moon-version-bump

Bumps a package version (patch, minor, or major), updates all dependent packages in the workspace, and generates a changelog entry from git commit history.

## Usage

```bash
npx moon-version-bump              # patch bump (default)
npx moon-version-bump minor        # minor bump
npx moon-version-bump --type major # major bump
npx moon-version-bump --dry-run    # preview without writing
```

## Behavior

1. Bumps the version in `package.json`
2. Updates all workspace packages that depend on this package
3. Generates a `CHANGELOG.md` entry from conventional commits since the last version change
4. Skips changelog generation when the version was already bumped (idempotency)
