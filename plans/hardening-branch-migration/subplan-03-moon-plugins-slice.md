---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-2
status: complete
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
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
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

- [x] Toolchain version bumped past 0.6.1 with an accurate CHANGELOG — 0.7.0,
      set by the one `feat` among three commits; no public item added or
      removed and no breaking footer, so the bump is minor
- [x] All three crates build and test green via moon — `:ci-check --force`,
      723 tasks, exit 0
- [ ] Release tags published — **blocked, and only two are due**; see below
- [x] Workspace still consumes the OLD toolchain tag — `.moon/toolchains.yml`
      unchanged, still `moon_nix_toolchain-v0.6.1`

## Release Status

`github-check` passes for both plugins: each has a valid WASM artifact and a
matching `## <version>` changelog section, and reports itself ready to publish.
The release itself has not run, for two reasons.

**Only two tags are due, not three.** This subplan expected release tags for
the toolchain, the extension and the runtime. The runtime takes no tag: the
donor gives it the `rust` tag alone, without `moon-plugin`, so it gets none of
the `github-check` / `moon-build` release tasks. It is a path dependency
compiled into both plugin artifacts, not a plugin consumers resolve over
`github://`. The criterion above is corrected to the toolchain and the
extension.

**The release flow cannot run yet.** `packages/moon/AGENTS.md` and the parent
plan both require a reviewed PR whose title contains `version packages`, merged
to `main`, to trigger `.github/workflows/release.yml`. `main` is still unpushed,
so no PR exists to merge and no tag can be produced. Once `main` reaches
`origin`, release `moon_nix_toolchain-v0.7.0` and `moon_nix_extension-v0.1.0`
through that flow. Subplan 10 must not bump the `.moon/toolchains.yml`
reference until those tag assets exist.

## Plan Gap Found

`.hooks/` belonged to no subplan, the same class of gap subplan 11 closed for
`flake.nix` and `.github/moon.yml`. This slice needed it: the extension's
fixture workspaces carry `package.json` files, and main's `validate-lockfile`
hook matched them with a `git ls-files 'packages/*/*/package.json'` pathspec —
whose `*` spans slashes — so it demanded lockfile entries for test fixtures that
are not workspace members. The donor's hook delegates the check to
`npm ci --dry-run`, which tracks real workspaces only. `.hooks/` is taken here
rather than left for a later slice, since nothing in this slice could be
committed without it.

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
