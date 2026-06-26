---
type: plan
has_subplans: false
parent_plan: plans/nix-toolchain-hardening.md
parallel_group: 4
status: pending
dependencies:
  plans: [plugin-typed-config, plugin-fail-closed, plugin-flake-shell-routing, plugin-cache-coherence]
  files:
    - packages/moon/moon-nix-toolchain/Cargo.toml
    - packages/moon/moon-nix-toolchain/Cargo.lock
    - packages/moon/moon-nix-toolchain/CHANGELOG.md
    - .moon/toolchains.yml
    - ../../drodan/drodan-platform/.moon/toolchains.yml
skills_to_consult: [moon-guide, git-guide, pull-request-guide, code-review-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 07 — Rollout and Release (moon_nix_toolchain-v0.6.0)

## Objective

Cut `moon_nix_toolchain-v0.6.0` — bump `Cargo.toml`, write the `## 0.6.0` CHANGELOG entry, build/validate the `wasm32-wasip1` artifact, and publish the GitHub release/tag via the existing `moon-plugin` tasks. Then bump the `github://…@moon_nix_toolchain-v0.6.0` pin in both consumers and enable the new fail-closed config in drodan (`failClosedByTag: [cmake]`, `failClosedByLanguage: [c]`). xonovex `main` is PR-protected (a PR per change); drodan is trunk-based gitlab (push to `main`).

## Tasks

1. **Bump the crate version `0.5.0` → `0.6.0` (regenerating `Cargo.lock`).**
   File: `packages/moon/moon-nix-toolchain/Cargo.toml:3`; lockfile `packages/moon/moon-nix-toolchain/Cargo.lock:2725`.
   ```toml
   # Cargo.toml
   [package]
   name = "moon_nix_toolchain"
   version = "0.6.0"   # was 0.5.0
   ```
   ```bash
   # regenerate the lockfile entry (Cargo.lock:2724-2725 name/version) without a full rebuild
   npx moon run moon-nix-toolchain:build   # build rewrites target/ + Cargo.lock to version 0.6.0
   # or, lockfile-only:  cargo update -p moon_nix_toolchain --precise 0.6.0 \
   #   --manifest-path packages/moon/moon-nix-toolchain/Cargo.toml
   ```
   Note: this is a backward-compatible minor bump — all new config keys default off, so existing consumers are unaffected. `Cargo.lock` line 2725 must end up `version = "0.6.0"`; commit both files together.

2. **Write the `## 0.6.0` CHANGELOG section.**
   File: `packages/moon/moon-nix-toolchain/CHANGELOG.md` (insert above the current `## 0.5.0` at line 3, matching the `### Minor Changes` prose-bullet style).
   ```markdown
   ## 0.6.0

   ### Minor Changes

   - Read the toolchain config as a typed `NixToolchainConfig` (`define_toolchain_config` hook) instead of an untyped `serde_json::Value`, so unknown keys and wrong-typed values surface as config errors at registration rather than silently no-opping.
   - Fail closed for projects that must build under nix. New `failClosedByTag` / `failClosedByLanguage` lists: when a matching project's task would be wrapped but `nix` is absent from `PATH`, the task errors instead of running unwrapped. Projects not matched keep the prior host-friendly no-op.
   - Route the shell selectors through project-flake wrapping: a project shipping its own `flake.nix` now honors `shellByTask`/`shellByToolchain`/`shellByTag`/`shellByLanguage`/`shell` against its own flake (`nix develop <projectRoot>#<shell>`), instead of always falling back to that flake's `default` devShell.
   - Cache coherence via the moon_pdk_api 2.0.4 tier-2 hooks (`hash_task_contents`, `setup_environment`): the resolved flake root, devShell, and wrap mode are folded into each task's cache hash, so a flake/devShell change invalidates affected tasks instead of serving a stale wrapped result.
   ```
   Note: the `github-check` task greps `^## ${version}( |$)`, so the header must be exactly `## 0.6.0`. Keep the four bullets aligned to the four sibling subplans (typed-config, fail-closed, flake-shell-routing, cache-coherence).

3. **Build + validate the wasm and dry-run the release locally.**
   Tasks inherited from `.moon/tasks/tag-moon-plugin.yml` (build → `wasm-opt`/`wasm-strip`; `github-check` validates name/version/CHANGELOG/wasm; dry-run prints the would-be release).
   ```bash
   npx moon run moon-nix-toolchain:build               # -> target/wasm32-wasip1/release/*.wasm
   npx moon run moon-nix-toolchain:github-check         # asserts '## 0.6.0' + non-empty valid wasm
   npx moon run moon-nix-toolchain:github-publish-dry-run
   #   prints: "[dry-run] would create release moon_nix_toolchain-v0.6.0 …"
   ```
   Note: `github-check` derives `tag=${name}-v${version}` from `Cargo.toml`, i.e. `moon_nix_toolchain-v0.6.0` (underscores). Confirm the dry-run shows that exact tag and a non-empty `.wasm` before opening the PR.

4. **Land the version PR on xonovex `main`, which cuts the release/tag.**
   Repo: `xonovex/platform` (remote `origin`, PR-protected `main`). The Release workflow (`.github/workflows/release.yml`) runs `:ci-publish` → `github-publish` only when the merged PR title contains `version packages`.
   ```bash
   git switch -c chore/moon-nix-toolchain-0.6.0
   git add packages/moon/moon-nix-toolchain/{Cargo.toml,Cargo.lock,CHANGELOG.md}
   git commit -m "chore(moon-nix-toolchain): version packages 0.6.0"
   git push -u origin chore/moon-nix-toolchain-0.6.0
   gh pr create --title "chore(moon-nix-toolchain): version packages (0.6.0)" \
     --body "Release moon_nix_toolchain-v0.6.0: typed config, fail-closed tags/languages, flake-project shell routing, cache coherence."
   # after review + merge, release.yml (pull_request closed, merged, title ~ 'version packages')
   # runs `npx moon ci … :ci-publish` -> github-publish -> gh release create moon_nix_toolchain-v0.6.0
   ```
   Manual fallback if the workflow does not fire (run on merged `main`):
   ```bash
   git switch main && git pull
   npx moon run moon-nix-toolchain:build moon-nix-toolchain:github-check
   npx moon run moon-nix-toolchain:github-publish   # gh release create moon_nix_toolchain-v0.6.0 --target $(git rev-parse HEAD)
   gh release view moon_nix_toolchain-v0.6.0         # verify the .wasm + .sha256 assets exist
   ```
   Note: the consumer pin in step 5/6 cannot resolve until this tag exists — gate steps 5–6 on `gh release view moon_nix_toolchain-v0.6.0` succeeding.

5. **Bump the xonovex consumer pin `v0.5.0` → `v0.6.0` (second PR, after the tag exists).**
   File: `.moon/toolchains.yml:18`.
   ```yaml
   nix:
     plugin: 'github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.6.0'
     shellByTag:
       go: go
       shell: shell
       rust: rust
       moon-plugin: rust
   ```
   ```bash
   git switch -c chore/bump-nix-toolchain-pin-0.6.0
   git commit -am "chore(toolchains): bump moon_nix_toolchain pin to v0.6.0"
   git push -u origin chore/bump-nix-toolchain-pin-0.6.0
   gh pr create --title "chore(toolchains): bump moon_nix_toolchain to v0.6.0" \
     --body "Resolves @moon_nix_toolchain-v0.6.0. No fail-closed config: xonovex nix usage is a convenience layer (rust/go/shell), not a hard determinism gate."
   ```
   Note: xonovex deliberately keeps pin-only — no `failClosed*` — because its tasks degrade gracefully without nix. Only the pin (line 18) changes; `shellByTag` is untouched.

6. **Bump drodan's pin and enable fail-closed for the C/cmake game projects (trunk push).**
   File: `drodan/drodan-platform/.moon/toolchains.yml:19`. drodan's `packages/game/*` are `language: c`, tagged `cmake` (e.g. `game-bin2c`, `game-worldgen`, `game-advanceddrone`), and run determinism gates that genuinely require the nix-pinned clang/cmake.
   ```yaml
   nix:
     plugin: 'github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.6.0'
     shellByTag:
       cmake: cc
       go: go
       shell: shell
       kubernetes: k8s
     failClosedByTag:
       - cmake
     failClosedByLanguage:
       - c
   ```
   ```bash
   # drodan is trunk-based gitlab; commit on main and push to the gitlab trunk (+ github mirror)
   git -C ../../drodan/drodan-platform commit -am \
     "chore(toolchains): bump moon_nix_toolchain to v0.6.0; fail closed for cmake/c"
   git -C ../../drodan/drodan-platform push drodan-gitlab HEAD:main
   git -C ../../drodan/drodan-platform push drodan-github HEAD:main   # keep the mirror in sync
   ```
   Note: `failClosedByTag: [cmake]` and `failClosedByLanguage: [c]` are redundant-by-design coverage so a future C project that drops the `cmake` tag is still gated. `shellByTag.cmake: cc` is preserved.

7. **Verify both repos: tasks still wrap, and fail-closed errors when `nix` leaves `PATH`.**
   ```bash
   # xonovex — a wrapped task still runs (nix present)
   npx moon run moon-nix-toolchain:lint
   MOON_DEBUG_WASM=1 npx moon run moon-nix-toolchain:fmt-check 2>&1 | grep -i "nix develop"

   # drodan — fail-closed: drop only the nix CLI dir from PATH, keep node/moon from the devShell
   cd ../../drodan/drodan-platform
   npx moon run game-bin2c:cmake-build          # baseline: wraps + builds with nix present
   nixdir="$(dirname "$(command -v nix)")"
   PATH="${PATH//$nixdir:/}" npx moon run game-bin2c:cmake-build
   #   EXPECT: a fail-closed error naming the cmake/c project (not a silent unwrapped run)
   ```
   Note: a project not matched by `failClosedByTag`/`failClosedByLanguage` must still no-op when nix is absent (host-friendly) — spot-check one non-game drodan task under the stripped `PATH` to confirm the guard is scoped.

## Validation Steps

Run in `packages/moon/moon-nix-toolchain` consumers via moon (all must pass):

```bash
# Plugin crate (xonovex)
npx moon run moon-nix-toolchain:fmt-check     # type_check/format: cargo fmt --check
npx moon run moon-nix-toolchain:lint          # lint: cargo clippy --all-targets -- -D warnings
npx moon run moon-nix-toolchain:build         # build: cargo build --release --target wasm32-wasip1 (+ wasm-opt/strip)
npx moon run moon-nix-toolchain:test          # tests: cargo test (deps: build)
npx moon run moon-nix-toolchain:github-check  # integration: '## 0.6.0' + valid non-empty wasm, tag moon_nix_toolchain-v0.6.0
npx moon run moon-nix-toolchain:github-publish-dry-run  # integration: would-create release preview

# Release artifact exists (after merge / publish)
gh release view moon_nix_toolchain-v0.6.0     # integration: .wasm + .sha256 assets present

# Consumers resolve the new pin and wrap
npx moon run moon-nix-toolchain:lint                                  # xonovex resolves v0.6.0
cd ../../drodan/drodan-platform && npx moon run game-bin2c:cmake-build # drodan resolves v0.6.0 + wraps
```

## Success Criteria

- [ ] `Cargo.toml:3` is `version = "0.6.0"` and `Cargo.lock:2725` matches.
- [ ] `CHANGELOG.md` has a `## 0.6.0` section with the four feature bullets; `github-check` passes.
- [ ] `moon-nix-toolchain:build` + `github-check` produce a valid non-empty `wasm32-wasip1` artifact.
- [ ] GitHub release/tag `moon_nix_toolchain-v0.6.0` exists with `.wasm` + `.sha256` assets (via `release.yml` `:ci-publish` or the manual `github-publish` fallback).
- [ ] xonovex `.moon/toolchains.yml:18` pins `@moon_nix_toolchain-v0.6.0` (merged via PR).
- [ ] drodan `.moon/toolchains.yml:19` pins `@moon_nix_toolchain-v0.6.0` and adds `failClosedByTag: [cmake]` + `failClosedByLanguage: [c]` (pushed to gitlab trunk + github mirror).
- [ ] A wrapped task in both repos still runs under nix; a drodan cmake/c task errors fail-closed with `nix` off `PATH`, while an unmatched task still no-ops.

## Files Modified/Created

- `packages/moon/moon-nix-toolchain/Cargo.toml` — version `0.5.0` → `0.6.0`.
- `packages/moon/moon-nix-toolchain/Cargo.lock` — regenerated `moon_nix_toolchain` version `0.6.0`.
- `packages/moon/moon-nix-toolchain/CHANGELOG.md` — new `## 0.6.0` section.
- `.moon/toolchains.yml` (xonovex) — pin bump to `@moon_nix_toolchain-v0.6.0`.
- `../../drodan/drodan-platform/.moon/toolchains.yml` — pin bump + `failClosedByTag`/`failClosedByLanguage`.
- GitHub release/tag `moon_nix_toolchain-v0.6.0` (created, not a file).

## Dependencies

All four sibling subplans must land in the plugin source **before** task 1, because v0.6.0 ships their compiled behavior in one `.wasm` and the CHANGELOG bullets assert each feature:
- **plugin-typed-config** — `define_toolchain_config` + the typed `NixToolchainConfig` struct that the new `failClosed*` keys deserialize into.
- **plugin-fail-closed** — defines `failClosedByTag`/`failClosedByLanguage`, which task 6 enables in drodan and task 7 verifies.
- **plugin-flake-shell-routing** — flake-project shell routing claimed in the 0.6.0 CHANGELOG.
- **plugin-cache-coherence** — `hash_task_contents`/`setup_environment` tier-2 hooks claimed in the 0.6.0 CHANGELOG.
Within this subplan: tasks 5 and 6 (consumer pin bumps) are gated on task 4 publishing the tag, since `github://…@moon_nix_toolchain-v0.6.0` cannot resolve until the release exists.

## Estimated Duration

~0.5 day of hands-on work (bump, CHANGELOG, build, two PRs, one trunk push, verification) plus xonovex PR review/merge latency for the two protected-`main` PRs.
