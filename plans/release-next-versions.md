---
type: plan
has_subplans: true
status: approved
updated: 2026-08-05
dependencies:
  plans: []
  subplans:
    - plans/release-next-versions/subplan-01-catalog-downgrade-5.1.0.md
    - plans/release-next-versions/subplan-02-npm-line-bumps.md
    - plans/release-next-versions/subplan-03-post-publish-toolchain-pin.md
parallel_groups:
  group-1: [subplan-01-catalog-downgrade-5.1.0]
  group-2: [subplan-02-npm-line-bumps]
  group-3: [subplan-03-post-publish-toolchain-pin]
skills_to_consult:
  - versioning-guide
  - npm-guide
  - moon-guide
  - git-guide
research_sources:
  documentation:
    - .github/workflows/release.yml
    - .moon/tasks/tag-npm.yml
    - .moon/tasks/tag-npm-platform.yml
    - .moon/tasks/tag-moon-plugin.yml
    - packages/agent/AGENTS.md
    - packages/moon/AGENTS.md
    - packages/skill/AGENTS.md
    - packages/shared/AGENTS.md
  versions:
    catalog-released: 5.0.0
    catalog-local: 6.0.0
    config-line-released: 0.1.22
    agent-line-released: 0.1.31
    moon-nix-toolchain-local: 0.7.0
    moon-nix-extension-local: 0.1.0
---

# Release Next Versions of Everything

## Overview

Prepare local `main` so the user can release the next version of every release line themselves. The work is three local prep commits (catalog version, two npm line bumps) plus one post-publish commit (toolchain pin), with the full gate suite green after each. The user pushes and publishes; the agent never pushes or publishes.

## Goals

- Catalog (73 skill and command packages plus `.claude-plugin/marketplace.json`) set to 5.1.0 in lockstep.
- Config npm line (10 `packages/config/*` plus `shared-core`) bumped 0.1.22 to 0.2.0 with changelog entries.
- Agent npm line (`agent-cli-go` plus 5 platform packages, lockstep with `optionalDependencies` refs) bumped 0.1.31 to 0.2.0 with the `## 0.2.0` changelog section `github-publish` requires.
- Moon crates untouched: toolchain 0.7.0 and extension 0.1.0 are already bumped with changelog headers.
- After the user publishes and tag `moon_nix_toolchain-v0.7.0` exists: `.moon/toolchains.yml` pin updated from v0.6.1 to v0.7.0.
- `npx moon run :ci-check` and `npx moon run :ci-publish-dry-run` pass after every commit.

## Current State

- Working tree clean; local `main` is 83 commits ahead of `origin/main`, nothing behind.
- All gates pass locally: `:ci-check` 928 tasks, `:ci-publish-dry-run` 100 tasks, both exit 0 (verified during research, warm cache).
- npm registry matches the released versions (`npm view`: config line 0.1.22, agent line 0.1.31), so every bump below is unreleased-on-top-of-released.
- Catalog sits at a never-published 6.0.0 (`38b499fd`); the published catalog is 5.0.0.
- Release automation (`release.yml`) is not used for this release: the user publishes manually via the `:ci-publish` tasks.

## Decisions (settled via plan-decide)

1. No pushing by the agent; the user pushes `origin/main` themselves.
2. Next catalog release is 5.1.0, downgrading the local 6.0.0 lockstep. The release notes the user writes must list the removals shipping under this minor: caveman and fable skills, skill-ablate, and the `acceptance-*`, `pr-*`, story-refine, version-bump commands. This knowingly ships removals under a minor version.
3. Minor bumps for both npm lines: config line 0.2.0, agent line 0.2.0.
4. No PR and no workflow changes; the user publishes manually and locally (`:ci-publish` tasks; needs `NPM_TOKEN`, `gh` auth, ghcr login). Publishing is idempotent, so full runs are retry-safe.
5. Toolchain pin update is in scope as a blocked-until-published step; `moon-nix-extension` adoption stays a separate future migration.
6. `AGENTS.md` release rules stay untouched; this manual release is a one-off exception.
7. Work happens only on `main`; `composable-workflow-*` branches are donation sources and are ignored.

## Proposed Approach

1. **Catalog to 5.1.0**: one lockstep commit setting every `packages/skill/*/package.json`, `packages/command/*/package.json`, and `.claude-plugin/marketplace.json` from 6.0.0 to 5.1.0, formatted with prettier, `package-lock.json` regenerated if it embeds workspace versions.
2. **Config line to 0.2.0**: run `npx moon run <project>:version-bump -- --type minor` per config package and `shared-core` (the script updates dependents and generates `CHANGELOG.md` from conventional commits); verify internal `@xonovex/*` refs stay exact.
3. **Agent line to 0.2.0**: same via `:version-bump` across `agent-cli-go` and the five platform packages so the lockstep and `optionalDependencies` refs move together; verify `agent-cli-go-github` platform refs and the generated `## 0.2.0` section.
4. **Gates after each commit**: `npx moon run :ci-check` and `npx moon run :ci-publish-dry-run`.
5. **Handoff**: user pushes, confirms CI green on the pushed head, publishes locally, verifies artifacts (npm versions, GitHub release tags, ghcr image).
6. **Post-publish**: once `moon_nix_toolchain-v0.7.0` exists, commit the `.moon/toolchains.yml` pin update on local `main`; user pushes again.

## Risk Assessment

- The version-bump script has not been exercised on a multi-package lockstep line in this session; a dry run (`--dry-run`) per line before the real run derisks it. If it mishandles the lockstep, fall back to manual edits per `packages/agent/AGENTS.md`.
- Local gate results used a warm moon cache; CI on the pushed head is the authoritative check and is the user's gate before publishing.
- Shipping removals under catalog 5.1.0 is a knowing semver violation (decision 2); mitigated only by release notes.
- The pin-update step is blocked on the user publishing; the plan is not complete until it lands.

## Child Plans

- group-1: [subplan-01-catalog-downgrade-5.1.0]
- group-2: [subplan-02-npm-line-bumps] (sequential after group-1: consumer `package.json` refs and `package-lock.json` overlap)
- group-3: [subplan-03-post-publish-toolchain-pin] (blocked until the user publishes and `moon_nix_toolchain-v0.7.0` exists)

## Success Criteria

- All 73 catalog packages and `marketplace.json` read 5.1.0; no other file drifts.
- All 11 config-line packages read 0.2.0 with changelog entries; all 6 agent-line packages read 0.2.0 with matching `optionalDependencies` refs and a `## 0.2.0` changelog section in `agent-cli-go`.
- `npx moon run :ci-check` and `npx moon run :ci-publish-dry-run` exit 0 after each commit.
- `.moon/toolchains.yml` pins `moon_nix_toolchain-v0.7.0` (post-publish).
- Nothing pushed or published by the agent.

## Estimated Effort

Small: three mechanical commits plus one one-line follow-up, dominated by gate runtime and the wait on the user's publish.
