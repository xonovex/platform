---
type: plan
has_subplans: false
parent_plan: ../composable-workflow-phases.md
parallel_group: 5
status: complete
dependencies:
  plans:
  - governance-onboarding-walking-skeleton
  files:
  - packages/command/command-workflow/commands/discovery-*.md
  - packages/command/command-workflow/commands/research-*.md
  - packages/command/command-workflow/commands/formulation-*.md
  - packages/command/command-workflow/commands/experience-design-*.md
  - packages/command/command-workflow/commands/solution-design-*.md
  - packages/command/command-workflow/commands/decision-*.md
  - packages/skill/skill-plan/**
skills_to_consult:
- command-guide
- skill-guide
- testing-guide
- user-research-guide
- accessibility-guide
- plan-guide
validation:
  type_check: not_applicable
  lint: passed
  build: passed
  tests: passed
  integration: passed
updated: '2026-07-16'
decision_refs:
- D-002
- D-015
- D-019
- D-023
- D-024
- D-025
- D-026
- D-032
- D-033
- D-035
control_refs:
- C-001
- C-008
- C-017
- C-018
- C-026
- C-027
- C-028
- C-029
- C-034
source_refs:
- S-ISO-12207
- S-ISO-15288
- S-ISO-5338
- S-ISO-9241-210
- S-W3C-WCAG22
- S-NIST-AIRMF
- S-LAW-EU-AIACT
- S-W3C-PROV
traceability_files:
- traceability/source-registry.md
- traceability/decision-source-matrix.md
- traceability/control-crosswalk.md
- traceability/subplan-traceability.md
---

# Discovery, Research, Design, Decision, and Planning

## Objective

Implement the early lifecycle capabilities under the new executor-neutral, provider-native, and governance-aware contracts. Preserve methodology neutrality and optional UX/architecture while allowing profiles to attach advisory or enforcing policies through harness/external points without embedding those mechanisms in every lifecycle operation. The lifecycle capabilities must remain usable in a workflow-only composition; governance hooks and external controls strengthen execution when selected but are not implicit dependencies unless a profile explicitly requires them.

## Tasks

1. Replace story/Gherkin-specific entry commands with neutral Discovery and Formulation loops.
2. Implement reusable Research with evidence, uncertainty, provenance, and bounded executor choices.
3. Implement optional Experience Design and Solution Design results with independent critique/revision/acceptance.
4. Implement reusable Decision results separate from evidence and authority.
5. Adapt Planning and subplans to consume opaque native references and publish provider-native results.
6. Define preferred executor classes per operation: deterministic collection, bounded model synthesis, adaptive agent exploration, human/qualified decisions.
7. Add policy hook intents for data access, external research, privacy, accessibility, security, architecture, and regulated applicability without platform mechanics.
8. Support advisory onboarding that recommends methods, skills, providers, and environment modules for the selected profile.
9. Remove hard BDD/user-story/Git/provider dependencies while retaining those skills as selectable methods.
10. Add exact-revision, fresh-context, provider, method, executor, and policy fixtures.

## Traceability

| Task | Implementation intent | Decision IDs | Control IDs | Source IDs |
| --- | --- | --- | --- | --- |
| 1 | Replace story/Gherkin-specific entry commands with neutral Discovery and Formulation loops. | D-002, D-015, D-019, D-023, D-024, D-025, D-026, D-032, D-033, D-035 | C-001, C-008, C-017, C-018, C-026, C-027, C-028, C-029, C-034 | S-ISO-12207, S-ISO-15288, S-ISO-5338, S-ISO-9241-210, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT, S-W3C-PROV |
| 2 | Implement reusable Research with evidence, uncertainty, provenance, and bounded executor choices. | D-002, D-015, D-019, D-023, D-024, D-025, D-026, D-032, D-033, D-035, D-012, D-031 | C-001, C-008, C-017, C-018, C-026, C-027, C-028, C-029, C-034, C-005, C-016 | S-ISO-12207, S-ISO-15288, S-ISO-5338, S-ISO-9241-210, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT, S-W3C-PROV, S-SLSA, S-IN-TOTO, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM |
| 3 | Implement optional Experience Design and Solution Design results with independent critique/revision/acceptance. | D-002, D-015, D-019, D-023, D-024, D-025, D-026, D-032, D-033, D-035 | C-001, C-008, C-017, C-018, C-026, C-027, C-028, C-029, C-034 | S-ISO-12207, S-ISO-15288, S-ISO-5338, S-ISO-9241-210, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT, S-W3C-PROV |
| 4 | Implement reusable Decision results separate from evidence and authority. | D-002, D-015, D-019, D-023, D-024, D-025, D-026, D-032, D-033, D-035 | C-001, C-008, C-017, C-018, C-026, C-027, C-028, C-029, C-034 | S-ISO-12207, S-ISO-15288, S-ISO-5338, S-ISO-9241-210, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT, S-W3C-PROV |
| 5 | Adapt Planning and subplans to consume opaque native references and publish provider-native results. | D-002, D-015, D-019, D-023, D-024, D-025, D-026, D-032, D-033, D-035, D-034 | C-001, C-008, C-017, C-018, C-026, C-027, C-028, C-029, C-034 | S-ISO-12207, S-ISO-15288, S-ISO-5338, S-ISO-9241-210, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT, S-W3C-PROV, S-HEXAGONAL |
| 6 | Define preferred executor classes per operation: deterministic collection, bounded model synthesis, adaptive agent exploration, human/qualified decisions. | D-002, D-015, D-019, D-023, D-024, D-025, D-026, D-032, D-033, D-035, D-031 | C-001, C-008, C-017, C-018, C-026, C-027, C-028, C-029, C-034, C-016, C-020, C-021 | S-ISO-12207, S-ISO-15288, S-ISO-5338, S-ISO-9241-210, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT, S-W3C-PROV, S-NIST-800218A, S-SPDX-AI |
| 7 | Add policy hook intents for data access, external research, privacy, accessibility, security, architecture, and regulated applicability without platform mechanics. | D-002, D-015, D-019, D-023, D-024, D-025, D-026, D-032, D-033, D-035, D-014, D-004, D-005, D-017 | C-001, C-008, C-017, C-018, C-026, C-027, C-028, C-029, C-034, C-009, C-010, C-033, C-006, C-037, C-038 | S-ISO-12207, S-ISO-15288, S-ISO-5338, S-ISO-9241-210, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT, S-W3C-PROV, S-OTEL-GENAI, S-NIST-PRIVACY, S-LAW-EU-GDPR, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2 |
| 8 | Support advisory onboarding that recommends methods, skills, providers, and environment modules for the selected profile. | D-002, D-015, D-019, D-023, D-024, D-025, D-026, D-032, D-033, D-035, D-013, D-028, D-034 | C-001, C-008, C-017, C-018, C-026, C-027, C-028, C-029, C-034, C-015, C-037 | S-ISO-12207, S-ISO-15288, S-ISO-5338, S-ISO-9241-210, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT, S-W3C-PROV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-HEXAGONAL |
| 9 | Remove hard BDD/user-story/Git/provider dependencies while retaining those skills as selectable methods. | D-002, D-015, D-019, D-023, D-024, D-025, D-026, D-032, D-033, D-035, D-034 | C-001, C-008, C-017, C-018, C-026, C-027, C-028, C-029, C-034 | S-ISO-12207, S-ISO-15288, S-ISO-5338, S-ISO-9241-210, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT, S-W3C-PROV, S-HEXAGONAL |
| 10 | Add exact-revision, fresh-context, provider, method, executor, and policy fixtures. | D-002, D-015, D-019, D-023, D-024, D-025, D-026, D-032, D-033, D-035, D-034 | C-001, C-008, C-017, C-018, C-026, C-027, C-028, C-029, C-034 | S-ISO-12207, S-ISO-15288, S-ISO-5338, S-ISO-9241-210, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT, S-W3C-PROV, S-HEXAGONAL |

Mapping IDs resolve through the parent plan’s `traceability/` artifacts. These mappings do not imply equivalence, certification or legal compliance.

## Validation Steps

1. Run neutral, UX, architecture, AI, regulated, and lightweight profile cases.
2. Verify simple tasks use deterministic or bounded model execution and complex exploration may select an agent.
3. Verify human/qualified decisions cannot be fabricated by a model.
4. Verify provider-native handoffs and fresh-context resume.
5. Verify hook/policy requirements are requested semantically and enforced only through selected adapters.

## Success Criteria

- [x] Early lifecycle operations are methodology-, provider-, and executor-neutral.
- [x] UX and Solution Design remain optional and composable.
- [x] Decisions retain evidence, authority, and supersession.
- [x] Planning no longer assumes files, Git, or one agent session.
- [x] Governance requirements compose without contaminating domain command logic.

## Files Modified/Created

- Early lifecycle workflow commands and method skills
- Plan, research, design, decision references and evals

## Dependencies

Depends on the walking skeleton and frozen contracts.

## Estimated Duration

Very large: consolidated early-lifecycle migration and evals.
