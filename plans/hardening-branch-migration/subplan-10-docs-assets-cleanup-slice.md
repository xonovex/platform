---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-7
status: pending
dependencies:
  plans: [plans/hardening-branch-migration/subplan-09-command-plugins-slice.md]
  files:
    [
      packages/diagram,
      packages/asset,
      plans,
      README.md,
      .claude-plugin/marketplace.json,
      .agents/plugins,
      .moon/toolchains.yml,
    ]
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - versioning-guide
  - skill-guide
  - moon-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 10: Docs, Assets, Reconciliation & Cleanup

## Objective

Close out the migration: diagrams, assets, plan documents, README, the single
marketplace/catalog reconciliation deferred by every earlier slice, the nix
toolchain reference bump from subplan 03's release, the zero-diff proof, and
removal of the donor branches and worktrees.

## Tasks

1. **Create the slice branch**:
   `git checkout -b migrate/docs-assets-cleanup origin/main`
2. **Checkout content paths**:
   `git checkout composable-workflow-platform-hardening -- packages/diagram packages/asset README.md`
   For `plans/`: bring the branch's plan documents over as historical records
   only if their status frontmatter is terminal (complete/rejected); drop
   stale in-progress plans that describe work this migration replaced.
3. **Reconcile registrations once** — intents `4668aade`, `85968666`,
   `708dfa1a`, `b01d38ab`:
   - `.claude-plugin/marketplace.json` and `.agents/plugins/marketplace.json`
     list exactly the post-migration catalog: all NEW skills from subplans
     06-08 registered, caveman/fable and the removed commands absent
   - bump the plugin catalog version in lockstep with the packages per the
     established policy; regenerate `package-lock.json`
4. **Bump the nix toolchain reference**: set `.moon/toolchains.yml` `plugin:` to
   the tag released by subplan 03; run a moon task end-to-end to prove the new
   WASM artifact loads.
5. **Zero-diff proof**:
   `git diff main composable-workflow-platform-hardening` — every remaining
   line must be an intentional non-migration (deliberate removals, dropped plan
   bookkeeping, regenerated lockfiles). Document the residue list in the PR;
   anything unexplained goes back to its owning slice.
6. **Verify `708dfa1a`'s diagram intent**: the diagram set matches the
   post-removal command catalog.
7. **After the PR merges — destructive cleanup, with user preview/confirmation
   per the worktree-removal gate**: delete branches
   `composable-workflow-platform-hardening` and
   `composable-workflow-implementations-merge` (local + origin) and remove the
   `../xonovex-platform-merge` worktree. `xonovex-platform-fable` and its
   worktree stay — hand off to the follow-up governance-skills plan. Keep the
   tag `salvage/runtime-probes-d1692d3e`: it is the only remaining anchor for
   the commit once `composable-workflow-implementations-merge` is gone, so
   confirm `git rev-list -1 salvage/runtime-probes-d1692d3e` still resolves
   after the deletions and push the tag to origin.

## Gates Inherited from Subplan 04

`packages/script/script-moon-release-validate` carries `runInCI: false` on its
`release-validate` task, and its end-to-end test in `src/validate.test.ts` is
skipped. Both fail today because the Codex marketplace still lists the command
plugins this migration removes. Once the marketplace and catalog files are
reconciled, return `release-validate` to that project's `ci-check`, drop the
`runInCI: false`, and un-skip the test.

`packages/script/script-moon-action-graph` is still on `main`. The donor deleted
it together with the diagrams that consume it, so its removal belongs here:
delete the package, `packages/diagram/diagram-moon-action`, and the root
`tsconfig.json` reference, and drop the per-project coverage floors added to
keep it green in the meantime.

## Toolchain Reference Bump — Preconditions

Subplan 03 bumped `moon_nix_toolchain` to 0.7.0 and left
`.moon/toolchains.yml` pointing at `moon_nix_toolchain-v0.6.1`, because a
consumer pin may only move once the tag assets exist. Before touching that
reference, confirm the release actually ran: `main` reached `origin`, a
`version packages` PR merged, and `moon_nix_toolchain-v0.7.0` and
`moon_nix_extension-v0.1.0` are published. `moon_nix_runtime` takes no tag — it
is a path dependency compiled into both plugin artifacts, carrying the `rust`
tag without `moon-plugin`.

## Validation Steps

- Full workspace `npx moon run :typecheck :lint :build :test`
- Marketplace validators (`85968666` intent) green; catalog lockstep check green
- A moon task runs successfully under the new toolchain plugin tag

## Success Criteria

- [ ] Zero-diff proof documented; no unexplained residue
- [ ] Marketplaces match the post-migration catalog exactly
- [ ] Toolchain reference on the new released tag, verified working
- [ ] Donor branches and merge worktree deleted (after confirmation);
      fable branch handed off to a follow-up plan
- [ ] Release flow used for any version bumps; no direct pushes

## Files Modified/Created

- `packages/diagram/**`, `packages/asset/**`, `plans/**`, `README.md`,
  `.claude-plugin/marketplace.json`, `.agents/plugins/marketplace.json`,
  `.moon/toolchains.yml`, `package-lock.json` (regenerated)

## Dependencies

Subplan 09 (everything else merged); subplan 03's release published.

## Estimated Duration

3-5 hours plus the cleanup pass.
