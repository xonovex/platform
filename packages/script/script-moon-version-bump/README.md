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
npx moon-version-bump --lockstep a,b,c --type minor   # move a release line as one
```

## Options

| Flag                      | Type    | Description                                                                                   |
| ------------------------- | ------- | --------------------------------------------------------------------------------------------- |
| `--type, -t`              | string  | Bump type: patch, minor, or major (default: patch)                                            |
| `--dry-run, -d`           | boolean | Preview changes without writing files                                                         |
| `--no-changelog`          | boolean | Skip changelog generation                                                                     |
| `--no-dependents`         | boolean | Skip updating dependent packages                                                              |
| `--changelog-path <path>` | string  | Custom changelog filename (default: `CHANGELOG.md`)                                           |
| `--preid <tag>`           | string  | Prerelease identifier (e.g. `beta` → `1.2.4-beta.0`)                                          |
| `--exact <version>`       | string  | Set exact version instead of bumping                                                          |
| `--git-base <ref>`        | string  | Override git ref for changelog commit range                                                   |
| `--include-types <types>` | string  | Comma-separated conventional commit types to include (default: `feat,fix,refactor,perf,docs`) |
| `--lockstep <packages>`   | string  | Comma-separated packages to move to one shared version in a single write                      |

## Behavior

1. Bumps the version in `package.json`
2. Updates all workspace packages that depend on this package
3. Generates a `CHANGELOG.md` entry from conventional commits since the last version change
4. Skips changelog generation when the version was already bumped (idempotency)

## Lockstep mode

`--lockstep` bumps a whole release line to one version. It reads the workspace,
plans every version and every exact `@xonovex/*` reference in memory, and writes
the result as one transaction, so no intermediate state where a reference and a
version disagree is ever observable on disk. That is what makes a line with a
dependency cycle solvable: `eslint-config-base` devDepends on `prettier-config`
while `prettier-config` depends on `eslint-config-base`, and no per-package
order can bump either one first.

```bash
npx moon run workspace-config:version-bump-lockstep -- \
  --lockstep eslint-config-base,prettier-config,ts-config-base --type minor --dry-run
```

A member is named by its package name (`@xonovex/prettier-config`) or by its
project directory (`prettier-config`). The run works from any directory inside
the workspace; the moon task above runs it from the workspace root. The
per-project `<project>:version-bump` task bumps one package; a release line goes
through the workspace task instead of a loop over that task.

Semantics:

- The target version is derived from the highest committed version among the
  members, so a member a peer already patch-bumped in the worktree is lifted to
  the shared version instead of dragging the line down.
- Every member is written at the target version, and every exact reference to a
  member is rewritten, including `optionalDependencies` such as the platform
  packages of `agent-cli-go`.
- A package outside the set that holds a reference to a member has that
  reference rewritten too, and follows the single-package dependent rule: it is
  patch-bumped once for the whole set when it is public and not already bumped
  in the worktree, and left at its version when it is private or already bumped.
  Propagation runs to a fixed point: a patch-bumped dependent is itself held by
  exact references elsewhere, so those move as well and their holders follow the
  same rule, until nothing changes. A dependent that keeps its version, because
  it is private or already bumped, ends the chain.
- Each member whose version actually moves gains a `## <version>` changelog
  section built from the conventional commits since its previous release,
  followed by the `Updated dependency ... to X` bullets for every internal
  reference that differs from `HEAD`. Re-running the same lockstep is a no-op
  and does not stack a second section.
