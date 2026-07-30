---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-1
status: complete
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
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
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

- [x] All four mapped main-commit intents verified present — `9dfb7522`'s
      `.prettierignore` entries for `*.golden.yaml` and `**/config/crd/bases/`
      had to be restored, the donor drops them; the other three survived intact
      (27 tsconfig references, 0 dangling)
- [x] Deferred-gate list documented — see Deferred Gates below
- [x] Workspace CI green on main's existing content with the new configs —
      `npx moon run :ci-check --force`, 696 tasks, exit 0, nothing cached
- [x] Lockfile regenerated, not copied

## Deferred Gates

Each gate below arrives with the donor's infrastructure but checks content that
is still main's, so it is held back here and belongs to the slice named.

| Gate | Held back to | Why |
| --- | --- | --- |
| `skill-validate` runs `moon-skill-validate-spec --strict` and depends on `script-moon-skill-validate-routing:routing-check` | 04 | Neither the binary nor the project exists on main; reverted to `moon-skill-validate` |
| `skill-audit-sources` in `ci-check` with `runInCI` enabled | 04 | The script reaches the network, which `.moon/AGENTS.md` forbids in `ci-check`, and the donor rewrote it (1,829 insertions across 12 files) |
| `coverage` in `tag-go.yml`'s `ci-check` (`GO_COVERAGE_MIN` 55) | 05 | `agent-cli-go` covers 44.8% on main; the donor's own `moon.yml` sets 81 because its tests arrive with `packages/agent` |
| `coverage` in `tag-typescript-script.yml`'s `ci-check` (50% floors) | 04 | Several script packages have no tests on main |
| `passWithNoTests` removed from the shared vitest config | 04 | `script-moon-version-detect` has no test file on main; the donor adds `src/detect.test.ts` |
| `noInlineConfig`, `no-warning-comments`, `sonarjs/cognitive-complexity: ["error", 30]` | 04 | 31 lint errors, all inside `packages/script`, which the donor rewrote wholesale |
| `.moon/toolchains.yml` `shellByTag` entries `ci: ci` and `docker: docker` | unassigned — see below | The root `flake.nix` defines no `ci` or `docker` devShell, and `agent-operator-go` carries the `docker` tag |
| `zizmor` step kept in `.github/workflows/ci.yml` | unassigned — see below | The donor moved it into a `.github/moon.yml` project, which is outside this slice's paths |

Note that the per-project coverage opt-ins on main are untouched:
`script-moon-command-validate` and the config packages declare their own
`ci-check: deps: [coverage]` with their own thresholds and still gate on them.

## Plan Gap Found

The root `flake.nix` and `.github/moon.yml` belong to no subplan. This slice
covers `.moon/` and `.github/workflows` but not either file, and two donor
changes depend on them — the `shellByTag` entries need the `ci` and `docker`
devShells, and the zizmor move needs the `.github` project. Assign both before
subplan 10 runs its zero-diff check, or that check will report them as residue.

## Files Modified/Created

- `packages/config/**`, `.moon/tasks/*.yml`, `.moon/toolchains.yml` (partial),
  `.moon/scripts/**`, `.devcontainer/**`, `.github/workflows/**`, root config
  files, `package-lock.json` (regenerated)

## Dependencies

None — runs parallel with subplan 01.

## Estimated Duration

2-4 hours.
