---
type: plan
has_subplans: false
parent_plan: ../composable-workflow-phases.md
parallel_group: 3
status: pending
dependencies:
  plans:
  - workflow-governance-contracts-profiles-and-execution
  - provider-policy-and-module-conformance
  files:
  - packages/skill/skill-azure-devops/**
  - packages/skill/skill-bitbucket/**
  - packages/skill/skill-bitrise/**
  - packages/skill/skill-aws/**
  - packages/skill/skill-datadog/**
  - .claude-plugin/marketplace.json
  - .agents/plugins/marketplace.json
  - package-lock.json
skills_to_consult:
- skill-guide
- testing-guide
- security-assurance-guide
- reliability-guide
- orthogonal-pattern-guide
- versioning-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
updated: '2026-07-14'
decision_refs:
- D-001
- D-003
- D-005
- D-009
- D-012
- D-013
- D-014
- D-015
- D-023
- D-025
- D-026
- D-027
- D-030
- D-033
- D-036
- D-037
- D-038
- D-039
control_refs:
- C-004
- C-007
- C-009
- C-011
- C-015
- C-027
- C-028
- C-030
- C-033
- C-034
- C-036
- C-037
- C-038
- C-040
- C-041
- C-042
- C-043
- C-044
- C-045
source_refs:
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
- traceability/platform-capability-matrix.md
- traceability/subplan-traceability.md
---

# Enterprise Platform Skills and Onboarding

## Objective

Add independently composable, source-backed skills and onboarding modules for the company stack: Azure DevOps, Bitbucket Cloud/Data Center, Bitrise, AWS and Datadog. Preserve provider-native references and evidence, distinguish cloud and self-managed editions, prefer temporary federated credentials, and make every setup previewable, verifiable, reversible and optional.

## Tasks

1. Audit existing Azure DevOps, Bitbucket, Bitrise, AWS and Datadog packages, references, scripts, provider operations and marketplace entries; document reuse, gaps and ownership boundaries.
2. Define a shared enterprise-platform skill convention covering product/edition detection, capability negotiation, native references, permissions, credential/data-flow disclosure, conformance, rollback, drift and progressive disclosure.
3. Create the Azure DevOps skill owner with separate references for Boards work items, Repos/pull requests and branch policies, Pipelines templates/approvals/artifacts, service connections, REST/service hooks and safe onboarding.
4. Implement Azure DevOps conformance fixtures for Services and supported Server baselines, including work-item relationships, PR/build evidence, reusable pipeline templates, approvals/checks and provider-native references.
5. Create the Bitbucket skill owner with an explicit Cloud versus Data Center capability matrix and no inferred parity.
6. Add Bitbucket Cloud references and fixtures for repositories/PRs, Pipelines, shared configurations, deployments, merge/custom checks, branch permissions, OIDC, REST and webhooks.
7. Add Bitbucket Data Center references and fixtures for version detection, REST, repositories/PRs, permissions, build/deployment status, webhooks/plugins and upgrade limitations.
8. Create the Bitrise skill owner for Workflows, Pipelines, Steps and versioning, triggers, secrets, artifacts, build statuses, API/webhooks, hosted/self-hosted execution and safe onboarding.
9. Add Bitrise-to-AWS OIDC onboarding with trust-policy preview, constrained claims, least-privilege roles, verification, rollback and prohibition of generated long-lived access keys.
10. Create the AWS skill owner for IAM/federation, OIDC, permissions boundaries and least privilege, Organizations/SCPs, CloudTrail, Config conformance packs, Security Hub, evidence publication and break-glass-safe onboarding.
11. Create the Datadog skill owner for CI/CD visibility, DORA metrics, OpenTelemetry, LLM observability, Audit Trail, Software Catalog, AWS integration, Cloud Security, evidence references and privacy-preserving telemetry onboarding.
12. Publish a mixed-enterprise-stack composition reference using Azure Boards for work items, Bitbucket for source and pull requests, Bitrise for mobile CI, AWS for runtime/governance and Datadog for observability, while preserving independent provider-native results.
13. Create idempotent setup, diagnostic, dry-run, verification, rollback and drift helpers that use native APIs/configuration and display exact permission, credential, network and data changes before mutation.
14. Add positive, negative and near-miss trigger evals plus live/pinned conformance fixtures for editions, tiers, unsupported capabilities, OIDC claims, secret exposure, artifact/status linkage, telemetry redaction and provider outages.
15. Register the five optional skill packages alphabetically, pin official sources and tested versions, update package/marketplace metadata, and publish migration, coexistence and uninstall guidance.

## Traceability

| Task | Implementation intent | Decision IDs | Control IDs | Source IDs |
| --- | --- | --- | --- | --- |
| 1 | Audit existing Azure DevOps, Bitbucket, Bitrise, AWS and Datadog packages, references, scripts, provider operations and marketplace entries; document reuse, gaps and ownership boundaries. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-AZDO-TEMPLATES, S-AZDO-APPROVALS, S-AZDO-BRANCH, S-AZDO-WORKITEMS, S-AZDO-SERVICE-CONNECTIONS, S-AZDO-REST, S-AZDO-SERVICE-HOOKS, S-BITBUCKET-CLOUD-PIPELINES, S-BITBUCKET-CLOUD-OIDC, S-BITBUCKET-CLOUD-DEPLOY, S-BITBUCKET-CLOUD-MERGE, S-BITBUCKET-CLOUD-BRANCH, S-BITBUCKET-CLOUD-SHARED, S-BITBUCKET-CLOUD-CUSTOM-CHECKS, S-BITBUCKET-CLOUD-REST, S-BITBUCKET-CLOUD-WEBHOOKS, S-BITBUCKET-DC, S-BITBUCKET-DC-REST, S-BITRISE-WORKFLOWS, S-BITRISE-SECRETS, S-BITRISE-ARTIFACTS, S-BITRISE-TRIGGERS, S-BITRISE-OIDC-AWS, S-BITRISE-BUILD-STATUS, S-BITRISE-VERIFIED-STEPS, S-AWS-IAM-BEST, S-AWS-OIDC, S-AWS-SCP, S-AWS-CLOUDTRAIL, S-AWS-CONFIG, S-AWS-SECURITY-HUB, S-DATADOG-CI, S-DATADOG-CD, S-DATADOG-OTEL, S-DATADOG-AUDIT, S-DATADOG-CATALOG, S-DATADOG-AWS, S-DATADOG-CLOUD-SECURITY, S-DATADOG-LLM, S-DATADOG-DORA |
| 2 | Define a shared enterprise-platform skill convention covering product/edition detection, capability negotiation, native references, permissions, credential/data-flow disclosure, conformance, rollback, drift and progressive disclosure. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-NIST-80053, S-AGENTSKILLS, S-W3C-PROV, S-AZDO-TEMPLATES, S-AZDO-APPROVALS, S-AZDO-BRANCH, S-AZDO-WORKITEMS, S-AZDO-SERVICE-CONNECTIONS, S-AZDO-REST, S-AZDO-SERVICE-HOOKS, S-BITBUCKET-CLOUD-PIPELINES, S-BITBUCKET-CLOUD-OIDC, S-BITBUCKET-CLOUD-DEPLOY, S-BITBUCKET-CLOUD-MERGE, S-BITBUCKET-CLOUD-BRANCH, S-BITBUCKET-CLOUD-SHARED, S-BITBUCKET-CLOUD-CUSTOM-CHECKS, S-BITBUCKET-CLOUD-REST, S-BITBUCKET-CLOUD-WEBHOOKS, S-BITBUCKET-DC, S-BITBUCKET-DC-REST, S-BITRISE-WORKFLOWS, S-BITRISE-SECRETS, S-BITRISE-ARTIFACTS, S-BITRISE-TRIGGERS, S-BITRISE-OIDC-AWS, S-BITRISE-BUILD-STATUS, S-BITRISE-VERIFIED-STEPS, S-AWS-IAM-BEST, S-AWS-OIDC, S-AWS-SCP, S-AWS-CLOUDTRAIL, S-AWS-CONFIG, S-AWS-SECURITY-HUB, S-DATADOG-CI, S-DATADOG-CD, S-DATADOG-OTEL, S-DATADOG-AUDIT, S-DATADOG-CATALOG, S-DATADOG-AWS, S-DATADOG-CLOUD-SECURITY, S-DATADOG-LLM, S-DATADOG-DORA |
| 3 | Create the Azure DevOps skill owner with separate references for Boards work items, Repos/pull requests and branch policies, Pipelines templates/approvals/artifacts, service connections, REST/service hooks and safe onboarding. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-AZDO-TEMPLATES, S-AZDO-APPROVALS, S-AZDO-BRANCH, S-AZDO-WORKITEMS, S-AZDO-SERVICE-CONNECTIONS, S-AZDO-REST, S-AZDO-SERVICE-HOOKS |
| 4 | Implement Azure DevOps conformance fixtures for Services and supported Server baselines, including work-item relationships, PR/build evidence, reusable pipeline templates, approvals/checks and provider-native references. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-AZDO-TEMPLATES, S-AZDO-APPROVALS, S-AZDO-BRANCH, S-AZDO-WORKITEMS, S-AZDO-SERVICE-CONNECTIONS, S-AZDO-REST, S-AZDO-SERVICE-HOOKS |
| 5 | Create the Bitbucket skill owner with an explicit Cloud versus Data Center capability matrix and no inferred parity. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-BITBUCKET-CLOUD-PIPELINES, S-BITBUCKET-DC, S-BITBUCKET-CLOUD-REST, S-BITBUCKET-DC-REST |
| 6 | Add Bitbucket Cloud references and fixtures for repositories/PRs, Pipelines, shared configurations, deployments, merge/custom checks, branch permissions, OIDC, REST and webhooks. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-BITBUCKET-CLOUD-PIPELINES, S-BITBUCKET-CLOUD-OIDC, S-BITBUCKET-CLOUD-DEPLOY, S-BITBUCKET-CLOUD-MERGE, S-BITBUCKET-CLOUD-BRANCH, S-BITBUCKET-CLOUD-SHARED, S-BITBUCKET-CLOUD-CUSTOM-CHECKS, S-BITBUCKET-CLOUD-REST, S-BITBUCKET-CLOUD-WEBHOOKS |
| 7 | Add Bitbucket Data Center references and fixtures for version detection, REST, repositories/PRs, permissions, build/deployment status, webhooks/plugins and upgrade limitations. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-BITBUCKET-DC, S-BITBUCKET-DC-REST |
| 8 | Create the Bitrise skill owner for Workflows, Pipelines, Steps and versioning, triggers, secrets, artifacts, build statuses, API/webhooks, hosted/self-hosted execution and safe onboarding. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-BITRISE-WORKFLOWS, S-BITRISE-SECRETS, S-BITRISE-ARTIFACTS, S-BITRISE-TRIGGERS, S-BITRISE-BUILD-STATUS, S-BITRISE-VERIFIED-STEPS |
| 9 | Add Bitrise-to-AWS OIDC onboarding with trust-policy preview, constrained claims, least-privilege roles, verification, rollback and prohibition of generated long-lived access keys. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-BITRISE-OIDC-AWS, S-AWS-IAM-BEST, S-AWS-OIDC |
| 10 | Create the AWS skill owner for IAM/federation, OIDC, permissions boundaries and least privilege, Organizations/SCPs, CloudTrail, Config conformance packs, Security Hub, evidence publication and break-glass-safe onboarding. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-AWS-IAM-BEST, S-AWS-OIDC, S-AWS-SCP, S-AWS-CLOUDTRAIL, S-AWS-CONFIG, S-AWS-SECURITY-HUB |
| 11 | Create the Datadog skill owner for CI/CD visibility, DORA metrics, OpenTelemetry, LLM observability, Audit Trail, Software Catalog, AWS integration, Cloud Security, evidence references and privacy-preserving telemetry onboarding. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-DATADOG-CI, S-DATADOG-CD, S-DATADOG-OTEL, S-DATADOG-AUDIT, S-DATADOG-CATALOG, S-DATADOG-AWS, S-DATADOG-CLOUD-SECURITY, S-DATADOG-LLM, S-DATADOG-DORA, S-NIST-PRIVACY, S-LAW-EU-GDPR |
| 12 | Publish a mixed-enterprise-stack composition reference using Azure Boards for work items, Bitbucket for source and pull requests, Bitrise for mobile CI, AWS for runtime/governance and Datadog for observability, while preserving independent provider-native results. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-AZDO-WORKITEMS, S-BITBUCKET-CLOUD-REST, S-BITBUCKET-DC-REST, S-BITRISE-WORKFLOWS, S-AWS-IAM-BEST, S-DATADOG-OTEL, S-W3C-PROV |
| 13 | Create idempotent setup, diagnostic, dry-run, verification, rollback and drift helpers that use native APIs/configuration and display exact permission, credential, network and data changes before mutation. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-NIST-80053, S-AWS-IAM-BEST, S-DATADOG-AUDIT, S-AZDO-TEMPLATES, S-AZDO-APPROVALS, S-AZDO-BRANCH, S-AZDO-WORKITEMS, S-AZDO-SERVICE-CONNECTIONS, S-AZDO-REST, S-AZDO-SERVICE-HOOKS, S-BITBUCKET-CLOUD-PIPELINES, S-BITBUCKET-CLOUD-OIDC, S-BITBUCKET-CLOUD-DEPLOY, S-BITBUCKET-CLOUD-MERGE, S-BITBUCKET-CLOUD-BRANCH, S-BITBUCKET-CLOUD-SHARED, S-BITBUCKET-CLOUD-CUSTOM-CHECKS, S-BITBUCKET-CLOUD-REST, S-BITBUCKET-CLOUD-WEBHOOKS, S-BITBUCKET-DC, S-BITBUCKET-DC-REST, S-BITRISE-WORKFLOWS, S-BITRISE-SECRETS, S-BITRISE-ARTIFACTS, S-BITRISE-TRIGGERS, S-BITRISE-OIDC-AWS, S-BITRISE-BUILD-STATUS, S-BITRISE-VERIFIED-STEPS, S-AWS-IAM-BEST, S-AWS-OIDC, S-AWS-SCP, S-AWS-CLOUDTRAIL, S-AWS-CONFIG, S-AWS-SECURITY-HUB, S-DATADOG-CI, S-DATADOG-CD, S-DATADOG-OTEL, S-DATADOG-AUDIT, S-DATADOG-CATALOG, S-DATADOG-AWS, S-DATADOG-CLOUD-SECURITY, S-DATADOG-LLM, S-DATADOG-DORA |
| 14 | Add positive, negative and near-miss trigger evals plus live/pinned conformance fixtures for editions, tiers, unsupported capabilities, OIDC claims, secret exposure, artifact/status linkage, telemetry redaction and provider outages. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-AZDO-TEMPLATES, S-AZDO-APPROVALS, S-AZDO-BRANCH, S-AZDO-WORKITEMS, S-AZDO-SERVICE-CONNECTIONS, S-AZDO-REST, S-AZDO-SERVICE-HOOKS, S-BITBUCKET-CLOUD-PIPELINES, S-BITBUCKET-CLOUD-OIDC, S-BITBUCKET-CLOUD-DEPLOY, S-BITBUCKET-CLOUD-MERGE, S-BITBUCKET-CLOUD-BRANCH, S-BITBUCKET-CLOUD-SHARED, S-BITBUCKET-CLOUD-CUSTOM-CHECKS, S-BITBUCKET-CLOUD-REST, S-BITBUCKET-CLOUD-WEBHOOKS, S-BITBUCKET-DC, S-BITBUCKET-DC-REST, S-BITRISE-WORKFLOWS, S-BITRISE-SECRETS, S-BITRISE-ARTIFACTS, S-BITRISE-TRIGGERS, S-BITRISE-OIDC-AWS, S-BITRISE-BUILD-STATUS, S-BITRISE-VERIFIED-STEPS, S-AWS-IAM-BEST, S-AWS-OIDC, S-AWS-SCP, S-AWS-CLOUDTRAIL, S-AWS-CONFIG, S-AWS-SECURITY-HUB, S-DATADOG-CI, S-DATADOG-CD, S-DATADOG-OTEL, S-DATADOG-AUDIT, S-DATADOG-CATALOG, S-DATADOG-AWS, S-DATADOG-CLOUD-SECURITY, S-DATADOG-LLM, S-DATADOG-DORA |
| 15 | Register the five optional skill packages alphabetically, pin official sources and tested versions, update package/marketplace metadata, and publish migration, coexistence and uninstall guidance. | D-036, D-037, D-038, D-039, D-003, D-009, D-013, D-014, D-015, D-025, D-026, D-027, D-030, D-033 | C-041, C-042, C-043, C-044, C-045, C-004, C-007, C-009, C-011, C-015, C-027, C-028, C-030, C-033, C-034, C-036, C-037, C-038, C-040 | S-AGENTSKILLS, S-AZDO-TEMPLATES, S-AZDO-APPROVALS, S-AZDO-BRANCH, S-AZDO-WORKITEMS, S-AZDO-SERVICE-CONNECTIONS, S-AZDO-REST, S-AZDO-SERVICE-HOOKS, S-BITBUCKET-CLOUD-PIPELINES, S-BITBUCKET-CLOUD-OIDC, S-BITBUCKET-CLOUD-DEPLOY, S-BITBUCKET-CLOUD-MERGE, S-BITBUCKET-CLOUD-BRANCH, S-BITBUCKET-CLOUD-SHARED, S-BITBUCKET-CLOUD-CUSTOM-CHECKS, S-BITBUCKET-CLOUD-REST, S-BITBUCKET-CLOUD-WEBHOOKS, S-BITBUCKET-DC, S-BITBUCKET-DC-REST, S-BITRISE-WORKFLOWS, S-BITRISE-SECRETS, S-BITRISE-ARTIFACTS, S-BITRISE-TRIGGERS, S-BITRISE-OIDC-AWS, S-BITRISE-BUILD-STATUS, S-BITRISE-VERIFIED-STEPS, S-AWS-IAM-BEST, S-AWS-OIDC, S-AWS-SCP, S-AWS-CLOUDTRAIL, S-AWS-CONFIG, S-AWS-SECURITY-HUB, S-DATADOG-CI, S-DATADOG-CD, S-DATADOG-OTEL, S-DATADOG-AUDIT, S-DATADOG-CATALOG, S-DATADOG-AWS, S-DATADOG-CLOUD-SECURITY, S-DATADOG-LLM, S-DATADOG-DORA |

Mapping IDs resolve through the parent plan’s `traceability/` artifacts. Platform documentation supports candidate capabilities; release requires edition/version-pinned conformance. The mappings do not imply certification or compliance.

## Validation Steps

1. Validate each skill package and progressive-disclosure references.
2. Run source-link and official-domain checks.
3. Run edition/version capability probes and negative unsupported-feature tests.
4. Verify OIDC/federated AWS access and absence of generated long-lived keys.
5. Verify provider-native work-item, PR, build, artifact, deployment, audit and telemetry references.
6. Run privacy/redaction, permission, rollback, drift, outage and rate-limit tests.
7. Exercise the mixed-stack composition end to end without a central result store.

## Success Criteria

- [ ] Five optional skill owners exist and can be installed independently.
- [ ] Azure DevOps Services/Server and Bitbucket Cloud/Data Center are detected and documented separately.
- [ ] Azure Boards, Bitbucket source/PRs, Bitrise CI, AWS runtime/governance and Datadog evidence compose through opaque native references.
- [ ] CI-to-AWS onboarding uses temporary federated credentials where supported and never creates static keys by default.
- [ ] Datadog onboarding explicitly controls content, redaction, retention, residency, access, sampling and cost.
- [ ] Every capability claim is backed by official documentation and a pinned-version conformance test.
- [ ] Setup is previewable, authorized, idempotent, verifiable, reversible and drift-aware.
- [ ] Unsupported editions, tiers and features fail clearly without silent fallback.

## Files Modified/Created

- `packages/skill/skill-azure-devops/**`
- `packages/skill/skill-bitbucket/**`
- `packages/skill/skill-bitrise/**`
- `packages/skill/skill-aws/**`
- `packages/skill/skill-datadog/**`
- marketplace and lockfile entries
- platform conformance fixtures and mixed-stack composition tests

## Dependencies

Depends on the common workflow/governance contracts and provider/module conformance. It is optional and runs in parallel with harness adapters and generic external automation. The walking skeleton, cross-cutting governance and final validation consume its fixtures when the enterprise-stack profile is selected.

## Estimated Duration

Large. Five focused skill owners plus a mixed-stack integration fixture; implementation may be split internally by platform while retaining one coordinated subplan and shared conformance vocabulary.
