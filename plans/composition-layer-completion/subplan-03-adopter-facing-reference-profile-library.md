---
type: plan
has_subplans: false
parent_plan: ../composition-layer-completion.md
parallel_group: 2
status: pending
dependencies:
  plans: []
  files:
    - packages/skill/skill-workflow/workflow-guide/assets/profiles/*.json
    - packages/skill/skill-workflow/workflow-guide/scripts/*.mjs
    - packages/skill/skill-workflow/workflow-guide/scripts/*.ts
    - packages/skill/skill-workflow/workflow-guide/references/profiles.md
    - packages/skill/skill-workflow/package.json
    - packages/skill/skill-workflow/moon.yml
    - packages/skill/skill-agent-governance/agent-governance-guide/assets/profiles/*.json
    - packages/skill/skill-agent-governance/agent-governance-guide/references/profiles.md
    - packages/command/command-workflow/commands/workflow-inspect.md
    - packages/command/command-workflow/commands/workflow-conformance.md
    - packages/command/command-workflow/commands/workflow-onboard-advise.md
    - packages/command/command-workflow/commands/workflow-modules.md
    - packages/command/command-workflow/commands/workflow-drift.md
    - packages/command/command-workflow/commands/workflow-governance-inspect.md
skills_to_consult:
  - plan-guide
  - skill-guide
  - command-guide
  - orthogonal-pattern-guide
  - testing-guide
  - moon-guide
validation:
  type_check: not_run
  lint: not_run
  build: not_run
  tests: not_run
updated: "2026-07-17"
---

# Adopter-Facing Reference Profile Library

## Objective

Ship a curated, selectable set of reference profiles so `--profile <reference>` resolves to
a real, worked composition — not a test fixture. The commands already accept
`--profile <reference>`; nothing ships for a reference to point at. This phase closes that
grammar-rich, library-poor gap.

## Tasks

1. Create the per-plane library homes (parent Decision 3):
   `packages/skill/skill-workflow/workflow-guide/assets/profiles/` and
   `packages/skill/skill-agent-governance/agent-governance-guide/assets/profiles/`,
   prettier-formatted JSON, distinct from `assets/fixtures/` (which `conformance.md` line 30
   designates test inputs). An integrated reference profile is a workflow profile whose
   governance facet names its governance profile by identity; `--profile <reference>`
   resolves the name in the owning plane's library and follows the cross-reference.
   The schema must satisfy the profile contract in both `profiles.md` files: identity,
   version, owner, scope, included capabilities, preserved results, topology,
   evidence/completion, axis requirements, actor/independence, enforcement guarantee,
   failure behavior; and the governance facets
   `lifecycle/governance/executor/enforcement/data/telemetry/distribution`.
2. Promote a curated set from the existing inline fixture profiles into real reference
   profiles — five, one per adoption mode with the team shapes distributed across them
   (parent Decision 4): workflow-only/solo from `solo-profile`, integrated/small-team from
   `governed-profile`, a regulated governance-only or external-enforcement-only entry from
   `security-profile` / `supply-chain-profile`, and enablement-only from
   `code-review-profile`.
3. Define selection wiring: update the argument descriptions in
   `command-workflow/commands/workflow-inspect.md`, `workflow-conformance.md`,
   `workflow-onboard-advise.md`, `workflow-modules.md`, `workflow-drift.md`, and
   `workflow-governance-inspect.md` (which also takes `--profile <reference>`) so
   `--profile <reference>` resolves to a library entry, and state how
   `workflow-onboard-advise` recommends one.
4. Add a "Reference profiles" section to
   `packages/skill/skill-workflow/workflow-guide/references/profiles.md` and
   `packages/skill/skill-agent-governance/agent-governance-guide/references/profiles.md`
   listing the shipped profiles and how to select or extend one; cross-reference the
   contract text rather than duplicating it.
5. Add a reference-profile validator
   `packages/skill/skill-workflow/workflow-guide/scripts/validate-reference-profiles.ts`
   (TypeScript per parent Decision 1) that runs each shipped profile — both plane libraries,
   following integrated pairings — through `validateProfile` / the governance
   profile-composition rules and confirms it references only capabilities that are shipped
   or explicitly marked adopter-supplied — consulting the Phase 4 capability registry to
   decide "unresolvable".
6. Wire the validator into the `skill-workflow` `package.json` `test` script and add
   `workflow-guide/assets/profiles/*.json` plus the cross-package
   `agent-governance-guide/assets/profiles/*.json` to the `skill-workflow` `moon.yml` `test`
   `inputs` (cross-package inputs are precedented by the governance `moon.yml`); keep one
   owner per profile and keep `assets/fixtures/` test-only. Do not run prettier or npm.

## Acceptance criteria

- Five shipped reference profiles exist — one per adoption mode, with each team shape
  appearing at least once — each validating against the profile contract via a CI-run
  script.
- `workflow-inspect`, `workflow-onboard-advise`, `workflow-conformance`, and
  `workflow-governance-inspect` docs describe how `--profile <reference>` resolves to a
  library entry, and the library entries are real files under the per-plane
  `assets/profiles/` directories (not fixtures).
- The reference-profile validator fails if a shipped profile omits a required contract field
  or names an unresolvable capability.
- Files under `assets/fixtures/` remain test data; the shipped library is a separate,
  clearly-labeled directory.

## Dependencies

Phase 4 (capability-registry-and-placeholder-marking) must land first. The reference-profile
validator (task 5) must reject a profile that names a capability neither shipped nor marked
adopter-supplied, which requires the Phase 4 registry — the parent wires this both ways:
Phase 3's validator "references only capabilities that are shipped or explicitly marked
adopter-supplied (feeds Phase 4/5)" and Phase 4 "ensure the Phase 3 reference-profile
validator … consult this registry". A profile selects capabilities the registry classifies,
so the registry is the upstream artifact.

## Decisions (settled 2026-07-17)

- **(a) Library home and serialization** (parent Decision 3): per-plane libraries —
  `workflow-guide/assets/profiles/*.json` and `agent-governance-guide/assets/profiles/*.json`,
  prettier-formatted JSON, distinct from `assets/fixtures/`. Integrated profiles pair by
  semantic reference from the workflow profile's governance facet; the Phase 5 completeness
  check validates the pair as one assembled selection.
- **(b) Coverage target** (parent Decision 4): five profiles, one per adoption mode, with the
  three team shapes distributed across them; within-mode shape variants are documented
  extension, not shipped files.
