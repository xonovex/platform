---
type: plan
has_subplans: false
parent_plan: ../composition-layer-completion.md
parallel_group: 1
status: pending
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
  type_check: not_run
  lint: not_run
  build: not_run
  tests: not_run
updated: "2026-07-17"
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
