---
type: plan
has_subplans: true
status: complete
approved_date: "2026-07-17"
dependencies:
  plans: []
  files:
    - packages/skill/skill-agent-governance/agent-governance-guide/scripts/*.mjs
    - packages/skill/skill-agent-governance/agent-governance-guide/scripts/*.ts
    - packages/skill/skill-agent-governance/agent-governance-guide/references/*.md
    - packages/skill/skill-agent-governance/agent-governance-guide/assets/**/*.json
    - packages/skill/skill-workflow/workflow-guide/scripts/*.mjs
    - packages/skill/skill-workflow/workflow-guide/scripts/*.ts
    - packages/skill/skill-workflow/workflow-guide/references/*.md
    - packages/skill/skill-workflow/workflow-guide/assets/**/*.json
    - packages/skill/skill-plan/plan-guide/scripts/*.mjs
    - packages/skill/skill-plan/plan-guide/references/*.md
    - packages/command/command-workflow/commands/workflow-*.md
    - packages/command/command-workflow/docs/*.md
    - packages/script/script-moon-skill-validate/src/*.ts
    - .moon/tasks/tag-skill.yml
  subplans:
    - catalog-vocabulary-consistency-guard
    - cross-package-reference-link-guard
    - adopter-facing-reference-profile-library
    - capability-registry-and-placeholder-marking
    - assembled-composition-completeness-check
proposed_subplans:
  - catalog-vocabulary-consistency-guard
  - cross-package-reference-link-guard
  - adopter-facing-reference-profile-library
  - capability-registry-and-placeholder-marking
  - assembled-composition-completeness-check
parallel_groups:
  - group: 1
    plans:
      - catalog-vocabulary-consistency-guard
      - cross-package-reference-link-guard
      - capability-registry-and-placeholder-marking
  - group: 2
    plans:
      - adopter-facing-reference-profile-library
  - group: 3
    plans:
      - assembled-composition-completeness-check
skills_to_consult:
  - plan-guide
  - skill-guide
  - orthogonal-pattern-guide
  - connascence-guide
  - microkernel-pattern-guide
  - testing-guide
  - moon-guide
  - command-guide
  - instruction-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
updated: "2026-07-18"
---

# Composition-Layer Completion

## Objective

Close the gaps in the platform's "adoption is composable — you assemble the pieces"
thesis. This plan is about the **selection and assembly** of composable pieces: making
the composition grammar self-consistent under CI, shipping a real selectable profile
library, proving an assembled composition cannot be silently incomplete, and marking every
placeholder piece explicitly. It deliberately does **not** build runtime enforcement (hook
adapters, admission control, `A3` triggers, policy engines) — that is the sibling
[runtime-enforcement completion plan](runtime-enforcement-completion.md), cross-referenced
but not absorbed here.

The composition **grammar** is already strong. This plan credits and locks it in place, then
adds the missing guardrails and the missing shipped library so an adopter can select a
worked composition, inspect it, and be told — by a check, not by prose — when the pieces do
not fit.

## Decisions (settled 2026-07-17)

1. **New validators are TypeScript.** Every validator this plan adds is authored in
   erasable-syntax TypeScript executed directly by Node via native type stripping, matching
   the sibling runtime plan's Decision 1. The root `engines` Node floor raise (`>=20.0.0` →
   `>=22.18.0`, target Node 24 LTS) is a shared precursor owned by the runtime plan's
   Phase 1 — whichever plan executes first lands it first.
2. **By-name references are out of scope (Phase 2).** The link guard validates relative
   markdown links only; prose mentions ("see `execution.md`", bolded skill names) are not
   parsed — mechanical resolution keeps the guard trustworthy, and skill renames already fail
   loudly via `marketplace.json` lockstep.
3. **Profile library home and serialization (Phase 3).** Per-plane homes:
   `workflow-guide/assets/profiles/*.json` and
   `agent-governance-guide/assets/profiles/*.json`, prettier-formatted JSON, distinct from
   `assets/fixtures/`. An integrated reference profile is a workflow profile whose governance
   facet names its governance profile by identity; `--profile <reference>` resolves the name
   in the owning plane's library and follows the cross-reference, and the Phase 5
   completeness check validates the pair as one assembled selection.
4. **Coverage target (Phase 3).** Five reference profiles — one per adoption mode — promoted
   from the existing fixtures, with the three team shapes distributed across them
   (workflow-only/solo, integrated/small-team, a regulated governance-only or
   external-enforcement-only entry, enablement-only from `code-review-profile`). Within-mode
   shape variants are documented extension, not shipped files.
5. **The three unshipped methods are marked, not shipped (Phase 4).** `example-mapping`,
   `user-research`, and `architecture-review` are classified `adopter-supplied` in the
   registry, each naming the reference contract to satisfy; shipping any later is a registry
   flip plus a normal skill-create pass.

## Current state (VERIFIED)

### The grammar exists and is largely complete — credit it

- **Two independent planes composing via semantic references.**
  `packages/command/command-workflow/docs/architecture-and-composition.md` lines 5-12
  define the Workflow and Governance planes and state they "integrate through semantic
  references"; lines 26-35 give the six-step effective-profile composition algorithm; lines
  14-24 enumerate the five adoption modes with an explicit "expected absence report" per
  mode. This is a genuine, working composition grammar.
- **Provider ports.** The result-provider port is owned by the workflow plane in
  `packages/skill/skill-workflow/workflow-guide/references/providers.md` (seven required
  capabilities: `resolve`, `read`, `publish`, `revise`, `relate`, `version`,
  `capabilities`), with a self-controlled non-file conformance fake in
  `workflow-guide/scripts/conformance-helpers.mjs` (`createTaskSystemProvider`,
  `exerciseTaskSystemProvider`). Governance provider ports (policy / configuration /
  evidence) are in `agent-governance-guide/references/provider-contracts.md` and enforced by
  `validateProviderContract` (`agent-governance-guide/scripts/conformance-helpers.mjs`
  lines 336-380).
- **The profile contract exists in both skills.**
  `skill-workflow/workflow-guide/references/profiles.md` (workflow profile contract,
  composition rules 1-7) and
  `skill-agent-governance/agent-governance-guide/references/profiles.md` (governance profile
  facets and effective-composition steps 1-7). `validateProfile` exists in both skills'
  `conformance-helpers.mjs`.
- **Axis resolution by precedence.**
  `packages/skill/skill-plan/plan-guide/references/early-lifecycle-contracts.md` line 7:
  "explicit command argument, selected workflow profile, project instructions, unambiguous
  environment detection, axis-specific default, then a visible request for selection."
- **One-concept-one-owner skills.** `packages/skill/AGENTS.md` ("Composable split")
  mandates exactly one owner skill per concept, cross-referenced by name, never copied.
- **The compose/inspect command surface exists.**
  `packages/command/command-workflow/commands/` ships `workflow-inspect.md`,
  `workflow-conformance.md`, `workflow-onboard-advise.md`, `workflow-modules.md`,
  `workflow-drift.md`, and `workflow-governance-inspect.md`. Each delegates to the owning
  skill's operation and takes a `--profile <reference>` argument.

### Gap 1 — Composition integrity is guarded in spots, not catalog-wide

**Only two dedicated cross-site vocabulary guards exist**, plus one intra-skill link check:

- `agent-governance-guide/scripts/validate-executor-vocabulary.mjs` — validates the five
  executor classes across four declaring sites (`references/execution.md` table,
  `conformance-helpers.mjs` `expectedVocabulary.executorClasses`,
  `assets/conformance-fixtures.json`, and plan-guide's
  `validate-early-lifecycle-fixtures.mjs` `allowedExecutors`), plus the work-shape→executor
  selection function, with mutation guards. This is the reference pattern.
- `agent-governance-guide/scripts/validate-event-intent-vocabulary.mjs` — validates the
  event-intent families across the owner (`references/events-and-capabilities.md`), the
  adapter prose view (`command-workflow/docs/harness-capabilities.md`), and the diagram
  (`diagram-agent-workflow/workflow-diagram.dot`), with mutation guards.
- `packages/script/script-moon-skill-validate/src/reference-file-links.ts`
  (`checkReferenceFileLinks`) — validates markdown links **inside a single skill's own
  `references/*.md`** resolve to sibling files. Wired via `.moon/tasks/tag-skill.yml`
  `skill-validate`.

**The other shared vocabularies are declared in multiple artifacts but coupled only
fixture↔helper, never against prose tables, diagrams, cross-package docs, or a second
machine-read declaration.** `agent-governance-guide/scripts/validate-fixtures.mjs` line 57
calls `validateVocabulary(fixture)`, which compares `assets/conformance-fixtures.json`
(11 vocabulary arrays) against `expectedVocabulary` in `conformance-helpers.mjs`. That is the
_only_ consistency check for those 11 vocabularies, and it stays entirely inside the
governance skill. Verified multi-declaration sites that are **not** guarded:

- **`moduleClassifications`** — the six values are declared a **second time as an
  independent machine-read constant** in
  `agent-governance-guide/scripts/governance-operations-helpers.mjs` lines 17-24
  (`allowedModuleClassifications`), which does **not** import `expectedVocabulary`
  (verified: no import). Also enumerated as a prose table in `references/modules.md`
  line 25. Nothing couples these three declarations.
- **`policyOutcomes`** — enumerated as a prose backtick-list in
  `references/policy-and-authority.md` line 13, unguarded against the helper constant.
- **`authorityZones`** — second prose/table site `references/catalog-and-inventory.md`.
- **`adoptionModes`** — second prose/table sites `references/architecture.md`,
  `references/catalog-and-inventory.md`, `references/external-enforcement-onboarding.md`,
  **and cross-package** in `command-workflow/docs/architecture-and-composition.md`
  (Adoption modes table).
- **`profileFacets`** — enumerated as prose facet lists in _both_ `profiles.md` files;
  consumed by `validateProfile` but never compared to those docs.
- **`independenceLevels`** — private const in
  `workflow-guide/scripts/independence-helpers.mjs` lines 5-10 (not exported), independently
  re-stated as a table in `agent-governance-guide/references/actors.md` lines 32-36 and
  42-48. No coupling. The workflow plane has **no** `validateVocabulary` analog at all.
- **Work-shape literals** (`mechanical` / `bounded-transform` / `adaptive`) — owner
  `workflow-guide/scripts/development-assurance-helpers.mjs` (`selectDevelopmentExecutor`),
  also in `references/development-contracts.md` table and
  `assets/development-assurance-fixtures.json`. Partially exercised by the executor guard's
  work-shape selection, but the literal set is not compared across the three sites.

**Single-source vocabularies that need no guard** (verified only one declaring site):
`partyKinds` (`operational-lifecycle-helpers.mjs` line 18 only),
`allowedPolicyDomains` / `allowedDecisionProviders` (`governance-operations-helpers.mjs`),
`provenanceVerificationMethods`, `capabilityMatrixFields`, and `onboardingStages` (each
only in `conformance-helpers.mjs` + its fixture, already coupled by `validateVocabulary`).

**The docs already claim more than the code enforces.**
`agent-governance-guide/references/actors.md` line 23 asserts: "every closed vocabulary here
(executor classes, policy outcomes, module kinds, module classifications) is declared once
and validated against every declaring site." Only **executor classes** actually is; policy
outcomes, module kinds, and module classifications are not validated against their prose /
second-helper sites. The claim is currently false for three of its four examples.

**Cross-skill / cross-package reference links are validated nowhere.**
`checkReferenceFileLinks` only scans a skill's own `references/` dir, and `command-workflow`
is a `command`-tagged package that `skill-validate` never runs on. Verified live cross-package
links with no validator: `command-workflow/docs/architecture-and-composition.md` lines 49 and
57 link into `../../../skill/skill-agent-governance/.../actors.md` and two `SKILL.md` files;
`docs/platform-onboarding.md` lines 7-11 link into five other skills'
`references/automation-and-enforcement.md` / `SKILL.md`. If any target is renamed, these break
silently.

### Gap 2 — Grammar-rich, library-poor

**Named profiles exist only as test fixtures.** Verified with
`grep -rhoE '"[a-z-]*profile"' skill-workflow/workflow-guide/assets/`:
`governed-profile`, `solo-profile`, `security-profile`, `delivery-profile`,
`quality-profile`, `risk-profile`, `supply-chain-profile`, `code-review-profile` — all inline
objects inside `conformance-fixtures.json`, `development-assurance-fixtures.json`,
`operational-lifecycle-fixtures.json`, and two scenario fixtures. Governance has only
`additive-cross-facet-profile`. There is **no shipped, adopter-facing, selectable profile**.
The commands accept `--profile <reference>` (verified in `workflow-inspect.md`,
`workflow-onboard-advise.md`, `workflow-conformance.md`, `workflow-modules.md`,
`workflow-drift.md`, and `workflow-governance-inspect.md`) but nothing ships for a reference
to point at.

### Gap 3 — No proven "you can't assemble it wrong" check

Validation today is **per-contract**, not per-assembly.
`workflow-guide/references/conformance.md` validates a single claimed result, handle,
provider, or profile (lines 4-13); its required failures include "profile topology has
invalid edges or missing prerequisites" (single profile) and "a mandatory cross-plane
requirement has no adequate enforcement guarantee" (line 25). In code, `validateProfile`
checks one profile's facets/enforcement and `validateComposition`
(`agent-governance-guide/scripts/conformance-helpers.mjs` lines 219-225) only checks module
**conflict**. **No helper, contract, or fixture takes a whole assembled selection** (workflow
profile + governance profile + selected capabilities + providers + modules + methods) and
verifies every selected capability is present and compatible with no dangling references.
Yet `command-workflow/commands/workflow-conformance.md` already promises "Validate each plane
independently before validating **their composition**." — a promise with no backing contract.

### Gap 4 — Placeholder pieces in the composition graph

- **Selectable-but-unshipped methods.**
  `plan-guide/references/early-lifecycle-contracts.md` line 9 names "User stories, BDD,
  example mapping, user research, accessibility review, architecture review, and other
  installed skills" as selectable method capabilities. Verified: `skill-user-stories`,
  `skill-bdd`, and `skill-accessibility` ship, but `skill-example-mapping`,
  `skill-user-research`, and a general architecture-review skill do **not** exist
  (`skill-adr` covers decision records, not review). A profile that selects one of the
  three missing methods creates a dangling capability reference that no check flags and no doc
  marks as adopter-supplied.
- **`A3` autonomy is already marked correctly (the model to follow).**
  `agent-governance-guide/references/autonomy.md` line 9 states "**`A3` is the eventual goal,
  not a description of what exists.** Its triggers, admission control, and escalation routing
  are targets an adopter builds and proves". The `A3` **build** is runtime work owned by the
  sibling runtime plan; this plan only ensures such placeholders are _marked_, and credits
  `autonomy.md` as the existing pattern.

## Out of scope

- **Publishing and release.** Excluded by the maintainer. No version bumps, no
  `marketplace.json` lockstep release, no `:ci-publish`, no PR/push. Wiring new validators
  into `moon` test tasks is in scope; releasing them is not.
- **Runtime enforcement.** Building hook adapters, policy engines, admission control, CI
  gates, or the `A3` trigger/escalation machinery. Those are the sibling
  runtime-enforcement completion plan. This plan only _selects_, _marks_, and _validates the
  consistency of_ the pieces; it cross-references the runtime plan for anything that is a
  runtime build.
- **New lifecycle capabilities or new governance modules.** This plan adds guards, a profile
  library, and a completeness check over the _existing_ catalog; it does not invent new
  semantic contracts.
- **Changing the composition grammar itself.** The two-plane model, adoption modes, axis
  precedence, and profile contract are treated as correct and are locked in, not redesigned.

## Phases

### Phase 1 — Catalog-wide vocabulary-consistency guard

**Objective.** Every closed composition vocabulary is declared once and validated against
_every_ declaring site (machine-read constants, fixture arrays, prose tables, diagrams,
cross-package doc tables), so a catalog cannot be assembled with a vocabulary that has
drifted between two artifacts — generalizing the `validate-executor-vocabulary.mjs` pattern
to the whole catalog.

**Tasks.**

- Build a data-driven vocabulary registry (one new validator, e.g.
  `agent-governance-guide/scripts/validate-composition-vocabulary.ts` — TypeScript per
  Decision 1 — following the structure of `validate-executor-vocabulary.mjs`) that, for each vocabulary, names the
  canonical owner and lists every declaring site with an extractor (constant import, JSON
  key, prose-table regex, prose backtick-list regex, `.dot` label parse). Cover:
  `policyOutcomes`, `moduleKinds`, `moduleClassifications`, `authorityZones`,
  `adoptionModes`, `profileFacets`, and `independenceLevels`; include mutation guards per the
  existing precedent (an invented value, a renamed value, a dropped value, an unparseable
  site — each must be caught).
- Extract sites verified in Current state: `policy-and-authority.md` line 13
  (`policyOutcomes`); `modules.md` line 25 (`moduleClassifications`, `moduleKinds`);
  `catalog-and-inventory.md`, `architecture.md`, `external-enforcement-onboarding.md`, and
  `command-workflow/docs/architecture-and-composition.md` (`adoptionModes`); both
  `profiles.md` facet lists (`profileFacets`); `actors.md` lines 32-36
  (`independenceLevels`). The later-added `command-workflow/docs/adoption-map.md` is a
  further view site: its modes table (`adoptionModes`), axes-section facet list
  (`profileFacets`), and module-classification prose (`moduleClassifications`) join the new
  guard, and its executor-class and event-intent enumerations join the two existing guards'
  site lists.
- **Remove the second independent machine-read declaration**: make
  `governance-operations-helpers.mjs` `allowedModuleClassifications` import
  `expectedVocabulary.moduleClassifications` (per one-owner rule,
  `packages/skill/AGENTS.md`), OR register it as a declaring site the new validator compares.
  Prefer the import — connascence of value collapses to connascence of name.
- Give the workflow plane its own vocabulary consistency check for `independenceLevels` and
  the work-shape literals (owner: `independence-helpers.mjs` / `development-assurance-helpers.mjs`;
  views: `actors.md`, `development-contracts.md`, `development-assurance-fixtures.json`),
  since the workflow skill has no `validateVocabulary` analog today.
- **Make `actors.md` line 23 true**: after the guard covers policy outcomes, module kinds,
  and module classifications, the "validated against every declaring site" claim becomes
  accurate; if any listed vocabulary is intentionally left single-source, correct the
  sentence to match reality.
- Wire the new validator(s) into the `test` script of each affected skill's `package.json`
  and add the newly-read cross-package files to the corresponding `moon.yml` `test` `inputs`
  (the governance `moon.yml` already lists `command-workflow/docs/*.md` and the diagram).

**Acceptance criteria.**

- Running the affected skills' `test` tasks fails when any covered vocabulary value is
  changed in exactly one site (verified by temporarily mutating each site and observing a
  non-zero exit and a message naming the diverging site).
- Every mutation guard reports a failure when the comparison is defeated (dud-guard count is
  zero), matching the pattern in the two existing validators.
- `governance-operations-helpers.mjs` no longer holds an independent copy of
  `moduleClassifications` (verified: it imports the owner constant, or the guard covers it).
- `actors.md` line 23's list of "validated against every declaring site" vocabularies matches
  the set the guards actually cover — no false claim remains.
- The new validators run in CI via `moon` (present in `package.json` `test` and reachable
  from `ci-check`).

### Phase 2 — Cross-skill / cross-package reference-link guard

**Objective.** A composition doc's semantic reference to a contract in another skill/package
is a fact that a check verifies, so a renamed or moved contract file cannot leave a dangling
link in the composition grammar.

**Tasks.**

- Extend link validation to relative markdown links that cross skill/package boundaries.
  Either add a repo-level check (new small script under `packages/script/`, invoked from a
  `moon` task) or generalize `script-moon-skill-validate` to resolve `../`-escaping links
  against the repository root. Scope: `command-workflow/docs/*.md`, every skill `SKILL.md`
  and `references/*.md`, and the `command-workflow/commands/*.md` delegations.
- Cover the verified live links: `architecture-and-composition.md` lines 49/57,
  `developer-quickstart.md` lines 198-199, `platform-onboarding.md` lines 7-11, and any
  other `](../../../skill/...)` targets.
- By-name cross-skill references (e.g. prose "see `execution.md`" /
  "**agent-governance-guide** owns…") are out of scope per Decision 2 — the guard validates
  relative markdown links only.
- Keep the existing intra-skill `checkReferenceFileLinks` behavior unchanged; this phase only
  adds the boundary-crossing coverage it deliberately skips.
- Wire into CI via a `moon` task (reuse `tag-skill.yml` or add a repo-root task); ensure the
  `command`-tagged `command-workflow` package's docs are actually scanned.

**Acceptance criteria.**

- Renaming a cross-package target (e.g. `agent-governance-guide/references/actors.md`) makes a
  `moon` task fail with a message naming the source file and broken link (verified by a
  temporary rename).
- All currently-live cross-package links resolve (the check passes on `main` as-is after the
  targets are confirmed present).
- External (`http(s):` / `mailto:`) and placeholder (`<topic>.md`, `{topic}.md`, `…`) links
  are skipped, matching the existing check's exclusions.

### Phase 3 — Adopter-facing reference profile library

**Objective.** Ship a curated, selectable set of reference profiles so `--profile <reference>`
resolves to a real, worked composition — one per adoption mode and/or team shape.

**Tasks.**

- Per Decision 3, create the per-plane library homes
  `skill-workflow/workflow-guide/assets/profiles/` and
  `agent-governance-guide/assets/profiles/` (prettier-formatted JSON), distinct from
  `assets/fixtures/` (which `conformance.md` line 30 explicitly says are test inputs, not
  shipped profiles); an integrated profile pairs by semantic reference from the workflow
  profile's governance facet. The schema
  must satisfy the profile contract in both `profiles.md` files (identity, version, owner,
  scope, included capabilities, preserved results, topology, evidence/completion, axis
  requirements, actor/independence, enforcement guarantee, failure behavior; governance
  facets `lifecycle/governance/executor/enforcement/data/telemetry/distribution`).
- Promote a curated set from the existing fixture profiles into real reference profiles.
  Coverage per Decision 4: five profiles, one per adoption mode — workflow-only,
  governance-only, enablement-only, external-enforcement-only, integrated — with the team
  shapes solo, small-team, and regulated distributed across them (reuse `solo-profile`,
  `governed-profile`, `security-profile`, `supply-chain-profile`, `code-review-profile` as
  starting points).
- Define selection: how `workflow-onboard-advise` recommends one and how `workflow-inspect` /
  `workflow-conformance` / `workflow-governance-inspect` resolve `--profile <reference>` to a
  library entry. Update those command docs' argument descriptions to point at the library, and add a short
  "Reference profiles" section to `workflow-guide/references/profiles.md` (and the governance
  `profiles.md`) that lists the shipped profiles and how to select or extend one.
- Add a validator that every shipped reference profile passes the profile contract
  (`validateProfile` / the governance profile-composition rules) and references only
  capabilities that are shipped or explicitly marked adopter-supplied (feeds Phase 4/5). Wire
  it into the skill `test` task.
- Follow `packages/skill/AGENTS.md`: prettier-format new assets; keep one owner per profile;
  cross-reference rather than duplicate contract text.

**Acceptance criteria.**

- At least one shipped reference profile exists per adoption mode (five) and/or per team shape
  (three), each validating against the profile contract via a CI-run script.
- `workflow-inspect`, `workflow-onboard-advise`, and `workflow-conformance` docs describe how
  `--profile <reference>` resolves to a library entry, and the library entries are real files
  (not fixtures).
- The reference-profile validator fails if a shipped profile omits a required contract field
  or names an unresolvable capability.
- Fixtures under `assets/fixtures/` remain test data; the shipped library is a separate,
  clearly-labeled directory.

### Phase 4 — Capability registry and explicit placeholder marking

**Objective.** Every capability a profile can select is either shipped or explicitly marked
adopter-supplied, and the composition graph names its placeholders instead of leaving dangling
references.

**Tasks.**

- Enumerate the selectable capability namespace: lifecycle capabilities (from
  `workflow-guide/references/profiles.md` default presentation and the early-lifecycle
  contract), selectable **methods** (`early-lifecycle-contracts.md` line 9), provider ports,
  executor classes, and governance modules. Produce a single registry (data file or reference
  table) mapping each selectable capability to `shipped` (owner skill/module) or
  `adopter-supplied` (with the reference contract it must satisfy).
- Mark the verified placeholders explicitly: `example mapping`, `user research`, and
  `architecture review` in `early-lifecycle-contracts.md` line 9 are selectable methods with
  no shipped skill — either annotate them inline as adopter-supplied capabilities (naming the
  contract to satisfy) or ship a reference skill; do not leave them as bare names implying a
  shipped skill. Per Decision 5, all three are marked adopter-supplied, each naming the
  contract to satisfy.
- Credit and preserve the existing correct marking pattern: `autonomy.md`'s explicit "`A3` is
  the eventual goal, not a description of what exists". Cross-reference the sibling
  runtime-enforcement plan for the `A3` trigger/admission/escalation **build**; keep only the
  _marking_ here.
- Ensure the Phase 3 reference-profile validator and the Phase 5 completeness check consult
  this registry so a profile that selects an unmarked-and-unshipped capability fails.

**Acceptance criteria.**

- A capability registry exists that classifies every selectable capability as `shipped` or
  `adopter-supplied`, with no unclassified entries.
- `early-lifecycle-contracts.md` no longer lists a selectable method that is neither shipped
  nor marked adopter-supplied (verified against the skill directory listing).
- A profile referencing a capability that is neither shipped nor marked adopter-supplied is
  rejected by a CI-run check.
- The `A3` placeholder remains marked; the runtime build is cross-referenced, not duplicated.

### Phase 5 — Whole-assembled-composition completeness check

**Objective.** Prove "you can't assemble it wrong": a full selection is validated as one
artifact for topology completeness and dangling references across both planes, backing the
promise `workflow-conformance` already makes.

**Tasks.**

- Define a composition-completeness contract (extend
  `workflow-guide/references/conformance.md` and/or `agent-governance-guide/.../conformance.md`
  with a "whole composition" section) that specifies the checks: profile topology complete
  across both planes; every selected capability present and compatible; no dangling capability
  / provider / module / method reference (resolved against the Phase 4 registry); adoption-mode
  absence report present (per `architecture-and-composition.md` lines 14-24); mandatory
  cross-plane guarantees bound to an adequate enforcement point.
- Implement a helper (e.g. `validateAssembledComposition`) in the appropriate
  `conformance-helpers.mjs` (or a new `composition-helpers.mjs`) that takes a full assembled
  selection and returns the first failure code, composing the existing `validateProfile`,
  `validateComposition`, and provider/enforcement validators rather than duplicating them.
- Add fixtures under `assets/fixtures/` (a passing integrated composition plus adversarial
  cases: a dangling method reference, a selected-but-absent capability, an incompatible
  provider, a missing cross-plane enforcement point) and a `validate-*.ts` (TypeScript per
  Decision 1) that exercises them with mutation guards.
- Wire `workflow-conformance` so its "validate their composition" step delegates to this
  contract; update the command doc if the operation name changes.
- Wire the new validator into the skill `test` task and `moon.yml` inputs.

**Acceptance criteria.**

- A CI-run script validates a shipped reference profile assembled into a full composition and
  fails on each adversarial fixture (dangling reference, absent capability, incompatible
  provider, missing enforcement) with a distinct failure code.
- `workflow-conformance`'s composition-validation promise is backed by a named contract and a
  helper, not prose alone.
- The completeness check consults the Phase 4 registry so an unshipped, unmarked capability in
  a composition is reported as a dangling reference.
- Mutation guards confirm the check cannot be silently defeated (dud-guard count zero).

## Risks and unknowns

- **Prose-table parsing brittleness.** The vocabulary guard must regex-extract enumerations
  from human-readable tables/sentences (e.g. `policy-and-authority.md` line 13). Format drift
  in a doc could produce false failures. Mitigation: reuse the tolerant extractor style from
  the two existing guards and treat "declaration not found" as a failure, so a reworded doc is
  caught rather than silently skipped.
- **Cross-package coupling.** Phase 1 and Phase 2 make governance/workflow skills and
  `command-workflow` docs mutually validating. `moon.yml` `inputs` must list the cross-package
  files (governance's already lists `command-workflow/docs/*.md` and the diagram) or CI caches
  stale. Mitigation: audit `inputs` when adding a site.
- **Fixture-vs-library confusion.** Promoting fixture profiles into a shipped library risks
  re-coupling test data and shipped assets. Mitigation: keep the library in a separate
  directory and keep `assets/fixtures/` as test-only, per `conformance.md` line 30.
- **Adopter-supplied vs ship decision (Phase 4) — DECIDED (Decision 5).** All three —
  `example-mapping` / `user-research` / `architecture-review` — are marked adopter-supplied;
  shipping any later is a registry flip plus a normal skill-create pass, outside this plan.
- **Sibling runtime plan boundary.** `A3` triggers, admission control, and enforcement
  adapters must stay in the runtime plan. Risk of scope bleed when marking placeholders.
  Mitigation: this plan only _marks and validates references_; any build is cross-referenced.
- **Release lockstep.** New scripts/skills touch `package.json`/`moon.yml`, and skill packages
  version in lockstep with commands and `marketplace.json`. Publishing is out of scope, but a
  new skill package (if any) would need registration in `marketplace.json` and a lockfile
  update to keep CI `npm ci` green — a local-only concern here.

## Success criteria

The composition layer is complete for this plan when:

1. Every multi-declared closed composition vocabulary
   (`policyOutcomes`, `moduleKinds`, `moduleClassifications`, `authorityZones`,
   `adoptionModes`, `profileFacets`, `independenceLevels`, work-shape literals) is validated
   against all its declaring sites by a CI-run guard with mutation guards; single-source
   vocabularies are recorded as intentionally unguarded; and `actors.md`'s "validated against
   every declaring site" claim is true.
2. Cross-skill / cross-package reference links are CI-validated; a renamed contract target
   fails a `moon` task.
3. A shipped, adopter-facing reference profile library exists (five worked profiles — one per
   adoption mode, with the three team shapes distributed across them, in per-plane
   `assets/profiles/` homes), each validated against the profile contract, and the
   compose/inspect commands resolve `--profile <reference>` to it.
4. Every selectable capability is classified `shipped` or `adopter-supplied`; no bare
   placeholder methods remain in the early-lifecycle contract; the `A3` placeholder stays
   marked with the runtime build cross-referenced.
5. A whole-assembled-composition completeness check exists, backs `workflow-conformance`'s
   composition promise, and rejects dangling references, absent capabilities, incompatible
   providers, and missing cross-plane enforcement — proving a catalog cannot be assembled
   inconsistent or incomplete.
6. All new validators pass locally via `moon` and are reachable from CI; no publishing or
runtime-enforcement work is performed.
</content>

</invoke>
