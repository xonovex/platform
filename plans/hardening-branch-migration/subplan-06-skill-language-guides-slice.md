---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-4
status: pending
dependencies:
  plans: [plans/hardening-branch-migration/subplan-04-script-validation-slice.md]
  files: [packages/skill]
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - skill-guide
  - moon-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 06: Skill Slice — Language & Framework Guides

## Objective

Migrate the language/framework skill packages — the first of three chunks of
the 1,103-file `packages/skill` delta. This chunk contains no registration-file
changes and no membership changes (no adds/removals), so it can run in parallel
with subplan 07.

## Tasks

1. **Fix chunk membership first**: list this chunk's packages —
   `skill-c99`, `skill-c99-opinionated`, `skill-c99-game-opinionated`,
   `skill-cmake`, `skill-lua`, `skill-lua-opinionated`, `skill-typescript`,
   `skill-typescript-to-lua`, `skill-python`, `skill-shell-scripting`,
   `skill-sql-postgresql`, `skill-hono`, `skill-hono-opinionated`, `skill-zod`,
   `skill-vitest`, `skill-react`, `skill-astro`, `skill-threejs`,
   `skill-docker`, `skill-kubernetes`, `skill-terraform`, `skill-moon`,
   `skill-npm` — then verify against the overlay/handoff graph (an overlay like
   `c99-opinionated` must land with or after its base) and adjust boundaries
   with subplans 07/08 so no cross-chunk handoff dangles.
2. **Create the slice branch and checkout the chunk**:
   ```bash
   git checkout -b migrate/skill-language-guides origin/main
   git checkout composable-workflow-platform-hardening -- \
     packages/skill/skill-c99 packages/skill/skill-cmake ... (full list)
   ```
3. **Resurrection audit**: no files from `b01d38ab` (caveman/fable removal) can
   appear — they're outside this chunk, but `git status` confirms.
4. **Verify main-side intents on this chunk's paths** (restrict each check to
   chunk packages): `f121e7e7` (c99 build-policy dedupe + genericity),
   `66b9ad55` (no vendor/tenant specifics), `6b332e83` (trigger-eval + sources
   coverage), `22f48559` (no contradictory authoring guidance), `6060d9ac`
   (mirrored android query labels), `87d452e0` (no phantom trigger words),
   `c3b920d2` (provider trigger queries target the knowledge delta — applies if
   any provider guide files live in this chunk).
5. **Leave registration files untouched**: no edits to
   `.claude-plugin/marketplace.json`, `.agents/plugins/*`, or catalog versions
   (subplan 10 reconciles those once).
6. **Run the skill validators** from subplan 04 scoped to the chunk, plus any
   gates deferred to skill content; commit, open the PR.

## Validation Steps

- `npx moon run <skill-project>:test` for every chunk package
  (`moon query projects --tags skill`)
- Skill catalog validators (handoff/ownership/manifest/eval/sources from
  `b7d68e8a`, `3f04baaa`) green workspace-wide — they see old + new skills mixed
  and must still pass; if a validator requires whole-catalog consistency this
  chunk can't satisfy alone, record it for subplan 08 and gate it there
- Trigger-eval suites for chunk skills pass

## Success Criteria

- [ ] Chunk membership documented and handoff-consistent
- [ ] All applicable mapped intents verified on chunk paths
- [ ] No registration-file changes in the diff
- [ ] Workspace validators green with the mixed catalog

## Files Modified/Created

- `packages/skill/skill-<language-and-framework packages>/**` only

## Dependencies

Subplan 04. Runs parallel with subplan 07 (disjoint package sets).

## Estimated Duration

4-8 hours.
