---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-6
status: complete
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
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 09: Command Plugins Slice

## Objective

Migrate `packages/command` (command-utility, command-workflow — 70 files).
Lands after all skill chunks because commands are thin delegators to skills:
every command's target skill must already exist on main in its migrated form.
Carries main's command-removal history, which the donor branch predates.

## Carried From Subplan 08

The donor replaces the whole workflow command surface: `create`, `review`,
`revise`, `decide`, `execute`, `publish`, `abandon` and the `workspace-*` set,
in place of main's `plan-*` commands. Until that lands, main's commands name
operations the migrated skills no longer register, so this slice inherits:

1. `.moon/tasks/tag-command.yml` — `command-validate` carries `runInCI: false`
   and is out of `ci-check`. Restore both once the command surface matches the
   skills, and confirm the delegation contract resolves.
2. `script-moon-skill-validate-spec/src/workflow-contracts.test.ts` — five tests
   are skipped; they read
   `packages/command/command-workflow/commands/<command>.md`. Un-skip them.
3. `script-moon-skill-validate-drift:drift-check` reports two findings, both
   command files; `budgets.json` was raised for
   `command-utility/commands/skill-optimize.md` (172 to 187) and
   `slashcommand-distill.md` (113 to 169). Reset both to the donor's values once
   the donor's command files replace main's.
4. `9ab3bd7d` — the plan lifecycle banners must still name the accept gate;
   verify the intent survives in whatever the new surface calls it.

**Decide the fate of `plan-delegate`.** `7464a219` added a delegation
supervision operation and command after the merge base; the donor has no
counterpart. Subplan 08 preserved the skill half as
`skill-plan/plan-guide/references/delegate.md`, registered in the guide, with a
`budgets.json` entry. Its command half is still main's `plan-delegate.md`.
Either carry delegation into the new command surface, or retire the operation
and remove `delegate.md`, its guide entry and its budget line together.

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

- [x] No removed command resurrected — `version-bump.md` came back with the
      donor's command-utility and was removed again
- [x] All four mapped intents verified — see below
- [x] Every delegation target resolves — both packages pass the validator,
      17 and 22 commands
- [ ] `packages/command` does **not** match the donor, by decision — see below

## Decision: Keep the plan-* Commands

The donor replaces main's plan-prefixed workflow commands with a generic
lifecycle: `create`, `review`, `revise`, `decide`, `execute`, `publish`,
`abandon` and a `workspace-*` set. That replacement was declined; the
plan-prefixed surface stays. Consequences, all deliberate:

- `packages/command/command-workflow` keeps main's 17 commands and will not
  match the donor. Subplan 10's zero-diff check must treat this package as an
  accepted difference, not residue.
- The twelve plan commands were repointed at the operation names the migrated
  `plan-guide` registers: `plan-continue` now delegates to **continue**,
  `plan-subplans-create` to **expand**, and so on. The command names are
  unchanged; only the operation each names moved.
- Three operations the donor dropped were restored into `plan-guide` because
  kept commands delegate to them: `accept.md`, `decide.md` and `reject.md`,
  renamed to the guide's convention, registered in its progressive-disclosure
  list, and given budget entries.
- `plan-delegate` is resolved by the same decision: the command stays and its
  operation, preserved as `delegate.md` in subplan 08, is now reachable.
- Five `workflow-contracts.test.ts` tests stay skipped permanently rather than
  temporarily. They assert the generic surface exists; they become live only if
  it is adopted later.

## The Donor's Command Validator Landed Here

Subplan 04 kept main's `script-moon-command-validate` because the donor's
stricter version rejected main's command content. That content is now migrated,
so the donor's validator replaces it. It is stricter in three ways that required
changes to the kept commands:

- **Hyphen titles.** It expects `# /<namespace>:<command> - <title>`; main used
  an em dash. All 17 kept titles converted.
- **No em dashes in prose.** The same punctuation rule the skill catalog
  enforces. Rewritten across 20 files in `command-workflow`, using a full stop
  where a comma would splice.
- **Argument hints are parsed per token.** `[--mode red-team|pre-mortem|...]`
  reads as five separate arguments. `plan-critique` now hints `[--mode <mode>]`
  and documents the allowed values in its Arguments section.

Its `bin-permissions` task was re-added: the donor's copy lacks it, and the
compiled entry point loses its executable bit whenever `tsc` rewrites it.

## Other Donor Removals Applied

- `command-utility/.codex-plugin/plugin.json` — the donor drops the codex
  manifest from both command packages while keeping the skill ones. Predates the
  merge base, so deliberate; it does not conflict with `85968666`, which is
  about registering skills in the codex marketplace.

## Version Hold Extended to Commands

The command packages and their `@xonovex/skill-*` pins carried the donor's
7.0.0. With the catalog held at 5.1.0 those pins resolve to nothing on the
registry and `npm install` fails with a 404, so both manifests are held too.
Subplan 10 bumps commands and skills together.

## Intent Verification

- `15b5a21e` — thin-delegation validator: PASS over both packages.
- `9ab3bd7d` — plan lifecycle banners carry the accept gate: 11 plan commands
  name it, and the banner still reads
  `research → decide → create → revise ⇄ critique → accept → …`.
- `f121e7e7` — genericity holds; no vendor or tenant specifics in command text.
- `4668aade` — command versions stay consistent with the catalog, both held at
  5.1.0 pending subplan 10's lockstep bump.
- `708dfa1a` / `a90a569a` — all nine removed commands stay removed and no
  standalone `skill-ablate` command exists.

## Files Modified/Created

- `packages/command/**`

## Dependencies

Subplan 08 (full skill catalog on main first).

## Estimated Duration

2-4 hours.
