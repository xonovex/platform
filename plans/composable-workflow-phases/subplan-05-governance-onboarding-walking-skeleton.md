---
type: plan
has_subplans: false
parent_plan: ../composable-workflow-phases.md
parallel_group: 4
status: complete
dependencies:
  plans:
  - agent-harness-adapters-and-onboarding
  - external-automation-ci-and-platform-enforcement
  - enterprise-platform-skills-and-onboarding
  files:
  - packages/skill/skill-agent-governance/evals/**
  - packages/agent/agent-cli-go/**
  - packages/agent/agent-operator-go/**
  - packages/skill/skill-github/**
skills_to_consult:
- testing-guide
- skill-guide
- security-assurance-guide
- git-guide
validation:
  type_check: passed
  lint: passed
  build: passed
  tests: passed
  integration: passed
updated: '2026-07-16'
decision_refs:
- D-001
- D-003
- D-004
- D-005
- D-009
- D-015
- D-016
- D-017
- D-018
- D-028
- D-033
- D-034
- D-035
- D-036
- D-037
- D-038
- D-039
control_refs:
- C-004
- C-006
- C-007
- C-011
- C-015
- C-028
- C-036
- C-037
- C-038
- C-041
- C-042
- C-043
- C-044
- C-045
source_refs:
- S-HARNESS-CLAUDE
- S-HARNESS-CODEX
- S-GITHUB-REUSABLE
- S-GITLAB-COMPONENTS
- S-NIST-80053
- S-W3C-PROV
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

# Governance and Onboarding Walking Skeleton

## Objective

Prove the two-plane architecture before broad lifecycle implementation. Demonstrate environment discovery, modular recommendation, preview and authorization, native harness configuration, deterministic enforcement, one bounded semantic evaluator or specialist agent, provider-native evidence, external CI enforcement, fresh-context recovery, drift detection, and rollback.

## Tasks

1. Choose one harness with mature hook support, one hosted CI/provider, the enterprise-platform composition fixture, and the self-controlled non-file provider fixture.
2. Exercise three adoption paths: governance-only around an ordinary agent task, workflow-only with provider-native results, and an integrated workflow-plus-governance composition.
3. Discover harness version, supported events/handler types, configuration layers, installed modules, CI controls, provider capabilities, and selected profiles.
4. Recommend a minimal composition: session context injection, protected-path/tool policy, post-change validation, audit evidence, and CI-required check.
5. Present an exact preview including native changes, requested permissions, data flow, network/model use, failure mode, and rollback.
6. Apply configuration only after authorization and verify it through native diagnostics.
7. Trigger a permitted tool operation and a denied operation; prove deterministic policy and explanation.
8. Trigger a bounded model evaluator or specialist agent for advisory review with schema, time/cost/tool limits, and a provider-native result reference.
9. Publish related evidence without YAML sidecars and recover it in a fresh process/session from opaque references.
10. Trigger the corresponding CI/external check, including one Azure DevOps/Bitbucket/Bitrise/AWS/Datadog path where selected, and demonstrate that mandatory policy remains effective if the harness hook is disabled.
11. Detect deliberate configuration drift, recommend remediation, apply or waive through an explicit result, then roll back all installed modules.
12. Run negative cases: untrusted project module, unsupported hook type, policy service unavailable, concurrent duplicate event, recursive agent launch, and expired exception.

## Traceability

| Task | Implementation intent | Decision IDs | Control IDs | Source IDs |
| --- | --- | --- | --- | --- |
| 1 | Choose one harness with mature hook support, one hosted CI/provider, the enterprise-platform composition fixture, and the self-controlled non-file provider fixture. | D-001, D-003, D-004, D-005, D-009, D-015, D-016, D-017, D-018, D-028, D-033, D-034, D-035 | C-004, C-006, C-007, C-011, C-015, C-028, C-036, C-037, C-038, C-034 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-GITHUB-REUSABLE, S-GITLAB-COMPONENTS, S-NIST-80053, S-W3C-PROV, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-ISO-12207, S-ISO-15288, S-HEXAGONAL |
| 2 | Exercise three adoption paths: governance-only around an ordinary agent task, workflow-only with provider-native results, and an integrated workflow-plus-governance composition. | D-001, D-003, D-004, D-005, D-009, D-015, D-016, D-017, D-018, D-028, D-033, D-034, D-035 | C-004, C-006, C-007, C-011, C-015, C-028, C-036, C-037, C-038, C-034 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-GITHUB-REUSABLE, S-GITLAB-COMPONENTS, S-NIST-80053, S-W3C-PROV, S-ISO-12207, S-ISO-15288, S-HEXAGONAL |
| 3 | Discover harness version, supported events/handler types, configuration layers, installed modules, CI controls, provider capabilities, and selected profiles. | D-001, D-003, D-004, D-005, D-009, D-015, D-016, D-017, D-018, D-028, D-033, D-034, D-035 | C-004, C-006, C-007, C-011, C-015, C-028, C-036, C-037, C-038, C-034 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-GITHUB-REUSABLE, S-GITLAB-COMPONENTS, S-NIST-80053, S-W3C-PROV, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-ISO-12207, S-ISO-15288, S-HEXAGONAL |
| 4 | Recommend a minimal composition: session context injection, protected-path/tool policy, post-change validation, audit evidence, and CI-required check. | D-001, D-003, D-004, D-005, D-009, D-015, D-016, D-017, D-018, D-028, D-033, D-034, D-035 | C-004, C-006, C-007, C-011, C-015, C-028, C-036, C-037, C-038 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-GITHUB-REUSABLE, S-GITLAB-COMPONENTS, S-NIST-80053, S-W3C-PROV |
| 5 | Present an exact preview including native changes, requested permissions, data flow, network/model use, failure mode, and rollback. | D-001, D-003, D-004, D-005, D-009, D-015, D-016, D-017, D-018, D-028, D-033, D-034, D-035, D-031, D-032, D-022, D-030 | C-004, C-006, C-007, C-011, C-015, C-028, C-036, C-037, C-038, C-016, C-017, C-018, C-020, C-021, C-014, C-035, C-040 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-GITHUB-REUSABLE, S-GITLAB-COMPONENTS, S-NIST-80053, S-W3C-PROV, S-NIST-AIRMF, S-NIST-800218A, S-LAW-EU-AIACT, S-SPDX-AI, S-ISO-12207, S-ISO-15288 |
| 6 | Apply configuration only after authorization and verify it through native diagnostics. | D-001, D-003, D-004, D-005, D-009, D-015, D-016, D-017, D-018, D-028, D-033, D-034, D-035 | C-004, C-006, C-007, C-011, C-015, C-028, C-036, C-037, C-038 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-GITHUB-REUSABLE, S-GITLAB-COMPONENTS, S-NIST-80053, S-W3C-PROV |
| 7 | Trigger a permitted tool operation and a denied operation; prove deterministic policy and explanation. | D-001, D-003, D-004, D-005, D-009, D-015, D-016, D-017, D-018, D-028, D-033, D-034, D-035 | C-004, C-006, C-007, C-011, C-015, C-028, C-036, C-037, C-038 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-GITHUB-REUSABLE, S-GITLAB-COMPONENTS, S-NIST-80053, S-W3C-PROV |
| 8 | Trigger a bounded model evaluator or specialist agent for advisory review with schema, time/cost/tool limits, and a provider-native result reference. | D-001, D-003, D-004, D-005, D-009, D-015, D-016, D-017, D-018, D-028, D-033, D-034, D-035, D-031, D-032 | C-004, C-006, C-007, C-011, C-015, C-028, C-036, C-037, C-038, C-016, C-017, C-018, C-020, C-021, C-034 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-GITHUB-REUSABLE, S-GITLAB-COMPONENTS, S-NIST-80053, S-W3C-PROV, S-NIST-AIRMF, S-NIST-800218A, S-LAW-EU-AIACT, S-SPDX-AI, S-ISO-12207, S-ISO-15288, S-HEXAGONAL |
| 9 | Publish related evidence without YAML sidecars and recover it in a fresh process/session from opaque references. | D-001, D-003, D-004, D-005, D-009, D-015, D-016, D-017, D-018, D-028, D-033, D-034, D-035 | C-004, C-006, C-007, C-011, C-015, C-028, C-036, C-037, C-038, C-034 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-GITHUB-REUSABLE, S-GITLAB-COMPONENTS, S-NIST-80053, S-W3C-PROV, S-ISO-12207, S-ISO-15288, S-HEXAGONAL |
| 10 | Trigger the corresponding CI/external check, including one Azure DevOps/Bitbucket/Bitrise/AWS/Datadog path where selected, and demonstrate that mandatory policy remains effective if the harness hook is disabled. | D-001, D-003, D-004, D-005, D-009, D-015, D-016, D-017, D-018, D-028, D-033, D-034, D-035 | C-004, C-006, C-007, C-011, C-015, C-028, C-036, C-037, C-038 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-GITHUB-REUSABLE, S-GITLAB-COMPONENTS, S-NIST-80053, S-W3C-PROV, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI |
| 11 | Detect deliberate configuration drift, recommend remediation, apply or waive through an explicit result, then roll back all installed modules. | D-001, D-003, D-004, D-005, D-009, D-015, D-016, D-017, D-018, D-028, D-033, D-034, D-035 | C-004, C-006, C-007, C-011, C-015, C-028, C-036, C-037, C-038 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-GITHUB-REUSABLE, S-GITLAB-COMPONENTS, S-NIST-80053, S-W3C-PROV |
| 12 | Run negative cases: untrusted project module, unsupported hook type, policy service unavailable, concurrent duplicate event, recursive agent launch, and expired exception. | D-001, D-003, D-004, D-005, D-009, D-015, D-016, D-017, D-018, D-028, D-033, D-034, D-035, D-029 | C-004, C-006, C-007, C-011, C-015, C-028, C-036, C-037, C-038, C-012, C-013 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-GITHUB-REUSABLE, S-GITLAB-COMPONENTS, S-NIST-80053, S-W3C-PROV, S-NIST-80061, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI |

Mapping IDs resolve through the parent plan’s `traceability/` artifacts. These mappings do not imply equivalence, certification or legal compliance.

## Validation Steps

1. Automated end-to-end fixture with recorded native references.
2. Fresh-context replay without original conversation or local sidecar state.
3. Concurrency/retry/idempotency/failure tests.
4. Permission, authority attenuation, cost, and data-redaction assertions.
5. Full rollback and post-rollback drift verification.

## Evidence

Two evidence classes back this subplan; neither is presented as the other.

- **Executed locally (2026-07-16):** `agent-governance-guide/assets/walking-skeleton/run-skeleton.sh --yes` — 17/17 checks, wired into `skill-agent-governance:test` so CI re-executes it. Demonstrates local environment discovery (observed harness `2.1.211 (Claude Code)`, configuration layers), preview-then-consent (the script stops without `--yes`), idempotent apply, allow/deny with explained denials, an independent locally re-invoked CI-shaped gate with the hook disabled, drift detection and remediation, checksum module trust, evidence dedup, recursion limit, exception expiry, and rollback.
- **Fixture-recorded only (not executed):** hosted CI/platform enforcement (GitHub required check, rulesets), the live-tenant provider round-trip behind opaque references, and the bounded LLM evaluator leg. These are validated compositions in `validate-walking-skeleton-fixtures.mjs`, not runtime conformance; live probes remain candidate runs requiring credentials.

## Success Criteria

- [x] Workflow-only, governance-only, and integrated adoption paths all work without hidden coupling (fixture-validated adoption paths).
- [x] Onboarding discovers and explains the actual environment (executed: local harness version and configuration layers).
- [x] User sees and authorizes all native changes and permissions (executed: preview stops without explicit consent).
- [x] Deterministic enforcement and bounded semantic advice coexist (deterministic leg executed; bounded-evaluator leg fixture-recorded).
- [x] Mandatory control has an independent external enforcement layer (executed locally as a re-invoked CI-shaped gate; hosted enforcement fixture-recorded).
- [x] Evidence is provider-native and recoverable by opaque reference (fixture-recorded; local run demonstrates deduplicated evidence records only).
- [x] Drift, exception, disablement, and rollback work (executed).
- [x] Unsupported or unsafe compositions fail clearly (executed: tampered module, unsupported intent, outage fail-closed; remainder fixture-recorded).

## Files Modified/Created

- Walking-skeleton fixtures/evals
- Selected harness adapter and GitHub/CI fixtures
- Agent CLI/operator integration fixture where applicable

## Dependencies

Depends on harness and external automation plans.

## Estimated Duration

Large: vertical integration and adversarial validation.
