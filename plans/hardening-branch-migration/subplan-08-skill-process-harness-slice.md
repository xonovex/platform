---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-5
status: pending
dependencies:
  plans:
    [
      plans/hardening-branch-migration/subplan-06-skill-language-guides-slice.md,
      plans/hardening-branch-migration/subplan-07-skill-domain-guides-slice.md,
    ]
  files: [packages/skill]
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - skill-guide
  - command-guide
  - moon-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 08: Skill Slice — Process & Harness Guides

## Objective

Migrate the remaining skill packages: process/workflow guides, harness-adapter
skills, and the skill-package `AGENTS.md`/`CLAUDE.md`. This chunk closes the
`packages/skill` delta and carries the highest intent-verification load — most
of main's 25 commits touched these packages — plus the caveman/fable
resurrection risk.

## Gates Inherited from Subplans 02 and 04

Everything below reads the skill catalog, so it stays deferred until this slice
lands it. Re-enable each and confirm `npx moon run :ci-check --force` is green.

1. `.moon/tasks/tag-skill.yml` — return `skill-validate` and
   `skill-audit-sources` to `ci-check`'s deps and drop the `runInCI: false` on
   both. `skill-validate` already runs `moon-skill-validate-spec --strict` with
   the routing dependency; it currently reports findings against the catalog.
2. `packages/script/script-moon-skill-validate-routing/moon.yml` — return
   `routing-check` to `ci-check` and drop `runInCI: false`. It fails today
   because `gitlab-guide` owns no validation-split routing scenario: another
   skill must carry one of its validation positives as a negative. Main commit
   `c3b920d2` retargeted those provider queries deliberately, so reconcile that
   intent against the donor's query set rather than overwriting blindly.
3. `packages/script/script-moon-skill-validate-drift/moon.yml` — return
   `drift-check` to `ci-check` and drop `runInCI: false`. Enforce mode reports
   36 findings across 598 catalog and command files today.
4. Un-skip the catalog-reading tests and restore the floors lowered to match
   them:
   - `script-moon-skill-validate-spec/src/workflow-contracts.test.ts` — the
     whole suite, which reads the workflow command and skill files
   - `script-moon-skill-validate-spec/src/template-assets.test.ts` — both
     template cases
   - `script-moon-skill-validate-spec/src/index.test.ts` — three tests reading
     live skills
   - `script-moon-skill-eval-common/src/routing-catalog.test.ts` — the
     catalog-wide validation-routing test
   - `script-moon-skill-validate-spec/moon.yml` — functions floor 89 back to 90
   - `script-moon-skill-validate-spec/vitest.config.ts` — per-file floors for
     `src/validate-skill.ts` back to statements 70, branches 50, functions 90,
     lines 75
5. Restore the runtime probe evidence recorded earlier in this document.

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

1. **Fix chunk membership**: everything in `packages/skill` not taken by
   subplans 06/07 — process guides (`skill-plan`, `skill-git`, `skill-github`,
   `skill-gitlab`, `skill-pull-request`, `skill-code-review`,
   `skill-code-quality`, `skill-testing`, `skill-tdd`, `skill-bdd`, `skill-ddd`,
   `skill-fdd`, `skill-fp`, `skill-oop`, `skill-connascence`,
   `skill-hexagonal-pattern`, `skill-microkernel-pattern`,
   `skill-orthogonal-pattern`, `skill-user-stories`, `skill-skill`,
   `skill-command`, `skill-instruction`, `skill-content`, `skill-reflect`,
   `skill-versioning`, `skill-workflow`), harness adapters
   (`skill-claude-code`, `skill-codex`, `skill-copilot`, `skill-kiro`,
   `skill-opencode`, `skill-pi`, `skill-llmstxt`), and
   `packages/skill/AGENTS.md`, `packages/skill/CLAUDE.md`. Compute the closing
   set mechanically:
   `git diff --name-only main composable-workflow-platform-hardening -- packages/skill`
   minus chunks 06/07.
2. **Create the slice branch and checkout the chunk**
   (`migrate/skill-process-harness`).
3. **Resurrection audit — CRITICAL**: `b01d38ab` removed `skill-caveman` and
   `skill-fable` on main; the donor branch still has them. They must NOT appear
   in `git status`. Exclude them from the checkout and confirm.
4. **Verify main-side intents on chunk paths**:
   - `c3b920d2`, `67fbb767` — provider-guide trigger queries and issue/board
     coverage
   - `2ad6505e` — provider guides do not model plaintext token handling
   - `27a4319c` — eval infrastructure failures not scored as non-triggers
   - `0d020e3e` — destructive worktree removals gated on preview + confirmation
   - `9ab3bd7d` — plan lifecycle banners include the accept gate
   - `a90a569a` — skill-ablate folded into skill-optimize's verify phase (the
     skill side; the command side is subplan 09)
   - plus catalog-wide: `66b9ad55`, `6b332e83`, `22f48559`, `87d452e0`
5. **Restore the runtime probe evidence** on the harness adapters. The donor
   branch records `Observed runtime | Not installed in the validation
environment` for codex, copilot, opencode and pi, and drops the probe
   paragraph from `skill-kiro`'s onboarding reference. Taking the donor state
   verbatim therefore discards credentialed probe results that exist only on
   `composable-workflow-implementations-merge`, anchored by the tag
   `salvage/runtime-probes-d1692d3e` (commit `d1692d3e`). After the checkout,
   reapply to each `<harness>-guide/references/capabilities.md` capability
   table:

   | Skill            | Observed runtime        | Probed       |
   | ---------------- | ----------------------- | ------------ |
   | `skill-codex`    | `codex-cli 0.144.4`     | `2026-07-16` |
   | `skill-copilot`  | `0.0.377 (Copilot CLI)` | `2026-07-16` |
   | `skill-opencode` | `1.14.30`               | `2026-07-16` |
   | `skill-pi`       | `0.80.2`                | `2026-07-16` |

   Carry the accompanying caveat with the values: the probe observed the
   installed CLI version on a host with a working credentialed install and
   exercised the guard exit-code contract locally; native hook registration was
   not exercised, so hook-level rows stay documentation-verified and must not be
   reported as runtime conformance. Take `skill-kiro`'s probe runbook from the
   same tag; kiro stays a candidate (not installed). Leave each guide's
   `Documentation snapshot` at the donor's value — only the observed-runtime
   rows are restored.

6. **Close deferred gates**: any validators subplans 04/06/07 deferred to "when
   the full catalog is migrated" are enabled and must pass now, except
   marketplace-registration checks (subplan 10).
7. **Run full validation**, commit, open the PR.

## Validation Steps

- Per-package skill tests via moon for every chunk package
- ALL catalog validators green — after this PR the skill catalog on main equals
  the branch's, so no mixed-catalog exceptions remain except registration
- Trigger-eval + sources coverage complete for every skill (`6b332e83` intent)

## Success Criteria

- [ ] `git diff main composable-workflow-platform-hardening -- packages/skill`
      empty after merge (except caveman/fable, which stay deleted)
- [ ] caveman/fable NOT resurrected
- [ ] Runtime probe evidence restored for codex, copilot, opencode and pi; kiro
      probe runbook present
- [ ] All mapped intents verified
- [ ] All deferred catalog gates closed (registration excepted)

## Files Modified/Created

- Remaining `packages/skill/**` packages and package-level docs

## Dependencies

Subplans 06 and 07 (closes the catalog they started).

## Estimated Duration

6-10 hours (largest verification load).
