---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-2
status: pending
dependencies:
  plans: [plans/hardening-branch-migration/subplan-02-infra-config-slice.md]
  files: [packages/moon]
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - moon-guide
  - versioning-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 03: Moon Plugins Slice

## Objective

Land the branch's `packages/moon` work: the new `moon-nix-extension` and
`moon-nix-runtime` crates (both 0.1.0, unreleased) and the `moon-nix-toolchain`
source changes — which the branch left at version 0.6.1, identical to the
published tag, so they are currently unreleasable. Fix the version, land, and
release through the normal flow.

## Tasks

1. **Create the slice branch**:
   `git checkout -b migrate/moon-plugins origin/main`
2. **Checkout the donor path**:
   `git checkout composable-workflow-platform-hardening -- packages/moon`
3. **Audit the toolchain delta against the published tag**:
   `git diff moon_nix_toolchain-v0.6.1 -- packages/moon/moon-nix-toolchain` —
   classify the changes (fix vs feature vs breaking) from the branch's
   conventional commits touching the crate
   (`git log 166c4f26..composable-workflow-platform-hardening --oneline -- packages/moon/moon-nix-toolchain`).
4. **Bump versions**: set `moon-nix-toolchain` to 0.6.2/0.7.0 per the
   classification (versioning-guide), update its CHANGELOG from the conventional
   commits; confirm `moon-nix-extension` and `moon-nix-runtime` CHANGELOGs and
   0.1.0 versions are release-ready.
5. **Verify the intent of `9dfb7522`** (prettier pass) still holds on the
   `packages/moon` files, and that the crates carry correct `moon.yml` tags so
   the inherited rust task templates from subplan 02 pick them up.
6. **Run validation**, commit, open the PR. After merge, run the release flow
   (version-packages PR → `release.yml`) so tags exist for all three crates.
   Do NOT touch `.moon/toolchains.yml` here — the reference bump is subplan 10.

## Validation Steps

- `npx moon run moon-nix-toolchain:build moon-nix-toolchain:test` (and the
  extension/runtime equivalents — confirm names via
  `moon query projects --tags moon-plugin`)
- cargo tests pass for all three crates; WASM artifacts build
- CHANGELOG entries match the commit classification

## Success Criteria

- [ ] Toolchain version bumped past 0.6.1 with an accurate CHANGELOG
- [ ] All three crates build and test green via moon
- [ ] Release tags published for toolchain (new version), extension 0.1.0,
      runtime 0.1.0
- [ ] Workspace still consumes the OLD toolchain tag — no config change in this
      slice

## Files Modified/Created

- `packages/moon/moon-nix-toolchain/**` (source + version + CHANGELOG)
- `packages/moon/moon-nix-extension/**` (new)
- `packages/moon/moon-nix-runtime/**` (new)
- `packages/moon/AGENTS.md`, `packages/moon/CLAUDE.md`

## Dependencies

Subplan 02 (needs the rust/moon-plugin `.moon/tasks` templates on main).
Nothing downstream depends on this slice; runs parallel with subplan 04.

## Estimated Duration

3-5 hours plus one release cycle.
