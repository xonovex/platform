---
type: plan
has_subplans: false
parent_plan: plans/moon-nix-extension.md
parallel_group: 4
status: pending
dependencies:
  plans:
    - 03-cache-contract-and-consumer-fixture
  files:
    - packages/moon/AGENTS.md
    - packages/moon/moon-nix-extension/Cargo.toml
    - packages/moon/moon-nix-extension/Cargo.lock
    - packages/moon/moon-nix-extension/CHANGELOG.md
    - packages/moon/moon-nix-extension/README.md
    - packages/moon/moon-nix-extension/moon.yml
    - packages/moon/moon-nix-extension/examples/**
    - packages/moon/moon-nix-toolchain/README.md
    - .moon/tasks/tag-moon-plugin.yml
    - .github/workflows/release.yml
skills_to_consult:
  - moon-guide
  - shell-scripting-guide
  - versioning-guide
  - git-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 04 — Publish 0.1.0 and prove the immutable locator in a pilot consumer

## Objective

Qualify and publish `moon_nix_extension-v0.1.0` through Xonovex's existing
reviewed version-packages workflow, verify the release assets/checksum/locator,
then land a separate pinned-consumer PR that exercises the GitHub release
without the Nix toolchain plugin. Document when the extension or existing
toolchain plugin is the right product.

## Context (read this first — no other context is assumed)

All code paths and validation commands resolve from the Xonovex repository
root. Subplan 03 must be complete with the local consumer fixture and all
cache/failure gates green. Re-locate anchors from the parent baseline
`80b3d773dde1e1dee516938096e9022cceccda0a` if earlier subplans changed them.

The release mechanism already exists and should be reused unchanged:

- `.moon/tasks/tag-moon-plugin.yml:3-34` builds the release WASI artifact,
  optimizes it with `wasm-opt`, strips it, and runs Cargo tests.
- Lines 56-118 validate the WASM and changelog, generate SHA-256 sidecars, and
  provide `github-publish` plus dry-run tasks. The crate name and version form
  tag `moon_nix_extension-v0.1.0` and assets
  `moon_nix_extension.wasm`/`.sha256`.
- `.github/workflows/release.yml:21-75` runs affected `:ci-publish` after a PR
  into `main` is merged with `version packages` in its title. This is the only
  sanctioned publication path; never invoke `github-publish` to bypass review
  or branch protection.
- `packages/moon/AGENTS.md` requires Cargo.toml/Cargo.lock/changelog alignment
  and consumer pin changes only after the release tag exists.
- Moon's official WASM guide confirms the monorepo locator form and recommends
  exact tags. The required locator is:

  ```text
  github://xonovex/platform/moon_nix_extension@moon_nix_extension-v0.1.0
  ```

  It also notes GitHub API rate limits and the `GITHUB_TOKEN` CI mitigation.

- Public Moon docs are currently v2.4 while Xonovex remains pinned to 2.3.5,
  and the WASM interface is documented as experimental. The release guarantees
  Moon 2.3.5/PDK 2.0.4. A current-2.4 smoke is useful compatibility evidence but
  must not silently change the workspace pin or introduce v2.4-only config.

Extensions and toolchains are workspace-scoped. Migrating one ordinary project
inside the current root Xonovex workspace cannot remove the root
`moon_nix_toolchain` locator for every other project. The 0.1.0 pilot is
therefore a dedicated nested Xonovex consumer workspace derived from the
black-box fixture, pinned to the published GitHub asset. Root-workspace or
Drodan migrations require later reviewed PRs after this pilot; do not create a
dual-plugin root configuration, even temporarily.

## Tasks

1. **Finish public product and adoption documentation.** Update both plugin
   READMEs and `packages/moon/AGENTS.md` with one consistent role matrix:

   - `moon_nix_extension`: recommended for new consumers, global native-
     toolchain mapping, dynamic central composition, mapping-gated base
     components, typed project/task overrides, lazy Nix realization, explicit
     consumer cache inputs.
   - `moon_nix_toolchain`: supported compatibility/special-purpose option for
     explicit `nix` selection, automatic `hash_task_contents`, eager
     `setup_environment`, project-flake discovery, and tag/language selectors.
   - every workspace uses exactly one of the two Nix plugins; migration replaces
     the old configuration atomically in one reviewed PR; arbitrary peer command
     replacement is unsupported; retirement needs a separate plan after hook
     parity and consumer migration.

   Include complete `.moon/extensions.yml`, `.moon/toolchains.yml` with null
   versions, inherited cache inputs, central `mkMoonShell`, scoped override,
   locked project flake, `GITHUB_TOKEN`, and exact pinned locator examples.
   Keep the README honest that a `.sha256` asset exists without claiming Moon's
   GitHub locator automatically verifies that sidecar.

2. **Qualify version and artifact locally.** Confirm `Cargo.toml` and the package
   entry in `Cargo.lock` both say 0.1.0; if an earlier implementation used a
   placeholder, change both together. Prepend an honest `## 0.1.0` changelog
   section summarizing the native-toolchain mapping, central composition,
   overrides, cache limitation, and standalone deployment. Run the complete
   runtime/toolchain/extension/fixture gates, then run
   `moon-nix-extension:github-check` and `github-publish-dry-run`. Inspect the
   optimized WASM with `wasm-validate` and verify the generated sidecar through
   `sha256sum -c`.

3. **Prepare the reviewed version-packages release PR.** Verify no
   `moon_nix_extension-v0.1.0` tag/release already exists. Ensure the PR changes
   the extension package (so affected `:ci-publish` selects it), contains only
   intended release qualification/docs/version files plus previously reviewed
   implementation commits, and has `version packages` in the title. Record the
   dry-run output and validation commands in the PR description. Opening,
   pushing, and merging the PR require explicit user/human authorization; do
   not publish directly from a local branch.

4. **Verify the release after the PR is merged.** Observe
   `.github/workflows/release.yml` to completion, then verify through GitHub that
   the tag points at the merge commit and the release contains exactly the
   non-empty `.wasm` and `.wasm.sha256` assets. Download both to a validated
   temporary directory, run `sha256sum -c`, and make Moon resolve the exact
   `github://` locator with `GITHUB_TOKEN` supplied in CI. A missing/mismatched
   asset, wrong tag target, or failed checksum blocks the consumer PR.

5. **Create the separate pinned Xonovex pilot consumer PR.** Add
   `packages/moon/moon-nix-extension/examples/pinned-consumer/` (or promote a
   clearly separated copy of the subplan-03 fixture) as a nested Moon workspace
   whose only Nix integration is the published extension locator. Keep the
   local-WASM fixture for development tests. Add a
   `released-consumer-smoke` Moon task that runs the pinned consumer's Node,
   polyglot, scoped-override, and locked-project-flake paths and is included in
   the extension CI after the release exists. Assert the consumer has no `nix`
   block/locator, no toolchain-plugin artifact, null native versions, and the
   documented cache inputs. This second PR may merge only after the release
   asset resolves and its full Moon CI passes.

6. **Record compatibility with the current public Moon line without changing
   support silently.** Run the pinned consumer against workspace Moon 2.3.5 as
   the blocking support gate. In an isolated temporary environment, run the
   same smoke against the current stable Moon 2.4.x and record its exact
   version/result in the PR. If it passes, document it as tested compatibility;
   if it fails, keep 2.3.5 as the declared support version and open a scoped
   follow-up rather than weakening the 0.1.0 fixture or upgrading Xonovex inside
   this release.

7. **Close the pilot with an explicit rollout handoff.** Record the published
   tag, checksum, release workflow run, local/pinned/current-Moon smoke results,
   and measured cold/warm startup observations in the plan/PR evidence. Do not
   modify Drodan configuration or remove existing Xonovex root toolchain pins in
   this child. List each prospective consumer and require its later migration
   PR to remove explicit `nix` task selection, add the extension locator and
   cache inputs, verify null version ownership, pass full CI, and remove the old
   pin only when that whole workspace no longer needs it.

## Validation Steps

Before the release PR:

```bash
npx moon run moon-nix-runtime:ci-check
npx moon run moon-nix-toolchain:ci-check
npx moon run moon-nix-extension:ci-check
npx moon run moon-nix-extension:integration
npx moon run moon-nix-extension:github-check
npx moon run moon-nix-extension:github-publish-dry-run
sha256sum -c packages/moon/moon-nix-extension/target/wasm32-wasip1/release/moon_nix_extension.wasm.sha256
nix flake check
```

After the release exists, in the separate consumer PR:

```bash
npx moon run moon-nix-extension:released-consumer-smoke
npx moon run moon-nix-extension:ci-check
npx moon ci :ci-check
```

Use `gh release view moon_nix_extension-v0.1.0` and a temporary download to
verify the tag/assets/checksum, but never call `gh release create` manually.
Attach the blocking 2.3.5 result and isolated 2.4.x compatibility result to the
consumer PR evidence.

## Success Criteria

- [ ] Both READMEs and `packages/moon/AGENTS.md` describe the distinct plugin
      roles, one-plugin steady state, retained toolchain use cases, cache/setup
      differences, and separate retirement gate consistently.
- [ ] Cargo.toml, Cargo.lock, changelog, crate name, WASM name, tag, and pinned
      locator all agree on `moon_nix_extension` 0.1.0.
- [ ] Complete CI, integration, GitHub check, dry-run, WASM validation, and
      local checksum verification pass before publication.
- [ ] A reviewed PR with `version packages` in its title—not a direct local
      publish—causes the existing release workflow to create the tag/release.
- [ ] The release tag targets the expected merge commit and contains exactly
      the non-empty WASM plus matching SHA-256 sidecar.
- [ ] Moon resolves the immutable GitHub locator successfully with documented
      GitHub API authentication guidance.
- [ ] The separate pinned consumer has no Nix toolchain plugin/selection,
      passes its full Moon 2.3.5 smoke/CI, and continues to declare all cache
      inputs and null native versions.
- [ ] Moon 2.4.x compatibility is explicitly recorded without changing the
      workspace pin or overstating support if it fails.
- [ ] No Drodan consumer or existing Xonovex root pin is changed implicitly;
      later workspace migrations have a documented checklist.

## Files Modified/Created

- Modified: `packages/moon/moon-nix-extension/{Cargo.toml,Cargo.lock,CHANGELOG.md,README.md,moon.yml}`
- Modified: `packages/moon/moon-nix-toolchain/README.md`
- Modified: `packages/moon/AGENTS.md`
- Created in the post-release consumer PR:
  `packages/moon/moon-nix-extension/examples/pinned-consumer/**`
- Read/validated without expected changes: `.moon/tasks/tag-moon-plugin.yml`,
  `.github/workflows/release.yml`

## Dependencies

Requires completed subplan `03-cache-contract-and-consumer-fixture`. This is
execution group 4. Publication is a mandatory reviewed/human merge gate; the
pinned consumer PR is sequential after the tag and assets exist.

## Estimated Duration

1–2 focused engineering days plus PR review/workflow latency.
