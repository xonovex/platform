---
type: plan
has_subplans: false
parent_plan: ../composition-layer-completion.md
parallel_group: 1
status: pending
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
  build: not_run
  tests: not_run
updated: "2026-07-17"
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
   review). Per Open decisions, either annotate them inline as adopter-supplied capabilities
   (naming the contract to satisfy) or ship reference skills; do not leave bare names
   implying a shipped skill.
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
  `example mapping`, `user research`, `architecture review` are now marked or shipped).
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

## Open decisions

The parent leaves one decision open for this phase (Risks and unknowns, "Adopter-supplied vs
ship decision (Phase 4)"):

- **(c) Ship vs mark the three missing method skills.** Whether to ship
  `example-mapping`, `user-research`, and `architecture-review` reference skills or mark them
  adopter-supplied capabilities. Shipping expands the catalog beyond this plan's stated
  intent; marking keeps the graph honest without new skills. Decide per capability; do not
  resolve it here.
