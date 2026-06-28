---
description: Bump a package version (patch/minor/major/exact/prerelease), propagate the change to workspace dependents, and generate a changelog entry from conventional commits
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - Skill
argument-hint: >-
  [patch|minor|major] [--type <patch|minor|major>] [--exact <version>] [--preid
  <tag>] [--dry-run] [--no-changelog] [--no-dependents] [--changelog-path <file>]
  [--git-base <ref>] [--include-types <a,b>]
---

# /xonovex-utility:version-bump — Bump Version, Propagate Dependents, Changelog

## Arguments

`/version-bump [patch|minor|major] [--type <type>] [--exact <version>] [--preid <tag>] [--dry-run] [--no-changelog] [--no-dependents] [--changelog-path <file>] [--git-base <ref>] [--include-types <a,b>]`

- `type` (optional positional): Bump level — `patch` (default), `minor`, or `major`
- `--type <type>` (optional): Bump level; same values as the positional argument
- `--exact <version>` (optional): Set this exact `X.Y.Z` / `X.Y.Z-tag.N` version instead of bumping
- `--preid <tag>` (optional): Prerelease identifier — e.g. `beta` → `1.2.4-beta.0`, then `-beta.1`
- `--dry-run` (optional): Preview every change without writing files
- `--no-changelog` (optional): Skip changelog generation
- `--no-dependents` (optional): Skip updating workspace packages that depend on this one
- `--changelog-path <file>` (optional): Changelog filename (default `CHANGELOG.md`)
- `--git-base <ref>` (optional): Override the git ref for the changelog commit range
- `--include-types <a,b>` (optional): Conventional-commit types to include (default `feat,fix,refactor,perf,docs`)

## Delegation

Load the `versioning-guide` skill (plugin `xonovex-skill-versioning`) and perform its
**version-bump** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
