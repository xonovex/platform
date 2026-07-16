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
    - packages/command/command-workflow/commands/acceptance-*.md
    - packages/command/command-workflow/commands/integration-*.md
    - packages/command/command-workflow/commands/transition-*.md
    - packages/command/command-workflow/commands/release-*.md
    - packages/command/command-workflow/commands/observe-*.md
    - packages/command/command-workflow/commands/incident-*.md
    - packages/command/command-workflow/commands/corrective-action-*.md
    - packages/command/command-workflow/commands/retirement-*.md
skills_to_consult:
  - command-guide
  - skill-guide
  - testing-guide
  - security-assurance-guide
  - reliability-guide
validation:
  type_check: not_applicable
  lint: passed
  build: passed
  tests: passed
  integration: passed
updated: "2026-07-16"
decision_refs:
  - D-020
  - D-021
  - D-022
  - D-023
  - D-024
  - D-026
  - D-029
  - D-032
  - D-035
control_refs:
  - C-001
  - C-003
  - C-012
  - C-013
  - C-014
  - C-020
  - C-022
  - C-024
  - C-025
  - C-027
  - C-034
  - C-035
  - C-038
  - C-040
source_refs:
  - S-NIST-80053
  - S-NIST-80061
  - S-LAW-EU-AIACT
  - S-LAW-EU-DORA
  - S-LAW-EU-CRA
  - S-LAW-EU-NIS2
  - S-GITHUB-ENV
  - S-ISO-12207
  - S-ISO-15288
traceability_files:
  - traceability/source-registry.md
  - traceability/decision-source-matrix.md
  - traceability/control-crosswalk.md
  - traceability/subplan-traceability.md
---

# Acceptance, Integration, Transition, Release, Observation, Incidents, and Retirement

## Objective

Implement accountable human decisions and privileged/operational capabilities with exact-revision authorization, external enforcement, least privilege, transition/rollback, observation, incident/corrective-action loops, and safe retirement. Agents may prepare evidence or advice, but authority and target-changing operations remain explicit. The lifecycle capabilities must remain usable in a workflow-only composition; governance hooks and external controls strengthen execution when selected but are not implicit dependencies unless a profile explicitly requires them.

## Tasks

1. Implement Acceptance evidence assembly separately from human sign-off.
2. Bind authorization to exact Deliverable revision, target, actor, evidence, policy version, and expiry.
3. Implement Integration validation and execution through provider/external enforcement; ordinary tool calls cannot bypass the explicit capability.
4. Define Transition planning/execution/verification/rollback for data, users, providers, flags, training, support, and resilience.
5. Implement Release through controlled automation and protected environments rather than autonomous agent side effects.
6. Implement Observation from monitoring, user, security, AI, cost, accessibility, and DORA outcome evidence.
7. Implement urgent Incident and Corrective Action results, including reporting applicability, containment, recovery, postmortem specialization, verification, and learning.
8. Implement Retirement for models, data, credentials, features, APIs, infrastructure, dependencies, and provider configurations.
9. Map privileged semantic intents to harness and external enforcement points with fail-closed behavior where required.
10. Add break-glass and exception flows with expiry, review, compensating controls, and post-use audit.
11. Add agent-assistance patterns for evidence summaries and incident investigation while prohibiting agents from fabricating human authorization or regulatory conclusions.
12. Test target drift, stale evidence, policy drift, release failure, rollback, incident escalation, emergency bypass, data deletion, and retirement verification.

## Traceability

| Task | Implementation intent                                                                                                                                                     | Decision IDs                                                                                                          | Control IDs                                                                                                                                       | Source IDs                                                                                                                                                                                                                                                                                               |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1    | Implement Acceptance evidence assembly separately from human sign-off.                                                                                                    | D-020, D-021, D-022, D-023, D-024, D-026, D-029, D-032, D-035                                                         | C-001, C-003, C-012, C-013, C-014, C-020, C-022, C-024, C-025, C-027, C-034, C-035, C-038, C-040                                                  | S-NIST-80053, S-NIST-80061, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-GITHUB-ENV, S-ISO-12207, S-ISO-15288                                                                                                                                                                           |
| 2    | Bind authorization to exact Deliverable revision, target, actor, evidence, policy version, and expiry.                                                                    | D-020, D-021, D-022, D-023, D-024, D-026, D-029, D-032, D-035, D-004, D-005, D-017                                    | C-001, C-003, C-012, C-013, C-014, C-020, C-022, C-024, C-025, C-027, C-034, C-035, C-038, C-040, C-006, C-037                                    | S-NIST-80053, S-NIST-80061, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-GITHUB-ENV, S-ISO-12207, S-ISO-15288, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI                                                                   |
| 3    | Implement Integration validation and execution through provider/external enforcement; ordinary tool calls cannot bypass the explicit capability.                          | D-020, D-021, D-022, D-023, D-024, D-026, D-029, D-032, D-035, D-015, D-033, D-034                                    | C-001, C-003, C-012, C-013, C-014, C-020, C-022, C-024, C-025, C-027, C-034, C-035, C-038, C-040, C-028                                           | S-NIST-80053, S-NIST-80061, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-GITHUB-ENV, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL                                                                                                                                                  |
| 4    | Define Transition planning/execution/verification/rollback for data, users, providers, flags, training, support, and resilience.                                          | D-020, D-021, D-022, D-023, D-024, D-026, D-029, D-032, D-035, D-030, D-015, D-033, D-034                             | C-001, C-003, C-012, C-013, C-014, C-020, C-022, C-024, C-025, C-027, C-034, C-035, C-038, C-040, C-028                                           | S-NIST-80053, S-NIST-80061, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-GITHUB-ENV, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL                                                                                                                                                  |
| 5    | Implement Release through controlled automation and protected environments rather than autonomous agent side effects.                                                     | D-020, D-021, D-022, D-023, D-024, D-026, D-029, D-032, D-035, D-013, D-028, D-030                                    | C-001, C-003, C-012, C-013, C-014, C-020, C-022, C-024, C-025, C-027, C-034, C-035, C-038, C-040, C-015, C-037                                    | S-NIST-80053, S-NIST-80061, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-GITHUB-ENV, S-ISO-12207, S-ISO-15288, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY                                                                                      |
| 6    | Implement Observation from monitoring, user, security, AI, cost, accessibility, and DORA outcome evidence.                                                                | D-020, D-021, D-022, D-023, D-024, D-026, D-029, D-032, D-035                                                         | C-001, C-003, C-012, C-013, C-014, C-020, C-022, C-024, C-025, C-027, C-034, C-035, C-038, C-040, C-026                                           | S-NIST-80053, S-NIST-80061, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-GITHUB-ENV, S-ISO-12207, S-ISO-15288, S-W3C-WCAG22, S-ISO-9241-210                                                                                                                                             |
| 7    | Implement urgent Incident and Corrective Action results, including reporting applicability, containment, recovery, postmortem specialization, verification, and learning. | D-020, D-021, D-022, D-023, D-024, D-026, D-029, D-032, D-035, D-025, D-030                                           | C-001, C-003, C-012, C-013, C-014, C-020, C-022, C-024, C-025, C-027, C-034, C-035, C-038, C-040                                                  | S-NIST-80053, S-NIST-80061, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-GITHUB-ENV, S-ISO-12207, S-ISO-15288, S-LAW-EU-GDPR                                                                                                                                                            |
| 8    | Implement Retirement for models, data, credentials, features, APIs, infrastructure, dependencies, and provider configurations.                                            | D-020, D-021, D-022, D-023, D-024, D-026, D-029, D-032, D-035, D-031, D-004, D-005, D-017, D-030, D-015, D-033, D-034 | C-001, C-003, C-012, C-013, C-014, C-020, C-022, C-024, C-025, C-027, C-034, C-035, C-038, C-040, C-016, C-017, C-018, C-021, C-006, C-037, C-028 | S-NIST-80053, S-NIST-80061, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-GITHUB-ENV, S-ISO-12207, S-ISO-15288, S-NIST-AIRMF, S-NIST-800218A, S-SPDX-AI, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-W3C-PROV, S-HEXAGONAL |
| 9    | Map privileged semantic intents to harness and external enforcement points with fail-closed behavior where required.                                                      | D-020, D-021, D-022, D-023, D-024, D-026, D-029, D-032, D-035, D-004, D-005, D-017                                    | C-001, C-003, C-012, C-013, C-014, C-020, C-022, C-024, C-025, C-027, C-034, C-035, C-038, C-040, C-006, C-037                                    | S-NIST-80053, S-NIST-80061, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-GITHUB-ENV, S-ISO-12207, S-ISO-15288, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI                                                                   |
| 10   | Add break-glass and exception flows with expiry, review, compensating controls, and post-use audit.                                                                       | D-020, D-021, D-022, D-023, D-024, D-026, D-029, D-032, D-035, D-004, D-005, D-017                                    | C-001, C-003, C-012, C-013, C-014, C-020, C-022, C-024, C-025, C-027, C-034, C-035, C-038, C-040, C-006, C-037                                    | S-NIST-80053, S-NIST-80061, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-GITHUB-ENV, S-ISO-12207, S-ISO-15288, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI                                                                   |
| 11   | Add agent-assistance patterns for evidence summaries and incident investigation while prohibiting agents from fabricating human authorization or regulatory conclusions.  | D-020, D-021, D-022, D-023, D-024, D-026, D-029, D-032, D-035, D-025                                                  | C-001, C-003, C-012, C-013, C-014, C-020, C-022, C-024, C-025, C-027, C-034, C-035, C-038, C-040                                                  | S-NIST-80053, S-NIST-80061, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-GITHUB-ENV, S-ISO-12207, S-ISO-15288, S-LAW-EU-GDPR                                                                                                                                                            |
| 12   | Test target drift, stale evidence, policy drift, release failure, rollback, incident escalation, emergency bypass, data deletion, and retirement verification.            | D-020, D-021, D-022, D-023, D-024, D-026, D-029, D-032, D-035, D-030                                                  | C-001, C-003, C-012, C-013, C-014, C-020, C-022, C-024, C-025, C-027, C-034, C-035, C-038, C-040                                                  | S-NIST-80053, S-NIST-80061, S-LAW-EU-AIACT, S-LAW-EU-DORA, S-LAW-EU-CRA, S-LAW-EU-NIS2, S-GITHUB-ENV, S-ISO-12207, S-ISO-15288                                                                                                                                                                           |

Mapping IDs resolve through the parent plan’s `traceability/` artifacts. These mappings do not imply equivalence, certification or legal compliance.

## Validation Steps

1. Verify human Acceptance and automated Integration remain separate.
2. Verify target-changing operations fail without valid authorization at external control points.
3. Run Transition/Release/Observation and rollback fixtures.
4. Run incident through corrective action, verification, learning, and closure.
5. Run break-glass and expired exception cases.
6. Verify Retirement removes/revokes selected resources and preserves required evidence.

## Success Criteria

- [x] Human authority cannot be impersonated by a script/model/agent.
- [x] Privileged actions are explicit, least-privilege, and externally enforceable.
- [x] Transition, Release, Observation, Incident, Corrective Action, and Retirement are independently recordable.
- [x] Break-glass and exceptions are time-bound, visible, and reviewed.
- [x] Operational evidence remains provider-native and exact-revision linked.

## Files Modified/Created

- Acceptance through Retirement workflow commands/skills/evals
- Privileged-operation and operational integration fixtures

## Dependencies

Depends on the walking skeleton and frozen contracts.

## Estimated Duration

Very large: authority, privileged automation, and operational lifecycle.
