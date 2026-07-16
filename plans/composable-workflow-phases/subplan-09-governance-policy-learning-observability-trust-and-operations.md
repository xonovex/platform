---
type: plan
has_subplans: false
parent_plan: ../composable-workflow-phases.md
parallel_group: 6
status: pending
dependencies:
  plans:
  - agent-harness-adapters-and-onboarding
  - external-automation-ci-and-platform-enforcement
  - workflow-discovery-research-design-decision-and-planning
  - workflow-development-delivery-inventory-assessment-review-and-qa
  - workflow-acceptance-integration-transition-release-observation-incidents-and-retirement
  - enterprise-platform-skills-and-onboarding
  files:
  - packages/skill/skill-agent-governance/**
  - packages/skill/skill-accessibility/**
  - packages/skill/skill-ai-governance/**
  - packages/skill/skill-security-assurance/**
  - packages/skill/skill-reliability/**
  - packages/skill/skill-reflect/**
skills_to_consult:
- skill-guide
- testing-guide
- reflect-guide
- security-assurance-guide
- reliability-guide
- versioning-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
updated: '2026-07-14'
decision_refs:
- D-003
- D-006
- D-007
- D-008
- D-011
- D-012
- D-014
- D-018
- D-023
- D-024
- D-025
- D-029
- D-030
- D-031
- D-032
- D-036
- D-037
- D-038
- D-039
control_refs:
- C-001
- C-002
- C-003
- C-004
- C-005
- C-009
- C-010
- C-011
- C-012
- C-013
- C-014
- C-016
- C-017
- C-018
- C-019
- C-020
- C-021
- C-022
- C-023
- C-024
- C-025
- C-026
- C-027
- C-032
- C-033
- C-038
- C-039
- C-040
- C-041
- C-042
- C-043
- C-044
- C-045
source_refs:
- S-NIST-80053
- S-NIST-80061
- S-NIST-AIRMF
- S-NIST-PRIVACY
- S-NIST-CSF
- S-LAW-EU-AIACT
- S-LAW-EU-DORA
- S-LAW-EU-GDPR
- S-OTEL-GENAI
- S-DORA-CAPS
- S-ISO-42001
- S-AZDO-TEMPLATES
- S-AZDO-APPROVALS
- S-AZDO-BRANCH
- S-AZDO-WORKITEMS
- S-AZDO-SERVICE-CONNECTIONS
- S-AZDO-REST
- S-AZDO-SERVICE-HOOKS
- S-BITBUCKET-CLOUD-PIPELINES
- S-BITBUCKET-CLOUD-OIDC
- S-BITBUCKET-CLOUD-DEPLOY
- S-BITBUCKET-CLOUD-MERGE
- S-BITBUCKET-CLOUD-BRANCH
- S-BITBUCKET-CLOUD-SHARED
- S-BITBUCKET-CLOUD-CUSTOM-CHECKS
- S-BITBUCKET-CLOUD-REST
- S-BITBUCKET-CLOUD-WEBHOOKS
- S-BITBUCKET-DC
- S-BITBUCKET-DC-REST
- S-BITRISE-WORKFLOWS
- S-BITRISE-SECRETS
- S-BITRISE-ARTIFACTS
- S-BITRISE-TRIGGERS
- S-BITRISE-OIDC-AWS
- S-BITRISE-BUILD-STATUS
- S-BITRISE-VERIFIED-STEPS
- S-AWS-IAM-BEST
- S-AWS-OIDC
- S-AWS-SCP
- S-AWS-CLOUDTRAIL
- S-AWS-CONFIG
- S-AWS-SECURITY-HUB
- S-DATADOG-CI
- S-DATADOG-CD
- S-DATADOG-OTEL
- S-DATADOG-AUDIT
- S-DATADOG-CATALOG
- S-DATADOG-AWS
- S-DATADOG-CLOUD-SECURITY
- S-DATADOG-LLM
- S-DATADOG-DORA
traceability_files:
- traceability/source-registry.md
- traceability/decision-source-matrix.md
- traceability/control-crosswalk.md
- traceability/subplan-traceability.md
---

# Governance, Policy, Learning, Observability, Trust, and Operations

## Objective

Apply the governance plane across lifecycle and non-lifecycle activity: policy selection and decisions, applicability/crosswalks, data protection, observability, module catalogs, trust and provenance, exceptions, drift, updates, learning, operational support, and continuous assurance.

## Tasks

1. Implement policy bundles and hook/external enforcement mappings for security, privacy, accessibility, reliability, AI governance, supply chain, data, cost, and regulated profiles.
2. Support simple deterministic policies, provider-native policies, and optional OPA/Rego without making one engine mandatory.
3. Define fail-closed/fail-visible behavior per policy and enforcement point; telemetry or advisory failures do not automatically block.
4. Implement exception, waiver, break-glass, compensating-control, expiry, and review workflows.
5. Implement data classification, prompt/context/tool redaction, retention, residency, consent, model routing, and external-transfer rules.
6. Instrument model, agent, MCP, tool, policy, CI/CD, provider, and privileged operations using OpenTelemetry-compatible semantics where feasible.
7. Prevent sensitive-content logging by default; support hashes, metadata, sampling, and role-based access.
8. Build a module/profile catalog with compatibility, owner, trust, permissions, conflicts, presets, lifecycle status, usage, and deprecation.
9. Add adoption-mode and authority-zone views so users can distinguish knowledge-only, advisory, evidence-producing, enforcing, configuration-changing, and privileged modules and understand which layer owns each control.
10. Extend Inventory/AIBOM to include the effective agent environment and executable governance modules when selected.
11. Implement drift detection for harness configuration, CI policies, module versions, provider capabilities, profile/control versions, and managed settings.
12. Implement safe update/canary/rollback/emergency-disable processes for governance modules.
13. Wire learning extraction from lifecycle, onboarding, policy denials, incidents, exceptions, drift, and module failures; promotion remains reviewed and reversible.
14. Add governance effectiveness metrics that avoid surveillance and gaming: control coverage, false positives, bypass attempts, rollback success, onboarding time, drift, DORA outcomes, and user friction.
15. Define support and ownership expectations, incident response for governance modules, and retirement of obsolete hooks/policies/plugins.
16. Run cross-platform tests for conflicts, concurrency, trust, data leakage, stale versions, missing enforcement, exception abuse, and telemetry outages.

## Traceability

| Task | Implementation intent | Decision IDs | Control IDs | Source IDs |
| --- | --- | --- | --- | --- |
| 1 | Implement policy bundles and hook/external enforcement mappings for security, privacy, accessibility, reliability, AI governance, supply chain, data, cost, and regulated profiles. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032, D-004, D-005, D-017, D-026 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040, C-006, C-037 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001, S-NIST-800218A, S-SPDX-AI, S-W3C-WCAG22, S-ISO-9241-210, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-SLSA, S-IN-TOTO, S-SIGSTORE, S-CYCLONEDX-MLBOM, S-LAW-EU-CRA, S-LAW-EU-NIS2 |
| 2 | Support simple deterministic policies, provider-native policies, and optional OPA/Rego without making one engine mandatory. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032, D-015, D-033, D-034 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040, C-028, C-034 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 3 | Define fail-closed/fail-visible behavior per policy and enforcement point; telemetry or advisory failures do not automatically block. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001 |
| 4 | Implement exception, waiver, break-glass, compensating-control, expiry, and review workflows. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032, D-004, D-005, D-017 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040, C-006, C-037 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI |
| 5 | Implement data classification, prompt/context/tool redaction, retention, residency, consent, model routing, and external-transfer rules. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001, S-NIST-800218A, S-SPDX-AI |
| 6 | Instrument model, agent, MCP, tool, policy, CI/CD, provider, and privileged operations using OpenTelemetry-compatible semantics where feasible. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032, D-013, D-028, D-015, D-033, D-034 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040, C-015, C-037, C-028, C-034 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001, S-NIST-800218A, S-SPDX-AI, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 7 | Prevent sensitive-content logging by default; support hashes, metadata, sampling, and role-based access. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001 |
| 8 | Build a module/profile catalog with compatibility, owner, trust, permissions, conflicts, presets, lifecycle status, usage, and deprecation. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001 |
| 9 | Add adoption-mode and authority-zone views so users can distinguish knowledge-only, advisory, evidence-producing, enforcing, configuration-changing, and privileged modules and understand which layer owns each control. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001 |
| 10 | Extend Inventory/AIBOM to include the effective agent environment and executable governance modules when selected. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032, D-013, D-028 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040, C-015, C-037 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001, S-NIST-800218A, S-SPDX-AI, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-SLSA, S-IN-TOTO, S-SIGSTORE, S-CYCLONEDX-MLBOM |
| 11 | Implement drift detection for harness configuration, CI policies, module versions, provider capabilities, profile/control versions, and managed settings. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032, D-004, D-005, D-017, D-015, D-033, D-034 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040, C-006, C-037, C-028, C-034 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 12 | Implement safe update/canary/rollback/emergency-disable processes for governance modules. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032, D-022 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040, C-035 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001, S-ISO-12207, S-ISO-15288 |
| 13 | Wire learning extraction from lifecycle, onboarding, policy denials, incidents, exceptions, drift, and module failures; promotion remains reviewed and reversible. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001 |
| 14 | Add governance effectiveness metrics that avoid surveillance and gaming: control coverage, false positives, bypass attempts, rollback success, onboarding time, drift, DORA outcomes, and user friction. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032, D-022 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040, C-035 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001, S-ISO-12207, S-ISO-15288 |
| 15 | Define support and ownership expectations, incident response for governance modules, and retirement of obsolete hooks/policies/plugins. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032, D-004, D-005, D-017, D-022 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040, C-006, C-037, C-035 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-ISO-12207, S-ISO-15288 |
| 16 | Run cross-platform tests for conflicts, concurrency, trust, data leakage, stale versions, missing enforcement, exception abuse, and telemetry outages. | D-003, D-006, D-007, D-008, D-011, D-012, D-014, D-018, D-023, D-024, D-025, D-029, D-030, D-031, D-032 | C-001, C-002, C-003, C-004, C-005, C-009, C-010, C-011, C-012, C-013, C-014, C-016, C-017, C-018, C-019, C-020, C-021, C-022, C-023, C-024, C-025, C-026, C-027, C-032, C-033, C-038, C-039, C-040 | S-NIST-80053, S-NIST-80061, S-NIST-AIRMF, S-NIST-PRIVACY, S-NIST-CSF, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-GDPR, S-OTEL-GENAI, S-DORA-CAPS, S-ISO-42001 |

Mapping IDs resolve through the parent plan’s `traceability/` artifacts. These mappings do not imply equivalence, certification or legal compliance.

## Validation Steps

1. Run representative lightweight, secure, AI, regulated, and organization-managed profiles.
2. Verify mandatory policies fail when no adequate enforcement point exists.
3. Verify exceptions and break-glass cannot become permanent defaults.
4. Verify telemetry redaction/retention and OTel-compatible correlation.
5. Verify module trust, updates, drift, rollback, and retirement.
6. Run learning promotion and stale-policy removal cases.

## Success Criteria

- [ ] Policies operate across lifecycle and general agent activity.
- [ ] Enforcement is layered, explainable, and profile-selected.
- [ ] Data and telemetry governance is explicit and privacy-preserving.
- [ ] Catalog composition is searchable, compatible, and conflict-aware.
- [ ] Modules are inventoried, versioned, owned, updateable, rollbackable, and retireable.
- [ ] Drift, exceptions, break-glass, and governance incidents are first-class.
- [ ] Learning improves modules and policies through governed adoption.

## Files Modified/Created

- Governance/policy/observability/catalog/trust references and packages
- Accessibility, AI, security, reliability, and reflect integrations
- Cross-platform policy and telemetry fixtures

## Dependencies

Depends on harness/CI adapters and all lifecycle groups.

## Estimated Duration

Very large: cross-cutting governance and operationalization.
