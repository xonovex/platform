---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-6
status: pending
dependencies:
  plans:
    [plans/hardening-branch-migration/subplan-08-skill-process-harness-slice.md]
  files: [packages/command]
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - command-guide
  - skill-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 09: Command Plugins Slice

## Objective

Migrate `packages/command` (command-utility, command-workflow — 70 files).
Lands after all skill chunks because commands are thin delegators to skills:
every command's target skill must already exist on main in its migrated form.
Carries main's command-removal history, which the donor branch predates.

## Tasks

1. **Create the slice branch**:
   `git checkout -b migrate/command-plugins origin/main`
2. **Checkout the donor path**:
   `git checkout composable-workflow-platform-hardening -- packages/command`
3. **Resurrection audit — CRITICAL**: `708dfa1a` removed the acceptance, pr,
   story-refine, and version-bump commands; `a90a569a` removed skill-ablate as a
   standalone command. The donor branch still has all of these. Remove every
   resurrected command file from the slice before committing
   (`git status` diff against `git show 708dfa1a --stat` and
   `git show a90a569a --stat`).
4. **Verify main-side intents**:
   - `15b5a21e` — every command passes the thin-delegation validator
   - `9ab3bd7d` — plan lifecycle banners carry the accept gate (command side)
   - `f121e7e7` — genericity holds in command text
   - `4668aade` — command package versions stay consistent with the catalog
     lockstep policy (actual bump handled in subplan 10)
5. **Cross-check delegation targets**: every skill a command delegates to
   exists on main post-subplan-08; no command references caveman, fable, or a
   pre-fold skill-ablate.
6. **Run validation** (thin-delegation validator especially), commit, open the
   PR.

## Validation Steps

- `npx moon run <command-project>:test` for both packages
  (`moon query projects --tags command`)
- Thin-delegation validator from `15b5a21e` green
- Catalog validators still green workspace-wide

## Success Criteria

- [ ] No removed command resurrected
- [ ] All four mapped intents verified
- [ ] Every delegation target resolves against main's skill catalog
- [ ] `git diff main composable-workflow-platform-hardening -- packages/command`
      empty after merge, modulo the deliberate removals

## Files Modified/Created

- `packages/command/**`

## Dependencies

Subplan 08 (full skill catalog on main first).

## Estimated Duration

2-4 hours.
