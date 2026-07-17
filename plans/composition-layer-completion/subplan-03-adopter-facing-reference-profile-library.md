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

1. Establish the library home and schema (see Open decisions): create a shipped-profile
   directory such as `packages/skill/skill-workflow/workflow-guide/assets/profiles/` (and,
   if governance facets warrant, `agent-governance-guide/agent-governance-guide/assets/profiles/`),
   distinct from `assets/fixtures/` (which `conformance.md` line 30 designates test inputs).
   The schema must satisfy the profile contract in both `profiles.md` files: identity,
   version, owner, scope, included capabilities, preserved results, topology,
   evidence/completion, axis requirements, actor/independence, enforcement guarantee,
   failure behavior; and the governance facets
   `lifecycle/governance/executor/enforcement/data/telemetry/distribution`.
2. Promote a curated set from the existing inline fixture profiles into real reference
   profiles, reusing `solo-profile`, `governed-profile`, `security-profile`,
   `supply-chain-profile`, and `code-review-profile` as starting points — meeting the
   coverage target chosen in Open decisions.
3. Define selection wiring: update the argument descriptions in
   `command-workflow/commands/workflow-inspect.md`, `workflow-conformance.md`,
   `workflow-onboard-advise.md`, `workflow-modules.md`, and `workflow-drift.md` so
   `--profile <reference>` resolves to a library entry, and state how
   `workflow-onboard-advise` recommends one.
4. Add a "Reference profiles" section to
   `packages/skill/skill-workflow/workflow-guide/references/profiles.md` and
   `packages/skill/skill-agent-governance/agent-governance-guide/references/profiles.md`
   listing the shipped profiles and how to select or extend one; cross-reference the
   contract text rather than duplicating it.
5. Add a reference-profile validator
   `packages/skill/skill-workflow/workflow-guide/scripts/validate-reference-profiles.mjs`
   that runs each shipped profile through `validateProfile` / the governance
   profile-composition rules and confirms it references only capabilities that are shipped
   or explicitly marked adopter-supplied — consulting the Phase 4 capability registry to
   decide "unresolvable".
6. Wire the validator into the `skill-workflow` `package.json` `test` script and add
   `workflow-guide/assets/profiles/*.json` to `moon.yml` `test` `inputs`; keep one owner per
   profile and keep `assets/fixtures/` test-only. Do not run prettier or npm.

## Acceptance criteria

- At least one shipped reference profile exists for the chosen coverage target (per adoption
  mode and/or per team shape), each validating against the profile contract via a CI-run
  script.
- `workflow-inspect`, `workflow-onboard-advise`, and `workflow-conformance` docs describe how
  `--profile <reference>` resolves to a library entry, and the library entries are real
  files under `assets/profiles/` (not fixtures).
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

## Open decisions

The parent leaves two decisions open for this phase:

- **(a) Library home and serialization.** Where the shipped profile library lives (e.g.
  `workflow-guide/assets/profiles/` and/or the governance skill) and its serialization
  format, kept distinct from `assets/fixtures/`. The parent flags the fixture-vs-library
  confusion risk but does not fix the location or format. State the choice; do not resolve
  it here.
- **(b) Coverage target.** Whether the minimum shipped set is one worked profile per
  adoption mode (workflow-only, governance-only, enablement-only, external-enforcement-only,
  integrated), per team shape (solo, small-team, regulated), or the cross-product of both.
  State the target; do not resolve it here.
