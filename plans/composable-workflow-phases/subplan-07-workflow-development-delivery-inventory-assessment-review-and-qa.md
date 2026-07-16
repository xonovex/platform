---
type: plan
has_subplans: false
parent_plan: ../composable-workflow-phases.md
parallel_group: 5
status: pending
dependencies:
  plans:
  - governance-onboarding-walking-skeleton
  files:
  - packages/command/command-workflow/commands/develop-*.md
  - packages/command/command-workflow/commands/deliver-*.md
  - packages/command/command-workflow/commands/inventory-*.md
  - packages/command/command-workflow/commands/assessment-*.md
  - packages/command/command-workflow/commands/review-*.md
  - packages/command/command-workflow/commands/qa-*.md
skills_to_consult:
- command-guide
- skill-guide
- testing-guide
- git-guide
- code-review-guide
- security-assurance-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
updated: '2026-07-14'
decision_refs:
- D-002
- D-003
- D-012
- D-015
- D-020
- D-024
- D-028
- D-031
- D-032
- D-033
- D-034
control_refs:
- C-005
- C-006
- C-007
- C-015
- C-016
- C-017
- C-018
- C-019
- C-021
- C-023
- C-026
- C-027
- C-028
- C-034
- C-037
- C-038
source_refs:
- S-NIST-800218
- S-NIST-800218A
- S-OWASP-ASVS
- S-OWASP-LLM
- S-SLSA
- S-SPDX-AI
- S-CYCLONEDX-MLBOM
- S-W3C-WCAG22
traceability_files:
- traceability/source-registry.md
- traceability/decision-source-matrix.md
- traceability/control-crosswalk.md
- traceability/subplan-traceability.md
---

# Development, Delivery, Inventory, Assessment, Review, and QA

## Objective

Implement development and assurance capabilities using deterministic-first execution, bounded agents where adaptive coding or investigation is required, exact-revision evidence, provider-native publication, harness/external controls, and reusable inventory/assessment semantics. The lifecycle capabilities must remain usable in a workflow-only composition; governance hooks and external controls strengthen execution when selected but are not implicit dependencies unless a profile explicitly requires them.

## Tasks

1. Define Development inputs/results, subplan assignments, parallel groups, workspace isolation, consolidation, and abandonment without treating consolidation as Integration.
2. Map coding work to executor choices: scripts for mechanical changes, bounded model transforms for narrow edits, agents for adaptive multi-step implementation.
3. Integrate semantic session/tool/result/validation intents with harness adapters while keeping runtime mechanics outside commands.
4. Publish Deliverable results through provider skills and bind exact native revisions.
5. Implement deterministic Inventory generation first; allow model enrichment only for non-authoritative descriptions and never guessed versions/components.
6. Support SBOM, AIBOM/AI-SBOM, ML-BOM, CBOM, service, and agent-environment inventory specializations.
7. Implement generic Assessment against any exact result revision and preserve deliverable-specific Review and QA.
8. Compose static tools, CI, scanners, bounded LLM review, independent agents, and human assessors with authoritative evidence origins.
9. Integrate GitHub/GitLab reusable CI modules and external required checks from subplan 04.
10. Define stale-evidence rules when the subject revision, policy, evaluator version, or environment changes.
11. Add concurrency, partial failure, retry, tool bypass, prompt injection, poisoned evidence, and assessor-independence tests.

## Traceability

| Task | Implementation intent | Decision IDs | Control IDs | Source IDs |
| --- | --- | --- | --- | --- |
| 1 | Define Development inputs/results, subplan assignments, parallel groups, workspace isolation, consolidation, and abandonment without treating consolidation as Integration. | D-002, D-003, D-012, D-015, D-020, D-024, D-028, D-031, D-032, D-033, D-034 | C-005, C-006, C-007, C-015, C-016, C-017, C-018, C-019, C-021, C-023, C-026, C-027, C-028, C-034, C-037, C-038 | S-NIST-800218, S-NIST-800218A, S-OWASP-ASVS, S-OWASP-LLM, S-SLSA, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-WCAG22 |
| 2 | Map coding work to executor choices: scripts for mechanical changes, bounded model transforms for narrow edits, agents for adaptive multi-step implementation. | D-002, D-003, D-012, D-015, D-020, D-024, D-028, D-031, D-032, D-033, D-034 | C-005, C-006, C-007, C-015, C-016, C-017, C-018, C-019, C-021, C-023, C-026, C-027, C-028, C-034, C-037, C-038, C-020 | S-NIST-800218, S-NIST-800218A, S-OWASP-ASVS, S-OWASP-LLM, S-SLSA, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT |
| 3 | Integrate semantic session/tool/result/validation intents with harness adapters while keeping runtime mechanics outside commands. | D-002, D-003, D-012, D-015, D-020, D-024, D-028, D-031, D-032, D-033, D-034, D-004, D-005, D-017 | C-005, C-006, C-007, C-015, C-016, C-017, C-018, C-019, C-021, C-023, C-026, C-027, C-028, C-034, C-037, C-038 | S-NIST-800218, S-NIST-800218A, S-OWASP-ASVS, S-OWASP-LLM, S-SLSA, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-WCAG22, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI |
| 4 | Publish Deliverable results through provider skills and bind exact native revisions. | D-002, D-003, D-012, D-015, D-020, D-024, D-028, D-031, D-032, D-033, D-034 | C-005, C-006, C-007, C-015, C-016, C-017, C-018, C-019, C-021, C-023, C-026, C-027, C-028, C-034, C-037, C-038 | S-NIST-800218, S-NIST-800218A, S-OWASP-ASVS, S-OWASP-LLM, S-SLSA, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-WCAG22, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 5 | Implement deterministic Inventory generation first; allow model enrichment only for non-authoritative descriptions and never guessed versions/components. | D-002, D-003, D-012, D-015, D-020, D-024, D-028, D-031, D-032, D-033, D-034 | C-005, C-006, C-007, C-015, C-016, C-017, C-018, C-019, C-021, C-023, C-026, C-027, C-028, C-034, C-037, C-038, C-020 | S-NIST-800218, S-NIST-800218A, S-OWASP-ASVS, S-OWASP-LLM, S-SLSA, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT, S-IN-TOTO, S-SIGSTORE |
| 6 | Support SBOM, AIBOM/AI-SBOM, ML-BOM, CBOM, service, and agent-environment inventory specializations. | D-002, D-003, D-012, D-015, D-020, D-024, D-028, D-031, D-032, D-033, D-034, D-013 | C-005, C-006, C-007, C-015, C-016, C-017, C-018, C-019, C-021, C-023, C-026, C-027, C-028, C-034, C-037, C-038, C-020 | S-NIST-800218, S-NIST-800218A, S-OWASP-ASVS, S-OWASP-LLM, S-SLSA, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-WCAG22, S-NIST-AIRMF, S-LAW-EU-AIACT, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-IN-TOTO, S-SIGSTORE |
| 7 | Implement generic Assessment against any exact result revision and preserve deliverable-specific Review and QA. | D-002, D-003, D-012, D-015, D-020, D-024, D-028, D-031, D-032, D-033, D-034 | C-005, C-006, C-007, C-015, C-016, C-017, C-018, C-019, C-021, C-023, C-026, C-027, C-028, C-034, C-037, C-038 | S-NIST-800218, S-NIST-800218A, S-OWASP-ASVS, S-OWASP-LLM, S-SLSA, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-WCAG22 |
| 8 | Compose static tools, CI, scanners, bounded LLM review, independent agents, and human assessors with authoritative evidence origins. | D-002, D-003, D-012, D-015, D-020, D-024, D-028, D-031, D-032, D-033, D-034 | C-005, C-006, C-007, C-015, C-016, C-017, C-018, C-019, C-021, C-023, C-026, C-027, C-028, C-034, C-037, C-038 | S-NIST-800218, S-NIST-800218A, S-OWASP-ASVS, S-OWASP-LLM, S-SLSA, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-WCAG22 |
| 9 | Integrate GitHub/GitLab reusable CI modules and external required checks from subplan 04. | D-002, D-003, D-012, D-015, D-020, D-024, D-028, D-031, D-032, D-033, D-034, D-013 | C-005, C-006, C-007, C-015, C-016, C-017, C-018, C-019, C-021, C-023, C-026, C-027, C-028, C-034, C-037, C-038 | S-NIST-800218, S-NIST-800218A, S-OWASP-ASVS, S-OWASP-LLM, S-SLSA, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-WCAG22, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY |
| 10 | Define stale-evidence rules when the subject revision, policy, evaluator version, or environment changes. | D-002, D-003, D-012, D-015, D-020, D-024, D-028, D-031, D-032, D-033, D-034, D-013 | C-005, C-006, C-007, C-015, C-016, C-017, C-018, C-019, C-021, C-023, C-026, C-027, C-028, C-034, C-037, C-038 | S-NIST-800218, S-NIST-800218A, S-OWASP-ASVS, S-OWASP-LLM, S-SLSA, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-WCAG22, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY |
| 11 | Add concurrency, partial failure, retry, tool bypass, prompt injection, poisoned evidence, and assessor-independence tests. | D-002, D-003, D-012, D-015, D-020, D-024, D-028, D-031, D-032, D-033, D-034 | C-005, C-006, C-007, C-015, C-016, C-017, C-018, C-019, C-021, C-023, C-026, C-027, C-028, C-034, C-037, C-038 | S-NIST-800218, S-NIST-800218A, S-OWASP-ASVS, S-OWASP-LLM, S-SLSA, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-WCAG22 |

Mapping IDs resolve through the parent plan’s `traceability/` artifacts. These mappings do not imply equivalence, certification or legal compliance.

## Validation Steps

1. Run simple deterministic edit, bounded model edit, and adaptive agent development fixtures.
2. Run parallel workspaces and deterministic consolidation.
3. Publish one local/non-file and one hosted Deliverable.
4. Generate/validate inventory without LLM-fabricated facts.
5. Run concurrent Review, QA, security, accessibility, AI, and supply-chain assessments.
6. Change the subject revision and verify stale evidence/acceptance behavior.

## Success Criteria

- [ ] Development uses the least autonomous suitable executor.
- [ ] Harness and CI governance are composable but independent from lifecycle semantics.
- [ ] Deliverables and evidence are exact-revision bound.
- [ ] Inventory is authoritative and interoperable where formats are selected.
- [ ] Assessment is generic; Review and QA remain focused specializations.
- [ ] Mandatory controls have independent enforcement evidence.

## Files Modified/Created

- Development, delivery, inventory, assessment, review, and QA command/skill families
- CI/evidence integration fixtures and evals

## Dependencies

Depends on the walking skeleton and frozen contracts.

## Estimated Duration

Very large: coding, delivery, inventory, assurance, and integration surfaces.
