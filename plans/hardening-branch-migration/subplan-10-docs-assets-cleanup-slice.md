---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-7
status: in_progress
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
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
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

## Carried From Subplan 09

- **`packages/command/command-workflow` will not match the donor, by decision.**
  The plan-prefixed command surface is kept instead of the donor's generic
  lifecycle. Treat that package as an accepted difference in the zero-diff
  check, not residue.
- **Bump the catalog once, here.** Every skill and command package, and their
  `@xonovex/skill-*` pins, are held at 5.1.0 while the donor sits at 7.0.0.
  Bump them together, in lockstep with both marketplace files.
- **Reset two budgets.** `budgets.json` was raised for
  `command-utility/commands/skill-optimize.md` (172 to 187) and
  `slashcommand-distill.md` (113 to 169) while main's command files were still
  in place. Those files are now the donor's, so recheck both against the
  donor's values.
- **Register the unregistered packages.** `skill-accessibility` and
  `skill-credential-management` shipped without marketplace entries, and the
  seven harness and workflow skills subplan 08 added need entries too. Both
  marketplace files, alphabetical.
- **Budget entries added outside the donor's set** cover the operations kept
  for the plan surface: `plan-guide`'s `accept.md`, `decide.md`, `reject.md`,
  `delegate.md` and its grown `SKILL.md`. Keep them when reconciling.

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

- [x] Zero-diff proof documented; no unexplained residue, see below
- [x] Marketplaces match the post-migration catalog exactly, and
      `release-validate` passes at 1,713 checks over 74 lockstep packages
- [ ] **Blocked:** toolchain reference on the new released tag, see below
- [x] Catalog bumped to 6.0.0 in lockstep
- [ ] **Blocked:** donor branches and worktree removal, see below
- [x] No direct pushes; nothing was released

## Zero-Diff Proof

`git diff main composable-workflow-platform-hardening` leaves 311 files, every
one an intentional non-migration:

| Category                                                         | Files |
| ---------------------------------------------------------------- | ----- |
| Catalog version held at 5.1.0 while the donor sits at 7.0.0      | 228   |
| Plan-prefixed command surface kept by decision                   | 31    |
| This migration's own plan documents                              | 12    |
| Runtime probe evidence restored from the salvage tag             | 5     |
| `plan-guide` operations kept for the plan-prefixed commands      | 5     |
| Non-terminal donor plans dropped per the terminal-status rule    | 5     |
| `moon-nix-toolchain` 0.7.0 release prepared on main              | 5     |
| Bare package specifier removed from the tsconfig `paths`         | 5     |
| Repository policy files changed on main                          | 4     |
| Prettier formatting applied on main                              | 4     |
| Validator bug fix and dead-test removal                          | 3     |
| `bin-permissions` task and its dependency                        | 2     |
| `version-bump.md`, removed by `708dfa1a`, correctly still absent | 1     |
| Regenerated lockfile                                             | 1     |

Nothing is unexplained. The largest category is the version hold, which
collapses to zero the moment the catalog is bumped.

## Residue the Proof Caught

The proof is what found these; none were visible from the slice diffs:

- **Five donor deletions in `packages/agent`** that subplan 05's path checkout
  left behind: `agent-cli-go/scripts/validate.sh`,
  `agent-cli-go/test/feature-parity.md`, the operator's
  `internal/validator/repository.go` and its test, and
  `internal/webhook/agentpolicy_verdict_test.go`. All predate the merge base.
- **Seven paths no subplan owned:** `.dockerignore`, `.gitlab/`, `.gitignore`,
  `.github/dependabot.yml`, `CONTRIBUTING.md`, the root `AGENTS.md` and
  `packages/command/AGENTS.md`. The root `AGENTS.md` mattered most: main's
  still described the `diagram` package this slice deletes, while the donor's
  carries the punctuation and exact-pin policies the migrated validators
  enforce.
- **A dangling diagram link** in the kept `command-workflow/README.md`,
  pointing at the deleted agent workflow diagram. Only `--force` exposed it;
  the cached run reported `composition-check` as passing.

## packages/diagram Removed Entirely

The donor deletes all three diagram projects, not just the moon-action pair:
`2f9d258b` took `diagram-agent-workflow`, `a5d2b732` took
`diagram-moon-action` with the action-graph generator, and `e5439b9d` folded
the group into `packages/asset`, which now carries the sandbox-isolation
diagrams. The agent workflow diagram was dropped rather than folded, so the
README link to it went too.

## Blocked: Toolchain Bump and Cleanup

Neither can proceed from here, and both are gates rather than oversights.

**Toolchain reference.** `.moon/toolchains.yml` still points at
`moon_nix_toolchain-v0.6.1`. Retargeting it at `v0.7.0` was tried and fails
with `plugin::loader::github::asset_missing`: moon resolves the `github://`
locator against a GitHub release, and only `v0.6.1` has one, carrying
`moon_nix_toolchain.wasm` and its checksum. Publishing `v0.7.0` requires `main`
to reach `origin` and a `version packages` PR to merge, so the reference stays
until then. The crate itself is already at 0.7.0.

**Catalog version.** Bumped to 6.0.0 in lockstep across all 74 skill and
command packages, their manifests, both marketplace files, and the exact
`@xonovex/skill-*` pins. The level is major because the migration removes nine
skill plugins, which breaks anyone installing them from either marketplace.
`release-validate` passes at 1,713 checks over 74 lockstep packages. Nothing is
published; the release flow still decides when this ships.

**Branch and worktree removal.** Deleting
`composable-workflow-platform-hardening` and
`composable-workflow-implementations-merge`, and removing
`../xonovex-platform-merge`, is destructive and gated on explicit
confirmation. It should also wait until the migration is pushed: while `main`
is local-only, those branches are the sole copies of the donor state. Keep the
tag `salvage/runtime-probes-d1692d3e` and push it, since it becomes the only
anchor for that commit.

## Files Modified/Created

- `packages/diagram/**`, `packages/asset/**`, `plans/**`, `README.md`,
  `.claude-plugin/marketplace.json`, `.agents/plugins/marketplace.json`,
  `.moon/toolchains.yml`, `package-lock.json` (regenerated)

## Dependencies

Subplan 09 (everything else merged); subplan 03's release published.

## Estimated Duration

3-5 hours plus the cleanup pass.
