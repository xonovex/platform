---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-1
status: pending
dependencies:
  plans: []
  files:
    [
      packages/config,
      .moon,
      .devcontainer,
      .github/workflows,
      tsconfig.json,
      tsconfig.options.json,
      .prettierignore,
      .npmrc,
      vocabulary.json,
      package.json,
    ]
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - moon-guide
  - typescript-guide
  - npm-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 02: Infra & Config Slice

## Objective

Land the branch's infrastructure state on main: shared configs
(`packages/config`), moon task templates and workspace config (`.moon/`), root
configs, devcontainer, and CI workflows. This is the foundation every later
slice validates against.

## Tasks

1. **Create the slice branch**:
   `git checkout -b migrate/infra-config origin/main`
2. **Checkout donor paths** from the frozen branch:
   ```bash
   git checkout composable-workflow-platform-hardening -- \
     packages/config .moon .devcontainer .github/workflows \
     tsconfig.json tsconfig.options.json .prettierignore .npmrc vocabulary.json
   ```
   Do NOT take `package.json`/`package-lock.json` wholesale — diff
   `package.json` manually, apply intended changes, run `npm install` to
   regenerate the lockfile.
3. **Defer coupled gates**: review each changed `.moon/tasks/*.yml` template. Any
   task that invokes a `packages/script` validator not yet on main, or gates
   content that only exists in branch form, is reverted here and moves to
   subplan 04 (script slice) — keep a written list in the PR description.
   Similarly keep `.moon/toolchains.yml`'s nix `plugin:` reference at the
   currently released tag; the bump belongs to subplan 10 after subplan 03's
   release.
4. **Verify main-side intents survive** (parent plan intent map):
   - `9dfb7522` — prettier formatting and `.prettierignore` entries still present
   - `2ea459ef` — `packages/config` still exports the shared vitest config the
     script package consumes
   - `15b5a21e` — the `.moon/tasks` command-validation wiring and `tsconfig.json`
     reference are preserved or superseded by an equivalent
   - `87d452e0` — no dangling project reference reappears in `tsconfig.json`
5. **Run validation** across the workspace, commit, open the PR.

## Validation Steps

- `npm install` clean; `npx prettier --check .` passes
- `npx moon run config-eslint:build config-typescript:build config-vitest:build`
  (adjust to actual project names via `moon query projects --tags config`)
- Full workspace: `npx moon run :typecheck :lint :build :test` for affected
  projects — the whole point is that main still passes with the new templates

## Success Criteria

- [ ] All four mapped main-commit intents verified present
- [ ] Deferred-gate list documented in the PR (empty is fine)
- [ ] Workspace CI green on main's existing content with the new configs
- [ ] Lockfile regenerated, not copied

## Files Modified/Created

- `packages/config/**`, `.moon/tasks/*.yml`, `.moon/toolchains.yml` (partial),
  `.moon/scripts/**`, `.devcontainer/**`, `.github/workflows/**`, root config
  files, `package-lock.json` (regenerated)

## Dependencies

None — runs parallel with subplan 01.

## Estimated Duration

2-4 hours.
