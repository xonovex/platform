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

# Subplan 07: Skill Slice — Domain & Engine Guides

## Objective

Migrate the domain/engine skill packages — second of three `packages/skill`
chunks. Same mechanics as subplan 06; disjoint package set, so both run in
parallel. Includes the domain skills the branch ADDED, which have no
counterpart on main.

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

- [ ] Chunk membership documented; NEW packages flagged
- [ ] Applicable mapped intents verified on chunk paths
- [ ] No registration-file changes in the diff
- [ ] Workspace validators green

## Files Modified/Created

- `packages/skill/skill-<domain-and-engine packages>/**` only

## Dependencies

Subplan 04. Runs parallel with subplan 06 (disjoint package sets).

## Estimated Duration

4-8 hours.
