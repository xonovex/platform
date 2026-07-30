---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-4
status: complete
dependencies:
  plans:
    [plans/hardening-branch-migration/subplan-04-script-validation-slice.md]
  files: [packages/skill]
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - skill-guide
  - moon-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
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

- [x] Chunk membership documented and handoff-consistent — the 23 listed
      packages all exist on both sides, plus `skill-accessibility` pulled in
      (see below)
- [x] All applicable mapped intents verified on chunk paths — see below
- [x] No registration-file changes in the diff
- [x] Workspace validators green with the mixed catalog —
      `:ci-check --force`, 797 tasks, exit 0, nothing cached;
      composition-check reports 268/268 handoffs, 1507/1507 links and 73
      manifest pairs

## Chunk Membership

The 23 named packages needed no boundary change: each exists on both sides, so
this chunk adds and removes no skill. One package was pulled forward:
`skill-accessibility`. Both `react-guide` and `astro-guide` gain a reference
naming **accessibility-guide** as the owning skill, and a handoff to a skill
outside the catalog dangles, which `composition-check` fails on. Its
marketplace registration is deliberately not added here; that stays with the
other registration files in subplan 10.

## Catalog Version Held at 5.1.0

The donor has all 72 skill packages at 7.0.0 in lockstep. Taking that for 23 of
them splits the catalog and breaks two things at once: the repo's lockstep rule,
and the exact-version pins overlays place on their bases — `c99-opinionated`
pins `@xonovex/skill-c99`, `hono-opinionated` pins `@xonovex/skill-hono`, and
so on. The version fields and those base pins are therefore held at main's
5.1.0, leaving the whole-catalog bump to subplan 10, which the parent plan
already reserves for catalog version reconciliation.

## Intent Verification

- `f121e7e7` — c99 build policy deduped: `c99-guide` owns
  `build-and-warnings.md`, and the opinionated overlay's
  `build-warnings-policy.md` opens by delegating to it rather than restating
  it. Present.
- `66b9ad55` — no vendor or tenant specifics in chunk guides. Present.
- `6b332e83` — trigger-eval and sources coverage: every chunk package has both
  `eval-queries.json` and `SOURCES.md`. Present.
- `87d452e0` — no phantom trigger words: every chunk package carries eval
  queries. Present.
- `22f48559`, `6060d9ac`, `c3b920d2` — not applicable to this chunk;
  `skill-skill`, the android guides and the provider guides all live in later
  chunks.

## Findings

- A path checkout never deletes, so 36 donor-side deletions had to be applied by
  hand: the per-package `prettier.config.ts` — which the donor drops across all
  72 skills in favour of the root config — and thirteen reference files whose
  content moved. Watch for the remaining 49 prettier configs in subplans 07 and 08.
- Comparing a slice against the donor must diff the working tree
  (`git diff <donor> -- <paths>`), not `main..HEAD`, which reads as empty until
  the slice is committed and hides whether the chunk actually matches.

## Files Modified/Created

- `packages/skill/skill-<language-and-framework packages>/**` only

## Dependencies

Subplan 04. Runs parallel with subplan 07 (disjoint package sets).

## Estimated Duration

4-8 hours.
