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

# Subplan 07: Skill Slice — Domain & Engine Guides

## Objective

Migrate the domain/engine skill packages — second of three `packages/skill`
chunks. Same mechanics as subplan 06; disjoint package set, so both run in
parallel. Includes the domain skills the branch ADDED, which have no
counterpart on main.

## Carried From Subplan 06

- **Hold the catalog version at 5.1.0.** The donor has all 72 skill packages at
  7.0.0 in lockstep. Bumping only this chunk splits the catalog and breaks the
  exact-version pins overlays place on their bases. Revert `version` in
  `package.json`, `.claude-plugin/plugin.json` and `.codex-plugin/plugin.json`,
  and any `@xonovex/skill-*` pin, back to 5.1.0. Subplan 10 bumps the catalog
  once.
- **Apply the donor's deletions by hand.** `git checkout <branch> -- <paths>`
  adds and modifies but never deletes. Every skill package loses its
  `prettier.config.ts` (the donor keeps only the root config), and several lose
  reference files whose content moved. Compute them with
  `git diff --name-status main <donor> -- <chunk paths>` and remove the `D`
  entries.
- **Check handoffs before committing.** Run
  `npx moon run script-moon-skill-validate-links:composition-check`. A guide
  naming a skill outside the catalog fails it; subplan 06 had to pull
  `skill-accessibility` forward for that reason.
- **Verify against the donor with a working-tree diff.**
  `git diff <donor> -- <chunk paths>` must be empty. `main..HEAD` reads empty
  until the slice is committed and proves nothing.

## Tasks

1. **Fix chunk membership**: `skill-audio`, `skill-cross-platform`,
   `skill-data-model`, `skill-data-oriented-design`, `skill-debugging`,
   `skill-ecs`, `skill-editor-viewport`, `skill-game-networking`,
   `skill-gpu-rendering`, `skill-gpu-rendering-vulkan`, `skill-imgui`,
   `skill-lock-free`, `skill-memory-management`, `skill-node-graph`,
   `skill-asset-pipeline`, `skill-accessibility`, `skill-credential-management`
   — verify against the overlay/handoff graph with subplans 06/08; new-on-branch
   packages (e.g. `skill-editor-viewport`, `skill-accessibility`) are flagged as
   NEW in the PR description.
2. **Create the slice branch and checkout the chunk** (same pattern as
   subplan 06, `migrate/skill-domain-guides`).
3. **Resurrection audit** against `b01d38ab` via `git status`.
4. **Verify main-side intents on chunk paths**: `66b9ad55`, `6b332e83`,
   `22f48559`, `87d452e0` (the generic catalog-wide intents restricted to this
   chunk's packages).
5. **Leave registration files untouched** — new packages ship unregistered until
   subplan 10 reconciles the marketplaces; confirm nothing in the chunk breaks
   the build while unregistered (if the catalog validators demand registration,
   record the exception for subplan 10 and gate registration there).
6. **Run validators scoped to the chunk + workspace-wide**, commit, open the PR.

## Validation Steps

- Per-package skill tests via moon for every chunk package
- Catalog validators green with the mixed catalog; exceptions recorded for
  subplan 10
- Trigger-eval and sources checks pass for chunk skills

## Success Criteria

- [x] Chunk membership documented; NEW packages flagged — see below
- [x] Applicable mapped intents verified on chunk paths — see below
- [x] No registration-file changes in the diff
- [x] Workspace validators green — `:ci-check --force`, 801 tasks, exit 0,
      nothing cached; composition-check reports 272/272 handoffs, 1491/1491
      links and 74 manifest pairs

## Chunk Membership

Sixteen packages migrated. Two corrections to the list above:

- `skill-accessibility` landed with subplan 06, which needed it to resolve a
  handoff from `react-guide` and `astro-guide`. It is already at donor state
  and is not re-migrated here.
- `skill-editor-viewport` is described above as new on the branch; it already
  exists on `main`, so it is an ordinary content migration.

**NEW:** `skill-credential-management` exists only on the donor. It ships
unregistered — no marketplace entry — because registration files are reconciled
once, in subplan 10. Nothing in the build depends on registration, and the
catalog validators pass without it.

## Intent Verification

- `66b9ad55` — no vendor or tenant specifics in chunk guides. Present.
- `6b332e83` — trigger-eval and sources coverage: all sixteen packages carry
  both `eval-queries.json` and `SOURCES.md`. Present.
- `87d452e0` — no phantom trigger words: every skill named by a chunk eval
  query resolves to a package that exists. Present.
- `22f48559` — not applicable; the authoring guidance it corrected lives in
  `skill-skill`, which is subplan 08's.

## Notes

- No dangling handoffs this time, so no boundary adjustment was needed — the
  chunk's overlays pin only `skill-gpu-rendering`, which is inside the chunk.
- `git diff <donor> -- <paths> --name-only` silently treats `--name-only` as a
  pathspec and diffs the whole tree. The flag has to precede the commit:
  `git diff --name-only <donor> -- <paths>`.

## Files Modified/Created

- `packages/skill/skill-<domain-and-engine packages>/**` only

## Dependencies

Subplan 04. Runs parallel with subplan 06 (disjoint package sets).

## Estimated Duration

4-8 hours.
