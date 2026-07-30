---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-2
status: complete
dependencies:
  plans: [plans/hardening-branch-migration/subplan-02-infra-config-slice.md]
  files: [flake.nix, nix, .github/moon.yml, .github/actions, .moon/toolchains.yml]
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - moon-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 11: Nix Shells & GitHub Project Slice

## Objective

Close the two paths subplan 02 found unowned: the root `flake.nix` (with its
`nix/` helpers) and `.github/moon.yml`. Both belong to the infrastructure tier,
and two of subplan 02's deferrals are waiting on them. Numbered last because the
other nine subplans already exist, but it runs in the infra tier — anywhere
after subplan 02.

## Why This Exists

Subplan 02 migrated `.moon/` and `.github/workflows` but its path list covered
neither the root `flake.nix` nor `.github/moon.yml`, and no other subplan claims
them. Two donor changes depend on them, so subplan 02 had to hold both back:

- `.moon/toolchains.yml` gains `shellByTag` entries `ci: ci` and `docker: docker`,
  which name devShells the root `flake.nix` does not define. `agent-operator-go`
  carries the `docker` tag, so landing them without the flake breaks it.
- The donor drops the zizmor step from `.github/workflows/ci.yml` because it
  moves the linter into a `github-actions-lint` task in `.github/moon.yml`.
  Taking the workflow change alone silently loses the security scan.

## Tasks

1. **Create the slice branch** from local `main` (not `origin/main` — see
   subplan 01 for why).
2. **Checkout donor paths**: `flake.nix`, `nix/`, `.github/moon.yml`,
   `.github/actions`.
3. **Re-enable the `shellByTag` entries** in `.moon/toolchains.yml`: `ci: ci`
   and `docker: docker`. Keep the nix `plugin:` reference on the released
   `moon_nix_toolchain-v0.6.1` tag — that bump is subplan 10's.
4. **Complete the zizmor move**: drop the standalone zizmor step from
   `.github/workflows/ci.yml`, now that `.github:ci-check` depends on
   `github-actions-lint`. Confirm zizmor still runs by resolving the task.
5. **Confirm the tag pairing holds**: `agent-operator-go` is tagged `docker` and
   its `docker-lint` needs `hadolint`, which `nix/docker.nix` supplies; the new
   `.github` project is tagged `ci` and its zizmor needs `nix/ci.nix`. Both
   helper files are already on `main`, so only the devShell wiring is new.
6. **Run validation**, commit, merge.

## Validation Steps

- `nix develop -c bash -c 'command -v hadolint && command -v zizmor'` resolves
  in the tagged shells
- `npx moon run github-actions:ci-check` passes, proving zizmor runs from moon
- Full workspace `npx moon run :ci-check --force` green with nothing cached

## Success Criteria

- [x] `flake.nix` defines `ci` and `docker` devShells and both resolve —
      `nix develop .#ci` supplies zizmor, `.#docker` supplies hadolint
- [x] `shellByTag` carries `ci` and `docker`; `agent-operator-go` still lints
- [x] zizmor runs through `github-actions:ci-check`, not the workflow step —
      audits all four workflow files, "No findings to report"
- [x] Nix toolchain `plugin:` still on `moon_nix_toolchain-v0.6.1`
- [x] Workspace `ci-check` green, uncached — 700 tasks, exit 0

## Outcome

moon maps `.github` to the project id `github-actions` through the `sources`
block in `.moon/workspace.yml`, which subplan 02 already carried; a project id
cannot start with a dot. The project count went 122 to 123 and `ci-check` went
696 tasks to 700.

## Files Modified/Created

- `flake.nix`, `nix/**`, `.github/moon.yml`, `.github/actions/**`,
  `.moon/toolchains.yml` (shellByTag), `.github/workflows/ci.yml` (zizmor step)

## Dependencies

Subplan 02 — needs the moon task templates and `.moon/toolchains.yml` in place.

## Estimated Duration

1-2 hours.
