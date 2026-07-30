# Sources

## Semantic Versioning 2.0.0

- **URL:** https://semver.org/spec/v2.0.0.html
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/semver.md`
  - `references/version-bump.md` → Choosing the Next Version
  - `SKILL.md` → Core Principles (Semantic Versioning)
- **Aspects extracted:**
  - Increment/reset rules, pre-release grammar and precedence, no-leading-zero and `[0-9A-Za-z-]` identifier constraints, build-metadata precedence exclusion

## Conventional Commits 1.0.0

- **URL:** https://www.conventionalcommits.org/en/v1.0.0/
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/changelog.md` → Deriving Entries From Commits
  - `SKILL.md` → Core Principles (Conventional commits drive the bump)
- **Aspects extracted:**
  - Commit grammar, breaking-change signals (`!` and `BREAKING CHANGE:` footer), `feat`→minor / `fix`→patch / breaking→major mapping

## Keep a Changelog 1.1.0

- **URL:** https://keepachangelog.com/en/1.1.0/
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/changelog.md` → Pick One Convention
- **Aspects extracted:**
  - Six change-type headings, `[x.y.z] - YYYY-MM-DD` / Unreleased format, "don't dump git logs"

## Changesets: config options

- **URL:** https://github.com/changesets/changesets/blob/main/docs/config-file-options.md
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/version-bump.md` → Propagating to Dependents
- **Aspects extracted:**
  - Internal-dependency range-update rule (`updateInternalDependencies` patch/minor; out-of-range default vs the experimental `updateInternalDependents: always`), `bumpVersionsWithWorkspaceProtocolOnly`, `linked` / `fixed` groups

## Changesets: changelog-github format

- **URL:** https://github.com/changesets/changesets/tree/main/packages/changelog-github
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/changelog.md` → Bullet Format
- **Aspects extracted:**
  - The `- [#PR](<repo>/pull/PR) [\`hash\`](<repo>/commit/hash) Thanks [@login](https://github.com/login)! - <description>` bullet, attributing the PR-author login rather than the git commit author name

## Changesets: changelog grouping

- **URL:** https://github.com/changesets/changesets/blob/main/docs/detailed-explanation.md
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/changelog.md` → Pick One Convention (Changesets-style)
- **Aspects extracted:**
  - Per-change bump levels rendered as `### Major Changes` / `### Minor Changes` / `### Patch Changes` groups, newest-first, sourced from intent files

## npm workspaces

- **URL:** https://docs.npmjs.com/cli/v11/using-npm/workspaces/
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/version-bump.md` → Propagating to Dependents
  - `references/version-detect.md`
- **Aspects extracted:**
  - Range-satisfaction symlinking, the `workspace:` protocol, absence of native sibling-range propagation

## Refresh Workflow

1. Re-fetch the sources above
2. Diff against the prior pull (or scan for newly added sections)
3. Update the corresponding `references/*.md`
4. Bump the **Last reviewed** dates above
