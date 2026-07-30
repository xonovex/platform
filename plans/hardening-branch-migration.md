---
type: plan
has_subplans: true
status: approved
updated: 2026-07-30
dependencies:
  plans: []
  subplans:
    - plans/hardening-branch-migration/subplan-01-salvage-side-branch-fixes.md
    - plans/hardening-branch-migration/subplan-02-infra-config-slice.md
    - plans/hardening-branch-migration/subplan-03-moon-plugins-slice.md
    - plans/hardening-branch-migration/subplan-04-script-validation-slice.md
    - plans/hardening-branch-migration/subplan-05-shared-agent-slice.md
    - plans/hardening-branch-migration/subplan-06-skill-language-guides-slice.md
    - plans/hardening-branch-migration/subplan-07-skill-domain-guides-slice.md
    - plans/hardening-branch-migration/subplan-08-skill-process-harness-slice.md
    - plans/hardening-branch-migration/subplan-09-command-plugins-slice.md
    - plans/hardening-branch-migration/subplan-10-docs-assets-cleanup-slice.md
parallel_groups:
  group-1: [subplan-01-salvage-side-branch-fixes, subplan-02-infra-config-slice]
  group-2: [subplan-03-moon-plugins-slice, subplan-04-script-validation-slice]
  group-3: [subplan-05-shared-agent-slice]
  group-4:
    [
      subplan-06-skill-language-guides-slice,
      subplan-07-skill-domain-guides-slice,
    ]
  group-5: [subplan-08-skill-process-harness-slice]
  group-6: [subplan-09-command-plugins-slice]
  group-7: [subplan-10-docs-assets-cleanup-slice]
proposed_subplans:
  - salvage-side-branch-fixes
  - infra-config-slice
  - moon-plugins-slice
  - script-validation-slice
  - shared-agent-slice
  - skill-language-guides-slice
  - skill-domain-guides-slice
  - skill-process-harness-slice
  - command-plugins-slice
  - docs-assets-cleanup-slice
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - versioning-guide
  - moon-guide
  - skill-guide
  - command-guide
  - typescript-guide
  - npm-guide
research_sources:
  documentation: []
  versions: {}
---

# Hardening Branch Migration

## Overview

Migrate the `composable-workflow-platform-hardening` branch (212 commits, ~1,900
files, +67k/-29k) to `main` without a rebase or big-bang merge, both of which are
ruled out by measurement: `main` carries 25 agent-migrated commits since merge-base
`166c4f26` that are not patch-identical to any branch commit, and a
`git merge-tree` dry run produces 393 conflicted files (325 in `packages/skill`).
The branch is frozen as a donor of final file state and extracted into small,
independently reviewable and releasable PR slices in dependency order.

## Goals

- Land every change from the hardening branch on `main` as a series of per-area
  PRs, each passing the full moon validation gate.
- Preserve the intent of all 25 commits `main` gained since the merge base,
  including its two deliberate removals (commands in `708dfa1a`, skills in
  `b01d38ab`).
- Salvage the 4 unique commits on `composable-workflow-implementations-merge`
  (npm-audit dependency fix, runtime-probe docs).
- Finish with `git diff main composable-workflow-platform-hardening` empty for all
  migrated paths, and all three side branches plus their worktrees removed.

## Current State

- `main` (canonical, protected, release via version-packages PR): 25 commits ahead
  of merge base `166c4f26` (2026-07-12).
- `composable-workflow-platform-hardening`: 212 commits ahead; change volume by
  area: `packages/skill` 1,103 files, `packages/script` 243, `packages/agent` 196,
  `packages/moon` 73, `packages/command` 70, `packages/shared` 34,
  `packages/config` 30, `packages/diagram` 25, `packages/asset` 21, plus `.moon/`,
  root configs, `plans/`, `.devcontainer/`, `.github/workflows/`.
- File overlap between the two sides: 499 files touched by both; 393 conflict in a
  dry-run merge.
- `composable-workflow-implementations-merge` (worktree
  `../xonovex-platform-merge`): 4 commits not on the hardening branch —
  `052865b1` (fix(deps): clear npm audit advisories) and `d1692d3e` +2
  (runtime-probe docs and plan bookkeeping).
- `xonovex-platform-fable` (worktree `../xonovex-platform-fable`): 12 unique
  commits of governance/contract-plane skills — a separate feature line,
  explicitly OUT OF SCOPE here; it gets its own plan after this one completes.
- Both sides have 74 skill packages, with differing membership (`main` removed
  caveman/fable; the branch added harness-adapter and domain skills).
- `packages/moon`: the branch adds two new unreleased Rust crates
  (`moon-nix-extension` 0.1.0, `moon-nix-runtime` 0.1.0) and modifies
  `moon-nix-toolchain` source while leaving its version at 0.6.1 — identical to
  the published `moon_nix_toolchain-v0.6.1` tag both sides consume from
  `.moon/toolchains.yml`. The crates are workspace-decoupled: moon loads the
  released WASM artifact, not the local build.

## Research Findings

Approach comparison (all measured in-session, 2026-07-30):

- **Rebase branch onto main** — rejected: replays conflict resolution across up to
  212 commits because none of main's 25 commits are patch-equivalent
  (`git cherry` marks all 25 as `+`).
- **Single merge** — rejected: 393 conflicted files at once, and it releases
  everything simultaneously, which contradicts the incremental-release requirement
  for commands/skills.
- **State-based slice extraction (chosen)** — per area:
  `git checkout -b migrate/<area> main && git checkout composable-workflow-platform-hardening -- <paths>`,
  reconcile, validate, PR. Works because the branch tip is a semantic superset of
  main's migrated fixes (migration direction was branch → main) and because
  releases are per-package, so branch commit history has no release value.

Intent-preservation map — each slice must verify these main-side commits still
hold after taking the branch state:

| Slice | Main commits whose intent must survive |
| --- | --- |
| infra-config | `9dfb7522` (prettier pass), `2ea459ef` (shared vitest config), `15b5a21e` (.moon/tasks, tsconfig), `87d452e0` (tsconfig) |
| moon-plugins | `9dfb7522` (prettier pass, packages/moon part) |
| script-validation | `2ea459ef`, `ef22afec`, `b7d68e8a`, `15b5a21e`, `3f04baaa` |
| shared-agent | `022af353`, `a2b8e0f4`, `9dfb7522` (agent/shared parts) |
| skill slices | `c3b920d2`, `67fbb767`, `f121e7e7`, `6060d9ac`, `b7d68e8a`, `3f04baaa`, `27a4319c`, `6b332e83`, `a90a569a`, `66b9ad55`, `2ad6505e`, `87d452e0`, `22f48559`, `9ab3bd7d`, `0d020e3e`, `85968666`, `b01d38ab` (removal) |
| command-plugins | `f121e7e7`, `a90a569a` (skill-ablate fold), `9ab3bd7d`, `708dfa1a` (removal), `4668aade` |
| docs-assets-cleanup | `4668aade`, `85968666`, `708dfa1a`, `b01d38ab` (marketplace/catalog reconciliation) |

## Proposed Approach

1. **Freeze the donor branch** — no further commits to
   `composable-workflow-platform-hardening`; all new work happens on `main`.
2. **Extract slices in dependency order** — infra/config first, then the
   validation scripts that gate everything else, then shared/agent, then skill
   content (three thematic chunks), then commands (which delegate to skills, per
   the `config -> shared -> agent` and thin-delegation contracts), then
   docs/assets and final cleanup.
3. **Per-slice procedure** (detailed in subplans): branch from `main`, checkout
   slice paths from the donor, audit `git status` for files main deliberately
   deleted, verify the mapped main-commit intents, regenerate
   `package-lock.json` and catalog versions instead of copying them, run the moon
   validation gate, open a PR.
4. **Shared-file discipline** — `.claude-plugin/marketplace.json`,
   `.agents/plugins/marketplace.json`, and catalog version bumps are touched by
   many slices; they are reconciled once, in the final slice, to avoid
   cross-slice conflicts.
5. **Progress tracking** — after each merged slice,
   `git diff --stat main composable-workflow-platform-hardening -- <remaining paths>`
   must shrink; the plan is complete when the full diff is empty.

## Risk Assessment

- **Resurrected deletions (high)** — a blind path checkout re-adds the commands
  removed in `708dfa1a` and skills removed in `b01d38ab`. Mitigation: every slice
  diffs added files against main's removal commits before committing.
- **Main-only improvements lost (medium)** — the agent migration may have adapted
  or improved fixes beyond the branch version. Mitigation: the
  intent-preservation map above; each slice's verification task walks its mapped
  commits.
- **Validator/content coupling (medium)** — branch validation scripts may gate
  skill/command content that only exists in branch form, breaking CI if the gate
  lands before its content. Mitigation: script slice runs the gate against main's
  content before PR; any gate that fails moves into the slice that carries its
  content.
- **Shared marketplace/catalog files (medium)** — touched by nearly every slice.
  Mitigation: single reconciliation in the final slice; content slices leave
  registration files untouched unless a build breaks without them.
- **Skill chunk cross-references (low)** — skills reference sibling skills
  (handoffs, overlays like `c99-opinionated` on `c99`). Mitigation: chunk
  boundaries follow the overlay/handoff graph; the ownership/handoff validator
  (`b7d68e8a`) runs per chunk.
- **Stale plugin version on the branch (medium)** — the branch modified
  `moon-nix-toolchain` source without bumping past the published 0.6.1 tag, so
  those changes are silently unreleased. Mitigation: the moon-plugins slice bumps
  the toolchain version, releases via the normal version-packages flow, and only
  then updates the `.moon/toolchains.yml` reference; `moon-nix-extension` and
  `moon-nix-runtime` get first releases the same way.
- **Branch drift during migration (low)** — the freeze makes the donor stable;
  the fable branch stays untouched until its own plan.

## Proposed Child Plans

| # | Subplan | Scope (paths) |
| --- | --- | --- |
| 1 | salvage-side-branch-fixes | cherry-pick `052865b1` deps fix; port runtime-probe docs from `composable-workflow-implementations-merge` |
| 2 | infra-config-slice | `packages/config`, `.moon/`, root configs, `.devcontainer/`, `.github/workflows/` |
| 3 | moon-plugins-slice | `packages/moon` — new `moon-nix-extension` and `moon-nix-runtime` crates, `moon-nix-toolchain` source changes with a version bump past 0.6.1; workspace `.moon/toolchains.yml` reference bumps only after the release exists |
| 4 | script-validation-slice | `packages/script`, coupled `.moon/tasks` gates |
| 5 | shared-agent-slice | `packages/shared`, `packages/agent` |
| 6 | skill-language-guides-slice | language/framework skill packages (c99*, cmake, lua*, python, shell, sql, typescript*, hono*, zod, vitest, react, ...) |
| 7 | skill-domain-guides-slice | domain/engine skill packages (audio, ecs, gpu-*, imgui, lock-free, memory-management, data-*, networking, ...) |
| 8 | skill-process-harness-slice | process + harness skill packages (plan, git, pr, review, skill, command, instruction, harness adapters, new additions) |
| 9 | command-plugins-slice | `packages/command` (delegates to skills, so lands after slices 6-8) |
| 10 | docs-assets-cleanup-slice | `packages/diagram`, `packages/asset`, `plans/`, README, marketplace/catalog reconciliation, zero-diff verification, branch + worktree removal |

Execution groups:

- **group-1**: salvage-side-branch-fixes, infra-config-slice (independent of each
  other)
- **group-2**: moon-plugins-slice, script-validation-slice (parallel; the moon
  crates need infra's `.moon/tasks` rust templates but nothing consumes them
  locally, so they don't block or get blocked by anything downstream)
- **group-3**: shared-agent-slice
- **group-4**: skill-language-guides-slice, skill-domain-guides-slice (parallel;
  neither touches registration files)
- **group-5**: skill-process-harness-slice
- **group-6**: command-plugins-slice
- **group-7**: docs-assets-cleanup-slice

## Success Criteria

- Every slice PR passes typecheck, lint, build, and test via the moon gate before
  merge.
- No file deleted by `708dfa1a` or `b01d38ab` exists on `main` afterwards unless
  a slice PR explicitly re-introduces it with a stated reason.
- All 25 mapped main-commit intents verified present in the final state
  (checked off per slice).
- `git diff main composable-workflow-platform-hardening` is empty.
- `composable-workflow-platform-hardening`, `composable-workflow-implementations-merge`
  branches and the `../xonovex-platform-merge` worktree are deleted;
  `xonovex-platform-fable` is handed off to a follow-up plan.
- Release flow untouched: every slice merges via PR; no direct pushes.

## Estimated Effort

- 10 subplans / PRs. Slices 1-2, 4-5, and 10 are small-to-medium (hours each,
  mostly verification). Slice 3 is medium (Rust review plus a release cycle).
  Slices 6-8 carry the bulk (1,103 skill files) and dominate the effort; slice 9
  is medium. Largest cost driver is per-slice intent verification and validation
  runs, not the file copies themselves.
