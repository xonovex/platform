---
type: plan
has_subplans: false
parent_plan: ../composition-layer-completion.md
parallel_group: 1
status: complete
dependencies:
  plans: []
  files:
    - packages/script/script-moon-skill-validate/src/*.ts
    - .moon/tasks/tag-skill.yml
    - packages/command/command-workflow/docs/*.md
    - packages/command/command-workflow/commands/*.md
skills_to_consult:
  - plan-guide
  - skill-guide
  - typescript-guide
  - testing-guide
  - moon-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
updated: "2026-07-18"
---

# Cross-Skill / Cross-Package Reference-Link Guard

## Objective

A composition doc's semantic reference to a contract in another skill/package is a fact a
check verifies, so a renamed or moved contract file cannot leave a dangling link in the
composition grammar. Today `checkReferenceFileLinks` only scans a single skill's own
`references/` dir and `command-workflow` (a `command`-tagged package) is never scanned by
`skill-validate`.

## Tasks

1. Generalize link validation to relative markdown links that cross skill/package
   boundaries: extend `packages/script/script-moon-skill-validate/src/reference-file-links.ts`
   (or add a sibling module under `packages/script/script-moon-skill-validate/src/`) to
   resolve `../`-escaping relative links against the repository root, while keeping the
   existing intra-skill `checkReferenceFileLinks` behavior unchanged.
2. Set the scan scope to `packages/command/command-workflow/docs/*.md`, every skill
   `SKILL.md` and `references/*.md`, and the `packages/command/command-workflow/commands/*.md`
   delegations.
3. Confirm coverage of the verified live cross-package links:
   `command-workflow/docs/architecture-and-composition.md` lines 49/57,
   `docs/developer-quickstart.md` lines 198-199, `docs/platform-onboarding.md` lines 7-11,
   and any other `](../../../skill/...)` target.
4. Preserve the existing exclusions: skip external (`http(s):`, `mailto:`) and placeholder
   (`<topic>.md`, `{topic}.md`, `…`) links, matching the current check's behavior.
5. Wire into CI: either extend the `skill-validate` task in `.moon/tasks/tag-skill.yml` or
   add a repo-root `moon` task so the `command`-tagged `command-workflow` package's docs are
   actually scanned (they are not today).
6. Extend `packages/script/script-moon-skill-validate/src/reference-file-links.test.ts` with
   cases for boundary-crossing resolution and for each skip rule.

## Acceptance criteria

- Temporarily renaming a cross-package target (e.g.
  `agent-governance-guide/references/actors.md`) makes the `moon` task fail with a message
  naming the source file and broken link.
- All currently-live cross-package links resolve — the check passes on `main` as-is once the
  targets are confirmed present.
- External (`http(s):` / `mailto:`) and placeholder (`<topic>.md`, `{topic}.md`, `…`) links
  are skipped, matching the existing check's exclusions.
- The new test in `reference-file-links.test.ts` fails before the guard is added and passes
  after.

## Dependencies

None. Phase 2 is independent of the other phases — the parent cites no cross-phase
prerequisite for the link guard. It can run concurrently with the vocabulary guard and the
capability registry.

## Decision (settled 2026-07-17)

**By-name cross-skill references are out of scope** (parent Decision 2). The guard validates
relative markdown links only; prose mentions — "see `execution.md`", bolded skill names — are
not parsed. Mechanical resolution keeps the guard trustworthy, and skill renames already fail
loudly via `marketplace.json` lockstep.

## Implementation notes (completed 2026-07-18)

`build`, `typecheck`, `lint`, and `tests` all pass via `moon` for
`script-moon-skill-validate` (24 vitest cases across 4 files); the guard resolves 64/64
cross-package links on `main`. Adversarial verification: renaming
`agent-governance-guide/references/actors.md` away makes `command-workflow:cross-package-links`
exit non-zero, naming the source (`architecture-and-composition.md`) and the exact broken
link; restoring the file returns the guard to 64/64.

What landed in `packages/script/script-moon-skill-validate/src/`:

- `reference-file-links.ts` — refactored to export the shared link primitives (`MD_LINK_RE`,
  `relativeLinkTarget`, `isFile`); `checkReferenceFileLinks` behavior is byte-identical (its
  five existing tests still pass).
- `cross-package-links.ts` — `checkCrossPackageLinks(repoRoot, report)` discovers the scan
  scope (every skill `SKILL.md` and `references/*.md`, plus `command-workflow/docs/*.md` and
  `commands/*.md`), and validates only links whose resolved target escapes the source file's
  own `packages/<layer>/<package>` root. External / placeholder / ellipsis / anchor forms are
  skipped, reusing `relativeLinkTarget`.
- `links-cli.ts` — the `moon-skill-links` bin: scans the repo root (cwd), prints PASS/FAIL,
  exits non-zero on any broken cross-package link.
- `cross-package-links.test.ts` — boundary-crossing resolve/break, intra-package ignored,
  each skip rule, in-page fragment, and an aggregate-scan case.

Wiring: a `cross-package-links` task on the `command-workflow` project runs the built CLI from
the workspace root and is added to that project's `ci-check`; its `inputs` list the full scan
scope so `moon ci --affected` re-runs it when any scoped file changes.

### Deviations from the tasks (with rationale)

- **Scope narrowed to boundary-crossing links, and `command-workflow/docs` was already
  covered.** Research the plan predated missed that `command-workflow/scripts/validate-documentation.mjs`
  (run by `command-workflow:test`) already resolves all links in `README.md`, `MIGRATION.md`,
  and `docs/**/*.md`, including cross-package ones. The genuine gaps were the
  `commands/*.md` delegations (never scanned) and cross-package links in skill `SKILL.md`
  (`index.ts`'s `REF_LINK_RE` matches only `references/…`). Validating only boundary-crossing
  links gives a clean division of labour — intra-package links stay owned by
  `checkReferenceFileLinks` (skill references) and `validate-documentation.mjs`
  (command-workflow docs) — instead of re-checking them.
- **New test file rather than extending `reference-file-links.test.ts`.** The new logic lives
  in `cross-package-links.ts`, so its tests are co-located in `cross-package-links.test.ts`;
  `reference-file-links.test.ts` stays focused on the (unchanged) intra-skill check.
- **Wired via a `command-workflow` task, not `tag-skill.yml`.** A repo-wide scan run once fits
  the composition hub (which already dependsOn the composed skills); running it per
  skill-tagged package would repeat the whole-repo scan and still skip the `command`-tagged
  `command-workflow`. The task invokes the built CLI directly (`node …/links-cli.js`) for
  deterministic execution; the `moon-skill-links` bin is kept for manual use and fresh
  `npm ci` installs.
