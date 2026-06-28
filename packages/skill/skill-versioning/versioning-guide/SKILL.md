---
name: versioning-guide
description: "Use when bumping a package version, cutting a release, or detecting which packages changed version. Triggers on prompts about semantic-version bumps (patch/minor/major, prerelease/preid, exact version), propagating a version change to workspace dependents, generating or updating a CHANGELOG from conventional commits, or finding packages whose version differs between two git refs — even when the user doesn't say 'semver' or 'version'."
---

# Versioning Guidelines

Bump versions, propagate them across a workspace, and record changelogs — following Semantic Versioning and Conventional Commits.

## Core Principles

- **Semantic Versioning** — bump MAJOR/MINOR/PATCH by the change's compatibility and reset the lower fields; a pre-release sorts below its release, see [references/semver.md](references/semver.md)
- **Conventional commits drive the bump** — `feat`→minor, `fix`→patch, a breaking change (`!` or `BREAKING CHANGE:`)→major, and breaking overrides the type, see [references/changelog.md](references/changelog.md)
- **Range-preserving propagation** — when the new version falls outside a dependent's range, rewrite that range in place keeping `^`/`~`/`workspace:` and patch-bump the unbumped non-private dependent, transitively; an in-range dependent stays untouched unless an always-bump policy is explicit, see [references/version-bump.md](references/version-bump.md)
- **Idempotent bumps** — skip a package that is already bumped, keying the signal on the right baseline (committed diff before commit, registry/tag after), see [references/version-bump.md](references/version-bump.md)
- **Changelog from intent, newest-first** — prepend a leveled entry; never paste raw git logs, see [references/changelog.md](references/changelog.md)

## Operations

- **version-bump** — compute the next version, propagate to workspace dependents, write a changelog entry — see [references/version-bump.md](references/version-bump.md)
- **version-detect** — list packages whose version changed between two git refs — see [references/version-detect.md](references/version-detect.md)

## Gotchas

- A header-only conventional-commit parser misses breaking changes — `feat!:` and a `BREAKING CHANGE:` footer must both be detected, and breaking overrides the type
- Overwriting a dependent's `^1.2.0` with a bare `1.3.0` strips the caret and the `workspace:` protocol — splice the version into the range, keep the operator
- Pre-release finalize lands on the same core (`1.2.3-beta.4` → `1.2.3`), not `1.2.4` — a pre-release sorts below its release
- Git-diff idempotency silently re-bumps once the previous bump is committed — change the "already done" signal after commit
- Propagating only to direct dependents leaves transitive dependents with stale ranges — recurse or re-run in topological order
- `\w`/`\d+` regexes mis-validate versions (they allow `_` and leading zeroes and reject valid multi-field pre-releases) — match the real grammar or use a semver library

## Progressive Disclosure

- Read [references/version-bump.md](references/version-bump.md) - Load when bumping a version and propagating the change to dependents
- Read [references/version-detect.md](references/version-detect.md) - Load when detecting which packages changed version between two refs
- Read [references/semver.md](references/semver.md) - Load when computing the next version, validating a version string, or handling pre-releases
- Read [references/changelog.md](references/changelog.md) - Load when generating or updating a changelog entry from commits
