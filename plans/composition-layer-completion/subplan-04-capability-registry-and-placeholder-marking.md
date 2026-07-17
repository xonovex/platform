---
type: plan
has_subplans: false
parent_plan: ../composition-layer-completion.md
parallel_group: 1
status: complete
dependencies:
  plans: []
  files:
    - packages/skill/skill-plan/plan-guide/references/early-lifecycle-contracts.md
    - packages/skill/skill-workflow/workflow-guide/references/profiles.md
    - packages/skill/skill-workflow/workflow-guide/references/providers.md
    - packages/skill/skill-agent-governance/agent-governance-guide/references/provider-contracts.md
    - packages/skill/skill-agent-governance/agent-governance-guide/references/execution.md
    - packages/skill/skill-agent-governance/agent-governance-guide/references/modules.md
    - packages/skill/skill-agent-governance/agent-governance-guide/references/autonomy.md
    - packages/skill/skill-agent-governance/agent-governance-guide/scripts/*.mjs
    - packages/skill/skill-agent-governance/agent-governance-guide/assets/*.json
    - packages/skill/skill-agent-governance/package.json
    - packages/skill/skill-agent-governance/moon.yml
skills_to_consult:
  - plan-guide
  - skill-guide
  - microkernel-pattern-guide
  - connascence-guide
  - testing-guide
validation:
  type_check: not_run
  lint: not_run
  build: pass
  tests: pass
updated: "2026-07-18"
---

# Capability Registry and Explicit Placeholder Marking

## Objective

Every capability a profile can select is either shipped or explicitly marked
adopter-supplied, and the composition graph names its placeholders instead of leaving
dangling references. This produces the registry the Phase 3 profile validator and the Phase 5
completeness check consult, and it marks the verified unshipped methods.

## Tasks

1. Enumerate the selectable capability namespace: lifecycle capabilities (from
   `workflow-guide/references/profiles.md` default presentation and
   `plan-guide/references/early-lifecycle-contracts.md`), selectable **methods**
   (`early-lifecycle-contracts.md` line 9), provider ports
   (`workflow-guide/references/providers.md`,
   `agent-governance-guide/references/provider-contracts.md`), executor classes
   (`agent-governance-guide/references/execution.md`), and governance modules
   (`agent-governance-guide/references/modules.md`).
2. Produce a single registry (data file or reference table) mapping each selectable
   capability to `shipped` (naming the owner skill/module) or `adopter-supplied` (naming the
   reference contract it must satisfy), with no unclassified entries.
3. Mark the verified placeholders in `plan-guide/references/early-lifecycle-contracts.md`
   line 9: `example mapping`, `user research`, and `architecture review` are selectable
   methods with no shipped skill (`skill-user-stories`, `skill-bdd`, and
   `skill-accessibility` ship; `skill-example-mapping`, `skill-user-research`, and a general
   architecture-review skill do not exist — `skill-adr` covers decision records, not
   review). Per the settled Decision below, annotate all three inline as adopter-supplied
   capabilities, naming the contract each must satisfy; do not leave bare names implying a
   shipped skill.
4. Credit and preserve the existing correct marking pattern: keep
   `agent-governance-guide/references/autonomy.md` line 9's "`A3` is the eventual goal, not
   a description of what exists"; cross-reference the sibling
   `runtime-enforcement-completion.md` plan for the `A3` trigger/admission-control/escalation
   **build**, and keep only the _marking_ in this plan.
5. Add a registry validator ensuring every selectable capability is classified `shipped` or
   `adopter-supplied` (no unclassified entries) and that a profile referencing an
   unclassified-and-unshipped capability is rejected; wire it into the owning skill's
   `package.json` `test` script and `moon.yml` `test` `inputs`.
6. Expose the registry so downstream consumers import it directly: the Phase 3
   reference-profile validator and the Phase 5 completeness check must be able to consult it
   without duplicating the classification.

## Acceptance criteria

- A capability registry classifies every selectable capability as `shipped` or
  `adopter-supplied`, with no unclassified entries (verified by the registry validator).
- `early-lifecycle-contracts.md` no longer lists a selectable method that is neither shipped
  nor marked adopter-supplied (verified against the `packages/skill/` directory listing —
  `example mapping`, `user research`, `architecture review` are now marked adopter-supplied).
- A profile referencing a capability that is neither shipped nor marked adopter-supplied is
  rejected by a CI-run check.
- The `A3` placeholder remains marked in `autonomy.md`; the runtime build is cross-referenced
  to `runtime-enforcement-completion.md`, not duplicated here.

## Dependencies

None as a prerequisite. Phase 4 enumerates the _existing_ catalog (profiles.md default
presentation, the early-lifecycle contract, provider ports, executor classes, governance
modules) and does not consume the Phase 3 shipped profiles, so it runs concurrently with the
vocabulary guard (Phase 1) and the link guard (Phase 2). Its registry feeds Phase 3 (the
reference-profile validator) and Phase 5 (the completeness check).

## Decision (settled 2026-07-17)

- **(c) Mark, don't ship** (parent Decision 5). All three — `example-mapping`,
  `user-research`, `architecture-review` — are classified `adopter-supplied` in the registry,
  each naming the reference contract to satisfy. No new packages, no marketplace churn in
  this plan; shipping any later is a registry flip plus a normal skill-create pass.

## Implementation notes (completed 2026-07-18)

`build` and `tests` pass via `moon` (`skill-agent-governance:test` — 17 passed, 0 failed,
registry validator included); skill packages define no `type_check`/`lint` task, so those
are `not_run`. The registry classifies **36 capabilities across 5 categories** with 12
mutation guards (dud-guard count zero). Adversarial verification (9/9): flipping a
placeholder to shipped, drifting the executor or module set from `expectedVocabulary`,
removing a registered method the doc still lists, adding a doc method absent from the
registry, an unclassified entry, and a missing shipped-owner package each fail the validator
naming the fault; the registry returns to green on restore; and `resolveCapability` returns
`shipped` / `adopter-supplied` / `dangling` as the Phase 3/5 fail-closed gate requires.

- `agent-governance-guide/assets/capability-registry.json` — the single classification
  record. Categories: `lifecycle-capability` (15, open), `method` (6 named + open tail),
  `provider-port` (4, closed), `executor-class` (5, closed), `governance-module` (6, closed).
  Each entry is `shipped` (names an owner skill) or `adopter-supplied` (names the contract an
  adopter must satisfy).
- `agent-governance-guide/scripts/capability-registry-helpers.mjs` — the single owner of the
  classification logic, imported directly by the Phase 3 reference-profile validator and the
  Phase 5 completeness check: `capabilityRegistry`, `registryEntries`, `capabilityNames`,
  `resolveCapability` (fail-closed → `dangling`), `isResolvableCapability`.
- `agent-governance-guide/scripts/validate-capability-registry.ts` — completeness (every entry
  classified with its required owner/contract field), consistency (`executor-class` and
  `governance-module` mirror `expectedVocabulary`; `method` matches the Method axis parsed
  from `early-lifecycle-contracts.md`), shipped-owner-package existence, placeholder
  correctness, fail-closed resolution, and mutation guards. Wired into the governance
  `package.json` test and `moon.yml` inputs (added `/packages/skill/skill-plan/plan-guide/references/*.md`).
- `early-lifecycle-contracts.md` Method axis now states that user stories, BDD, and
  accessibility review ship as skills while example mapping, user research, and architecture
  review are adopter-supplied, naming the contract and pointing at the registry — no bare name
  implies a shipped skill.
- `autonomy.md`'s `A3` marking is preserved unchanged; the `A3`
  trigger/admission-control/escalation **build** is cross-referenced to the sibling
  [runtime-enforcement-completion.md](../runtime-enforcement-completion.md), not duplicated.

### Design decisions

- **Registry home: `agent-governance-guide`** (the subplan's wiring targets), mirroring the
  fixtures-JSON + helpers-mjs + validate-ts idiom. It reads cross-package (plan-guide
  early-lifecycle contract, plus the executor/module vocabularies it already owns).
- **A profile selects a whole provider *port*, not individual operations**, so the
  `provider-port` category lists the four ports (result-provider, policy, configuration,
  evidence-and-telemetry); the per-port operation sets remain owned by the provider
  conformance validators.
- **Governance modules are selected by capability, constrained to a classification**, so the
  `governance-module` category is the six-value classification vocabulary; `knowledge-only`
  and `advisory` ship with the guide, while `evidence-producing` / `enforcing` /
  `configuration-changing` / `privileged` modules are adopter-native.
- **`example-mapping` is classified adopter-supplied per Decision 5** even though bdd-guide
  discusses the technique — no dedicated `skill-example-mapping` ships; the contract notes the
  bdd-guide coverage.
- **Open categories** (`lifecycle-capability`, `method`'s tail) list the shipped canonical
  set and are not closed-set-checked against prose; the closed categories are consistency-
  checked against their machine owners.
