---
type: plan
has_subplans: false
parent_plan: plans/release-next-versions.md
parallel_group: group-3
status: complete
dependencies:
  plans:
    - plans/release-next-versions/subplan-01-catalog-downgrade-5.1.0.md
    - plans/release-next-versions/subplan-02-npm-line-bumps.md
  files:
    - .moon/toolchains.yml
skills_to_consult:
  - moon-guide
  - git-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 03: Post-Publish Toolchain Pin Update

## Objective

After the user has published and tag `moon_nix_toolchain-v0.7.0` exists, update the workspace pin in `.moon/toolchains.yml` from v0.6.1 to v0.7.0 in one commit on local `main`.

## Blocked Until

The user confirms publishing is done. This subplan must not start on an assumption; `packages/moon/AGENTS.md` forbids moving a pin before the release tag and its assets exist.

## Tasks

1. Verify the release exists and carries assets:

   ```bash
   gh release view moon_nix_toolchain-v0.7.0 --json assets,tagName
   ```

   The wasm asset and its `.sha256` must be listed. If the release is missing, STOP and report; do not edit the pin.

2. Edit `.moon/toolchains.yml:18`:

   ```yaml
   plugin: "github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.7.0"
   ```

3. Exercise the new plugin: run `npx moon run :ci-check` so moon fetches v0.7.0 and the full gate passes under it. A resolution or gate failure here is a 0.7.0 regression; report it instead of pinning back silently.
4. Commit: `chore(moon): pin moon_nix_toolchain to v0.7.0`.

## Validation Steps

- `npx moon run :ci-check` exits 0 under the v0.7.0 plugin.
- `npx moon run :ci-publish-dry-run` exits 0.

## Success Criteria

- [x] `gh release view moon_nix_toolchain-v0.7.0` shows the tag with wasm and checksum assets before any edit.
- [x] `.moon/toolchains.yml` pins `moon_nix_toolchain-v0.7.0`; no other line changes.
- [x] Both gates exit 0 with the new pin active.
- [x] Exactly one new commit on local `main`; nothing pushed.

## Files Modified

- `.moon/toolchains.yml`

## Dependencies

Subplans 01 and 02 complete, plus the external event: user has pushed and published, and the v0.7.0 release exists.

## Estimated Duration

Minimal: one line, one gate run, dominated by waiting on the publish.
