---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-1
status: pending
dependencies:
  plans: []
  files: [package.json, package-lock.json, packages/skill]
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - npm-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 01: Salvage Side-Branch Fixes

## Objective

Rescue the 4 commits unique to `composable-workflow-implementations-merge` before
that branch is deleted: the npm-audit dependency fix and the runtime-probe
documentation. Ship as one small PR so the branch can be retired.

## Tasks

1. **Create the slice branch** from up-to-date main:
   `git fetch origin && git checkout -b migrate/salvage-side-branch origin/main`
2. **Cherry-pick the dependency fix**:
   `git cherry-pick 052865b1` (fix(deps): clear npm audit advisories via
   transitive updates). On lockfile conflict, abort the file-level merge and
   regenerate instead: take main's `package.json` + the commit's intent, run
   `npm install`, then `npm audit` to confirm the advisories are actually clear
   against today's registry state.
3. **Port the runtime-probe docs** from `d1692d3e` (docs(skill): record runtime
   probes for four harnesses and add probe runbooks): inspect
   `git show d1692d3e --stat` and checkout only its doc paths from
   `composable-workflow-implementations-merge`. Skip the two `chore(plan)`
   bookkeeping commits — they reference subplans that don't exist on main.
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

- [ ] `npm audit` clean for the advisories `052865b1` addressed
- [ ] Runtime-probe docs present and passing any skill/docs validators on main
- [ ] No resurrected files; no unrelated diff noise
- [ ] PR merged via normal flow

## Files Modified/Created

- `package.json`, `package-lock.json` (regenerated)
- Runtime-probe doc files under `packages/skill/**` (exact list from
  `git show d1692d3e --stat`)

## Dependencies

None — runs parallel with subplan 02.

## Estimated Duration

1-2 hours.
