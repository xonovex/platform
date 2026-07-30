---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-1
status: in_progress
dependencies:
  plans: []
  files: [package.json, package-lock.json, packages/skill]
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - npm-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pending
---

# Subplan 01: Salvage Side-Branch Fixes

## Objective

Rescue the 4 commits unique to `composable-workflow-implementations-merge` before
that branch is deleted: the npm-audit dependency fix and the runtime-probe
documentation. Ship as one small PR so the branch can be retired.

## Tasks

1. **Create the slice branch** from local `main`:
   `git fetch origin && git checkout -b migrate/salvage-side-branch main`.
   Branch from local `main`, not `origin/main`: `origin/main` sits at
   `166c4f26`, which is the merge base, so all 25 main-side commits in the
   parent plan's intent-preservation map are local and unpushed. Branching from
   the remote ref would silently drop every commit this migration exists to
   preserve. The same applies to all later slices until `main` is pushed.
2. **Cherry-pick the dependency fix**:
   `git cherry-pick 052865b1` (fix(deps): clear npm audit advisories via
   transitive updates). On lockfile conflict, abort the file-level merge and
   regenerate instead: take main's `package.json` + the commit's intent, run
   `npm install`, then `npm audit` to confirm the advisories are actually clear
   against today's registry state.
3. **Hand the runtime-probe docs to subplan 08** — they cannot land here. All
   11 paths in `d1692d3e` are absent from `main`, so there is nothing to patch:
   - Five (`skill-aws`, `skill-azure-devops`, `skill-bitbucket`,
     `skill-bitrise`, `skill-datadog` provider-conformance references) belong to
     skills the donor branch deliberately culled in `ea0a33d4`. Porting them
     would resurrect removed skills — the parent plan's top risk. Dropped, not
     salvaged.
   - Five (`skill-codex`, `skill-copilot`, `skill-kiro`, `skill-opencode`,
     `skill-pi`) belong to harness adapters that arrive on `main` in subplan 08,
     which names them explicitly. Their evidence is real salvage: the donor
     branch regresses these rows to `Not installed in the validation
     environment`, so taking donor state verbatim would discard credentialed
     probe results. Anchored by the tag `salvage/runtime-probes-d1692d3e` so
     branch deletion in subplan 08 cannot destroy it; subplan 08 task 5 carries
     the values and subplan 10 guards the tag.
   - One (`plans/composable-workflow-phases/VALIDATION.txt`) is plan
     bookkeeping, excluded by this subplan's own skip rule.

   Skip the two `chore(plan)` bookkeeping commits — they reference subplans that
   don't exist on main.
4. **Verify no resurrections**: `git status` must show no files that main's
   removal commits (`708dfa1a`, `b01d38ab`) deleted.
5. **Run validation** (see below), commit with conventional messages, open the
   PR.

## Validation Steps

- `npm install` succeeds with a clean tree afterwards
- `npm audit` reports no advisories at the fixed severities
- `npx moon run :lint :typecheck :build :test` on affected projects (moon
  affected detection: `npx moon run :ci --affected` if configured, otherwise the
  touched projects' tasks)

## Success Criteria

- [x] `npm audit` clean for the advisories `052865b1` addressed — all six
      (`@babel/core`, `js-yaml`, `markdown-it`, `form-data`, `linkify-it`,
      `undici`) are gone; the workspace went from 15 advisories to 6
- [x] Runtime-probe docs resolved — not portable here, handed to subplan 08 and
      anchored by `salvage/runtime-probes-d1692d3e` (see task 3)
- [x] No resurrected files; no unrelated diff noise — all 27 paths deleted by
      `708dfa1a` / `b01d38ab` verified absent; floating-range lockfile drift
      (`yargs`, `@types/node`) reverted so the diff stays scoped
- [ ] PR merged via normal flow

## Outcome

Two commits, both lockfile-only, no `package.json` or direct-dependency changes:

- `9dbe83b5` — cherry-picked `052865b1`; the lockfile auto-merged cleanly, and
  only `plans/composable-workflow-phases/VALIDATION.txt` conflicted (modify/delete),
  dropped under this subplan's skip-bookkeeping rule.
- `27cc427f` — clears `fast-uri` and `postcss`, both disclosed after the
  salvaged pass and both non-breaking.

Six high advisories remain, all in one chain: `brace-expansion` →
`minimatch` → `glob` → `@mark.probst/typescript-json-schema` →
`quicktype-typescript-input` → `quicktype`. Clearing it requires
`quicktype@26.0.0`, a breaking major bump, which is out of scope for a
lockfile-only salvage slice and needs its own change.

Validation: `npx moon run :typecheck :lint :build :test` — 291 tasks, exit 0.

## Files Modified/Created

- `package.json`, `package-lock.json` (regenerated)
- Runtime-probe doc files under `packages/skill/**` (exact list from
  `git show d1692d3e --stat`)

## Dependencies

None — runs parallel with subplan 02.

## Estimated Duration

1-2 hours.
