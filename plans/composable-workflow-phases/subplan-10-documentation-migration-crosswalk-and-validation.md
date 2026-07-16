---
type: plan
has_subplans: false
parent_plan: ../composable-workflow-phases.md
parallel_group: 7
status: pending
dependencies:
  plans:
  - governance-policy-learning-observability-trust-and-operations
  - enterprise-platform-skills-and-onboarding
  files:
  - .claude-plugin/marketplace.json
  - .agents/plugins/marketplace.json
  - package-lock.json
  - packages/command/command-workflow/**
  - packages/diagram/diagram-agent-workflow/**
  - README.md
skills_to_consult:
- versioning-guide
- instruction-guide
- command-guide
- skill-guide
- testing-guide
- git-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
updated: '2026-07-14'
decision_refs:
- D-001
- D-002
- D-003
- D-004
- D-005
- D-006
- D-007
- D-008
- D-009
- D-010
- D-011
- D-012
- D-013
- D-014
- D-015
- D-016
- D-017
- D-018
- D-019
- D-020
- D-021
- D-022
- D-023
- D-024
- D-025
- D-026
- D-027
- D-028
- D-029
- D-030
- D-031
- D-032
- D-033
- D-034
- D-035
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
- C-006
- C-007
- C-008
- C-009
- C-010
- C-011
- C-012
- C-013
- C-014
- C-015
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
- C-028
- C-029
- C-030
- C-031
- C-032
- C-033
- C-034
- C-035
- C-036
- C-037
- C-038
- C-039
- C-040
- C-041
- C-042
- C-043
- C-044
- C-045
source_refs:
- S-AGENTSKILLS
- S-CODEX-MANAGED
- S-CODEX-PLUGINS
- S-CODEX-SKILLS
- S-CYCLONEDX-MLBOM
- S-DORA-CAPS
- S-GITHUB-COMPOSITE
- S-GITHUB-ENV
- S-GITHUB-REUSABLE
- S-GITHUB-RULESETS
- S-GITHUB-SECURE
- S-GITLAB-COMPLIANCE
- S-GITLAB-COMPONENTS
- S-GITLAB-PIPELINE-POLICY
- S-HARNESS-CLAUDE
- S-HARNESS-CODEX
- S-HARNESS-COPILOT
- S-HARNESS-KIRO
- S-HARNESS-OPENCODE
- S-HARNESS-PI
- S-HEXAGONAL
- S-IN-TOTO
- S-ISO-12207
- S-ISO-15288
- S-ISO-23894
- S-ISO-42001
- S-ISO-5338
- S-ISO-9241-210
- S-LAW-EU-AIACT
- S-LAW-EU-CRA
- S-LAW-EU-DORA
- S-LAW-EU-GDPR
- S-LAW-EU-NIS2
- S-MCP
- S-NIST-800207
- S-NIST-800218
- S-NIST-800218A
- S-NIST-80053
- S-NIST-80061
- S-NIST-AIRMF
- S-NIST-CSF
- S-NIST-PRIVACY
- S-OPA
- S-OPENSSF-SCORECARD
- S-OTEL-GENAI
- S-OWASP-ASVS
- S-OWASP-LLM
- S-OWASP-SAMM
- S-SIGSTORE
- S-SLSA
- S-SPDX-AI
- S-TUF
- S-W3C-PROV
- S-W3C-WCAG22
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

# Documentation, Migration, Crosswalk, and Validation

## Objective

Reconcile the expanded workflow and governance architecture across packages, platform skills, provider owners, onboarding modules, CI/CD, policies, diagrams, source/control crosswalks, migration, evals, versions, and release artifacts. Make the modularity and boundaries understandable to users without implying that every module, hook, skill, provider, or lifecycle capability is mandatory.

## Tasks

1. Reconcile manifests, dependencies, marketplaces, package metadata, Moon configuration, and lockfile for workflow, governance, platform, and provider modules.
2. Publish the two-plane architecture and explain profiles, executors, hooks, external enforcement, onboarding, providers, evidence, trust, and observability.
3. Publish a composition guide covering workflow-only, governance-only, enablement-only, external-enforcement-only, and integrated adoption, plus a decision table for script, script-plus-model, bounded agent, human, and external execution.
4. Publish a harness capability matrix for Claude Code, Codex, Kiro, Copilot CLI/cloud agent, Pi, and OpenCode, with tested version/date and unsupported/experimental features.
5. Publish harness-specific onboarding guides with preview, permissions, native paths/config scopes, trust, dry-run, diagnostics, rollback, and drift.
6. Publish GitHub, GitLab, Azure DevOps, Bitbucket, Bitrise, AWS and Datadog onboarding guides using reusable workflows/actions/rulesets/environments and components/pipeline policies/compliance frameworks.
7. Publish execution-selection guidance: deterministic script, script plus bounded LLM, bounded agent, human, or external system.
8. Publish security guidance for project executable modules, plugins/extensions, MCP servers, secrets, model data flow, least privilege, sandboxing, and supply-chain provenance.
9. Publish policy-decision/enforcement patterns, OPA as optional implementation, defense in depth, exception, and break-glass guidance.
10. Update lifecycle command migration and document that governance/onboarding modules operate independently of lifecycle operations.
11. Redesign diagrams with separate workflow and governance planes, semantic event intents, native adapters, external enforcement, provider-native evidence, onboarding, and feedback.
12. Publish and validate the authoritative source registry, decision–source matrix, control/article crosswalk, platform capability matrix, subplan task traceability and mapping-status caveats; never treat crosswalks as certification or equivalence.
13. Add full test/eval matrices for platform compatibility including the five enterprise skills, hooks, modules, onboarding, policy, concurrency, idempotency, agent recursion, authority, CI, providers, lifecycle, drift, rollback, and telemetry.
14. Run static searches to prevent claims of universal hook parity, skills-as-enforcement, mandatory YAML, hidden agent launches, or automatic compliance.
15. Validate the deliberate eleven-subplan exception and apply breaking-release versioning and publish migration/rollback notes.

## Traceability

| Task | Implementation intent | Decision IDs | Control IDs | Source IDs |
| --- | --- | --- | --- | --- |
| 1 | Reconcile manifests, dependencies, marketplaces, package metadata, Moon configuration, and lockfile for workflow, governance, platform, and provider modules. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 2 | Publish the two-plane architecture and explain profiles, executors, hooks, external enforcement, onboarding, providers, evidence, trust, and observability. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 3 | Publish a composition guide covering workflow-only, governance-only, enablement-only, external-enforcement-only, and integrated adoption, plus a decision table for script, script-plus-model, bounded agent, human, and external execution. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 4 | Publish a harness capability matrix for Claude Code, Codex, Kiro, Copilot CLI/cloud agent, Pi, and OpenCode, with tested version/date and unsupported/experimental features. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 5 | Publish harness-specific onboarding guides with preview, permissions, native paths/config scopes, trust, dry-run, diagnostics, rollback, and drift. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 6 | Publish GitHub, GitLab, Azure DevOps, Bitbucket, Bitrise, AWS and Datadog onboarding guides using reusable workflows/actions/rulesets/environments and components/pipeline policies/compliance frameworks. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 7 | Publish execution-selection guidance: deterministic script, script plus bounded LLM, bounded agent, human, or external system. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 8 | Publish security guidance for project executable modules, plugins/extensions, MCP servers, secrets, model data flow, least privilege, sandboxing, and supply-chain provenance. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 9 | Publish policy-decision/enforcement patterns, OPA as optional implementation, defense in depth, exception, and break-glass guidance. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 10 | Update lifecycle command migration and document that governance/onboarding modules operate independently of lifecycle operations. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 11 | Redesign diagrams with separate workflow and governance planes, semantic event intents, native adapters, external enforcement, provider-native evidence, onboarding, and feedback. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 12 | Add source/control crosswalks and legal/compliance caveats from the prior approved plan. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 13 | Add full test/eval matrices for platform compatibility including the five enterprise skills, hooks, modules, onboarding, policy, concurrency, idempotency, agent recursion, authority, CI, providers, lifecycle, drift, rollback, and telemetry. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 14 | Run static searches to prevent claims of universal hook parity, skills-as-enforcement, mandatory YAML, hidden agent launches, or automatic compliance. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |
| 15 | Validate the deliberate eleven-subplan exception and apply breaking-release versioning and publish migration/rollback notes. | D-024, D-025, D-026, D-030, D-004, D-005, D-010, D-013, D-014, D-015, D-023, D-027, D-033, D-034 | C-027, C-037, C-040, C-001, C-009, C-015, C-028, C-030, C-031, C-033, C-038 | S-NIST-80053, S-ISO-12207, S-AGENTSKILLS, S-CODEX-MANAGED, S-CODEX-PLUGINS, S-CODEX-SKILLS, S-CYCLONEDX-MLBOM, S-DORA-CAPS, S-GITHUB-COMPOSITE, S-GITHUB-ENV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITHUB-SECURE, S-GITLAB-COMPLIANCE |

Mapping IDs resolve through the parent plan’s `traceability/` artifacts. These mappings do not imply equivalence, certification or legal compliance.

## Validation Steps

1. `npm run fmt:check`, lint, typecheck, build, and test.
2. Skill validators and trigger/output evals.
3. Harness capability and onboarding conformance suites.
4. GitHub/GitLab external enforcement fixtures.
5. Provider/policy/module conformance and walking skeleton.
6. Lifecycle result/profile/crosswalk/evidence integration tests.
7. Diagram build and generated-asset diff.
8. Run automated source/decision/control/task traceability validation, link/version/supersession review, legal-current-text flags, licensed-ISO verification flags, and platform conformance checks against primary documentation.

## Success Criteria

- [ ] Documentation clearly explains the two independent composable planes and all supported adoption modes.
- [ ] Every platform guide is versioned and names limitations.
- [ ] Users can choose individual modules or presets and see the effective composition.
- [ ] Onboarding guidance always includes preview, consent, verification, rollback, and trust.
- [ ] Skills are not represented as proof of enforcement.
- [ ] CI and harness examples use native reusable mechanisms and least privilege.
- [ ] No example introduces mandatory workflow YAML/sidecars or a universal hook schema.
- [ ] Every decision, control, platform claim and numbered subplan task has resolvable traceability.
- [ ] Legal, licensed-standard, vendor-conformance and synthesis statuses remain explicit.
- [ ] Full validation passes and breaking migration is complete.

## Files Modified/Created

- Marketplaces, package metadata, lockfile, README/AGENTS guides
- Workflow and governance commands/skills/platform references
- Diagram source and generated assets
- Compatibility, crosswalk, migration, security, and onboarding documentation

## Dependencies

Depends on all previous groups.

## Estimated Duration

Very large: repository-wide reconciliation and exhaustive validation.
