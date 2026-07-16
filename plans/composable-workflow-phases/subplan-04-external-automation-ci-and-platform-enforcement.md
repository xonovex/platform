---
type: plan
has_subplans: false
parent_plan: ../composable-workflow-phases.md
parallel_group: 3
status: complete
dependencies:
  plans:
  - provider-policy-and-module-conformance
  files:
  - packages/skill/skill-github/**
  - packages/skill/skill-gitlab/**
  - packages/agent/agent-operator-go/**
  - packages/skill/skill-agent-governance/references/external-enforcement/**
skills_to_consult:
- skill-guide
- testing-guide
- git-guide
- security-assurance-guide
- versioning-guide
validation:
  type_check: passed
  lint: passed
  build: passed
  tests: passed
  integration: passed
updated: '2026-07-16'
decision_refs:
- D-006
- D-007
- D-012
- D-013
- D-018
- D-028
- D-030
- D-034
control_refs:
- C-002
- C-005
- C-006
- C-010
- C-011
- C-015
- C-032
- C-035
- C-037
- C-038
source_refs:
- S-GITHUB-REUSABLE
- S-GITHUB-COMPOSITE
- S-GITHUB-RULESETS
- S-GITHUB-ENV
- S-GITHUB-SECURE
- S-GITLAB-COMPONENTS
- S-GITLAB-PIPELINE-POLICY
- S-GITLAB-COMPLIANCE
- S-NIST-80053
traceability_files:
- traceability/source-registry.md
- traceability/decision-source-matrix.md
- traceability/control-crosswalk.md
- traceability/subplan-traceability.md
---

# External Automation, CI/CD, and Platform Enforcement

## Objective

Implement modular enforcement and onboarding outside agent harnesses. Extend GitHub and GitLab provider owners with native reusable CI/CD and policy mechanisms, connect Xonovex AgentPolicy/admission controls, and define optional integrations for Git hooks, pre-commit, policy decision services, secrets, deployments, and GRC systems.

## Tasks

1. Define an external-enforcement adapter contract mapping semantic policy intents to CI jobs, repository rules, protected environments, admission controls, provider permissions, or external services.
2. Add GitHub references and modules for reusable workflows, composite actions, version/SHA pinning, required checks, rulesets, protected environments, reviewers, least-privilege tokens, secure use, attestations, and evidence publication.
3. Add GitLab references and modules for versioned CI/CD components, typed inputs, component testing, dependency pinning, pipeline execution policies, compliance frameworks, security-policy projects, protected environments, and evidence publication.
4. Document provider-native merge/precedence semantics and avoid hidden global configuration or copied boilerplate.
5. Add generic Git hook/pre-commit patterns for advisory local checks while clarifying that client-side hooks are bypassable and not sufficient for mandatory governance.
6. Map Xonovex AgentPolicy and admission webhooks to runtime class, security context, network, timeout/resource, image, toolchain, and namespace-level enforcement intents.
7. Define optional policy-decision-service integration, including OPA as one implementation, with cached/failed decision behavior and historical policy version evidence.
8. Define privileged-operation controls for Integration, Release, secrets, data deletion, infrastructure, and Retirement at external enforcement points.
9. Create onboarding modules that discover existing CI/rules/policies, propose reusable components, preview changes, apply natively, verify, and roll back.
10. Provide governance-only adoption recipes where GitHub/GitLab CI and repository controls enforce selected policies without installing agent-harness or lifecycle modules.
11. Add fixtures for GitHub and GitLab plus one Kubernetes admission scenario; test bypass attempts and independent enforcement layers.

## Traceability

| Task | Implementation intent | Decision IDs | Control IDs | Source IDs |
| --- | --- | --- | --- | --- |
| 1 | Define an external-enforcement adapter contract mapping semantic policy intents to CI jobs, repository rules, protected environments, admission controls, provider permissions, or external services. | D-006, D-007, D-012, D-013, D-018, D-028, D-030, D-034, D-004, D-005, D-017, D-015, D-033 | C-002, C-005, C-006, C-010, C-011, C-015, C-032, C-035, C-037, C-038, C-028, C-034 | S-GITHUB-REUSABLE, S-GITHUB-COMPOSITE, S-GITHUB-RULESETS, S-GITHUB-ENV, S-GITHUB-SECURE, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-GITLAB-COMPLIANCE, S-NIST-80053, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 2 | Add GitHub references and modules for reusable workflows, composite actions, version/SHA pinning, required checks, rulesets, protected environments, reviewers, least-privilege tokens, secure use, attestations, and evidence publication. | D-006, D-007, D-012, D-013, D-018, D-028, D-030, D-034, D-004, D-005, D-017, D-015, D-033 | C-002, C-005, C-006, C-010, C-011, C-015, C-032, C-035, C-037, C-038, C-028, C-034 | S-GITHUB-REUSABLE, S-GITHUB-COMPOSITE, S-GITHUB-RULESETS, S-GITHUB-ENV, S-GITHUB-SECURE, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-GITLAB-COMPLIANCE, S-NIST-80053, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 3 | Add GitLab references and modules for versioned CI/CD components, typed inputs, component testing, dependency pinning, pipeline execution policies, compliance frameworks, security-policy projects, protected environments, and evidence publication. | D-006, D-007, D-012, D-013, D-018, D-028, D-030, D-034, D-004, D-005, D-017, D-023, D-024, D-025, D-026, D-015, D-033 | C-002, C-005, C-006, C-010, C-011, C-015, C-032, C-035, C-037, C-038, C-001, C-027, C-028, C-034 | S-GITHUB-REUSABLE, S-GITHUB-COMPOSITE, S-GITHUB-RULESETS, S-GITHUB-ENV, S-GITHUB-SECURE, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-GITLAB-COMPLIANCE, S-NIST-80053, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-LAW-EU-GDPR, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 4 | Document provider-native merge/precedence semantics and avoid hidden global configuration or copied boilerplate. | D-006, D-007, D-012, D-013, D-018, D-028, D-030, D-034, D-004, D-005, D-017, D-015, D-033 | C-002, C-005, C-006, C-010, C-011, C-015, C-032, C-035, C-037, C-038, C-028, C-034 | S-GITHUB-REUSABLE, S-GITHUB-COMPOSITE, S-GITHUB-RULESETS, S-GITHUB-ENV, S-GITHUB-SECURE, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-GITLAB-COMPLIANCE, S-NIST-80053, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 5 | Add generic Git hook/pre-commit patterns for advisory local checks while clarifying that client-side hooks are bypassable and not sufficient for mandatory governance. | D-006, D-007, D-012, D-013, D-018, D-028, D-030, D-034, D-004, D-005, D-017 | C-002, C-005, C-006, C-010, C-011, C-015, C-032, C-035, C-037, C-038 | S-GITHUB-REUSABLE, S-GITHUB-COMPOSITE, S-GITHUB-RULESETS, S-GITHUB-ENV, S-GITHUB-SECURE, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-GITLAB-COMPLIANCE, S-NIST-80053, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI |
| 6 | Map Xonovex AgentPolicy and admission webhooks to runtime class, security context, network, timeout/resource, image, toolchain, and namespace-level enforcement intents. | D-006, D-007, D-012, D-013, D-018, D-028, D-030, D-034, D-004, D-005, D-017 | C-002, C-005, C-006, C-010, C-011, C-015, C-032, C-035, C-037, C-038 | S-GITHUB-REUSABLE, S-GITHUB-COMPOSITE, S-GITHUB-RULESETS, S-GITHUB-ENV, S-GITHUB-SECURE, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-GITLAB-COMPLIANCE, S-NIST-80053, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI |
| 7 | Define optional policy-decision-service integration, including OPA as one implementation, with cached/failed decision behavior and historical policy version evidence. | D-006, D-007, D-012, D-013, D-018, D-028, D-030, D-034 | C-002, C-005, C-006, C-010, C-011, C-015, C-032, C-035, C-037, C-038 | S-GITHUB-REUSABLE, S-GITHUB-COMPOSITE, S-GITHUB-RULESETS, S-GITHUB-ENV, S-GITHUB-SECURE, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-GITLAB-COMPLIANCE, S-NIST-80053 |
| 8 | Define privileged-operation controls for Integration, Release, secrets, data deletion, infrastructure, and Retirement at external enforcement points. | D-006, D-007, D-012, D-013, D-018, D-028, D-030, D-034, D-022 | C-002, C-005, C-006, C-010, C-011, C-015, C-032, C-035, C-037, C-038, C-014, C-040 | S-GITHUB-REUSABLE, S-GITHUB-COMPOSITE, S-GITHUB-RULESETS, S-GITHUB-ENV, S-GITHUB-SECURE, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-GITLAB-COMPLIANCE, S-NIST-80053, S-ISO-12207, S-ISO-15288 |
| 9 | Create onboarding modules that discover existing CI/rules/policies, propose reusable components, preview changes, apply natively, verify, and roll back. | D-006, D-007, D-012, D-013, D-018, D-028, D-030, D-034 | C-002, C-005, C-006, C-010, C-011, C-015, C-032, C-035, C-037, C-038 | S-GITHUB-REUSABLE, S-GITHUB-COMPOSITE, S-GITHUB-RULESETS, S-GITHUB-ENV, S-GITHUB-SECURE, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-GITLAB-COMPLIANCE, S-NIST-80053 |
| 10 | Provide governance-only adoption recipes where GitHub/GitLab CI and repository controls enforce selected policies without installing agent-harness or lifecycle modules. | D-006, D-007, D-012, D-013, D-018, D-028, D-030, D-034, D-004, D-005, D-017 | C-002, C-005, C-006, C-010, C-011, C-015, C-032, C-035, C-037, C-038 | S-GITHUB-REUSABLE, S-GITHUB-COMPOSITE, S-GITHUB-RULESETS, S-GITHUB-ENV, S-GITHUB-SECURE, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-GITLAB-COMPLIANCE, S-NIST-80053, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI |
| 11 | Add fixtures for GitHub and GitLab plus one Kubernetes admission scenario; test bypass attempts and independent enforcement layers. | D-006, D-007, D-012, D-013, D-018, D-028, D-030, D-034 | C-002, C-005, C-006, C-010, C-011, C-015, C-032, C-035, C-037, C-038 | S-GITHUB-REUSABLE, S-GITHUB-COMPOSITE, S-GITHUB-RULESETS, S-GITHUB-ENV, S-GITHUB-SECURE, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-GITLAB-COMPLIANCE, S-NIST-80053 |

Mapping IDs resolve through the parent plan’s `traceability/` artifacts. These mappings do not imply equivalence, certification or legal compliance.

## Validation Steps

1. Test GitHub reusable workflow/action and ruleset fixtures with version-pinned references.
2. Test GitLab component, pipeline-policy, and compliance-framework fixtures including naming/merge conflicts.
3. Test mandatory policy at both harness and external control layers.
4. Test least privilege, environment approval, rollback, and policy service outage behavior.
5. Verify no platform template is treated as a universal workflow representation.

## Success Criteria

- [x] GitHub and GitLab onboarding uses native reusable/versioned mechanisms.
- [x] External enforcement is independent from harness hooks and can provide defense in depth.
- [x] AgentPolicy/admission controls map to semantic governance intents.
- [x] Privileged actions have enforceable external gates.
- [x] Setup modules are previewable, reversible, and provider-native.
- [x] Platform-specific merge, precedence, and failure behavior is documented and tested.

## Files Modified/Created

- GitHub and GitLab provider skill references/modules/evals
- Agent operator policy/admission documentation and tests
- External-enforcement references and fixtures

## Dependencies

Depends on contracts and module conformance; may run in parallel with harness adapters.

## Estimated Duration

Very large: multiple platform-native control systems and integration fixtures.
