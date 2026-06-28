# version-bump: Bump a Version, Propagate Dependents, Update Changelog

Compute the next version for a package, propagate the change to every workspace package that depends on it, and prepend a changelog entry derived from conventional commits.

## Contents

- [Goal](#goal)
- [Core Workflow](#core-workflow)
- [Choosing the Next Version](#choosing-the-next-version)
- [Propagating to Dependents](#propagating-to-dependents)
- [Idempotency](#idempotency)
- [Output](#output)
- [Error Handling](#error-handling)
- [Gotchas](#gotchas)

## Goal

- Move a package from its current version to the next one (level bump, exact set, or prerelease)
- Keep workspace dependents consistent and self-versioned
- Record the change in a changelog without hand-editing

## Core Workflow

1. **Read** the package's `package.json`; abort if `name` or `version` is missing.
2. **Choose the next version** — an exact override, a prerelease (`--preid`), or a level (`patch`/`minor`/`major`) taken from the argument or inferred from conventional commits. See [semver.md](semver.md).
3. **Skip if already bumped** — see [Idempotency](#idempotency).
4. **Write** the new `version` (unless `--dry-run`).
5. **Propagate to dependents** — for every other workspace package that references this one in `dependencies` / `devDependencies` / `peerDependencies` / `optionalDependencies`, rewrite the range _preserving its prefix_ and patch-bump that dependent once (non-private, not already bumped) **only when the new version falls outside its existing range** — an in-range dependent (e.g. `^1.2.0` still satisfying the new version) needs neither. Cascade transitively (topological order or to a fixpoint).
6. **Generate the changelog** unless `--no-changelog`. See [changelog.md](changelog.md).
7. **Report** old→new for the package and each updated dependent.

## Choosing the Next Version

| Input                       | Result                                                                                      |
| --------------------------- | ------------------------------------------------------------------------------------------- |
| `--exact X.Y.Z[-tag.N]`     | set exactly that version (validate it first)                                                |
| `--preid <id>`              | first prerelease `X.Y.Z-id.0`; on a matching existing prerelease, increment the counter     |
| `major` / `minor` / `patch` | increment that field and reset lower fields; on an existing prerelease, finalize to `X.Y.Z` |

Default level is `patch`. Validate any version string before writing — see [semver.md](semver.md).

## Propagating to Dependents

- Rewrite the dependent's declared range **in place, keeping `^` / `~` / `workspace:`** — replace only the version part, never the operator or protocol.
- A range that already satisfies the new version needs no dependent release; rewrite and patch-bump only when the new version falls outside the existing range (or when policy is "always bump dependents" — be explicit about which you apply).
- Patch-bump a dependent at most once per run; skip `private` dependents (they ship no version).
- Cascade transitively: a bumped dependent is itself a dependency for others, so re-propagate or process packages in dependency order.

## Idempotency

- Treat "already bumped" as: the committed version (`git show HEAD:<path>`) already differs from the working-tree version → skip.
- This holds only while the bump is **uncommitted**. After committing, the working tree equals `HEAD` and the signal disappears — compare the working version against the published registry version or a release tag instead, or a second run bumps again.
- Keep this baseline aligned with [version-detect](version-detect.md): comparing detection against `HEAD~1` but idempotency against `HEAD` can disagree.

## Output

```
@scope/pkg: 1.4.2 -> 1.5.0
@scope/consumer: 2.1.0 -> 2.1.1 (dependency updated)
Bumped @scope/pkg to 1.5.0, updated deps in 3 file(s).
```

Prefix every line with `[dry-run]` when previewing.

## Error Handling

- **Error** — no `package.json`, or missing `name`/`version` → abort.
- **Error** — invalid `--exact` value or an unknown level → abort, stating the expected format.
- **Warning** — no previous version reachable for the changelog → write the version, skip the changelog.

## Gotchas

- Finalizing a pre-release lands on its own core (`1.2.3-beta.4` → `1.2.3`); jumping to `1.2.4` orphans the release the prerelease was leading up to.
- `--no-dependents` and `--no-changelog` are independent — skipping dependents still writes the primary package's changelog, and skipping the changelog still propagates to dependents.
