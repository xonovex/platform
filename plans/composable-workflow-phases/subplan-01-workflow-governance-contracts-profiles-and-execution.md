---
type: plan
has_subplans: false
parent_plan: ../composable-workflow-phases.md
parallel_group: 1
status: complete
dependencies:
  plans: []
  files:
  - packages/skill/skill-workflow/**
  - packages/skill/skill-agent-governance/**
  - packages/command/command-workflow/commands/workflow-*.md
skills_to_consult:
- skill-guide
- command-guide
- orthogonal-pattern-guide
- microkernel-pattern-guide
- hexagonal-pattern-guide
- testing-guide
validation:
  type_check: not_applicable
  lint: passed
  build: passed
  tests: passed
  integration: passed
updated: '2026-07-16'
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
- D-016
- D-017
- D-018
- D-019
- D-023
- D-024
- D-025
- D-026
- D-027
- D-028
- D-029
- D-032
- D-033
- D-034
- D-035
control_refs:
- C-001
- C-002
- C-003
- C-007
- C-008
- C-012
- C-027
- C-028
- C-029
- C-030
- C-031
- C-032
- C-034
- C-035
- C-036
- C-037
- C-038
source_refs:
- S-ISO-12207
- S-ISO-15288
- S-NIST-80053
- S-NIST-800207
- S-AGENTSKILLS
- S-MCP
- S-OPA
- S-HEXAGONAL
traceability_files:
- traceability/source-registry.md
- traceability/decision-source-matrix.md
- traceability/control-crosswalk.md
- traceability/subplan-traceability.md
---

# Workflow, Governance, Profiles, and Execution Contracts

## Objective

Freeze the semantic kernel shared by the workflow and governance planes: result contracts, opaque provider references, profile composition, executor classes, semantic enforcement intents, policy decision/enforcement separation, module conformance, actor requirements, onboarding lifecycle, exception handling, and interface vocabulary. This plan defines meaning and conformance only; it does not prescribe a universal hook file, policy language, runtime, provider, or configuration format.

## Tasks

1. Audit current ownership across workflow commands, skills, providers, agent CLI/operator policies, Git hooks, CI configuration, and marketplace packaging.
2. Define the architecture planes and the allowed dependency directions between workflow, governance, execution, enforcement, enablement, provider, observability, and distribution concerns.
3. Define independent adoption modes: workflow-only, governance-only, enablement-only, external-enforcement-only, and integrated compositions; prohibit either plane from becoming a prerequisite for the other.
4. Define trust and authority zones for organization, project/repository, user, session/runtime, and external systems, including non-weakening and conflict rules.
5. Preserve all lifecycle and cross-cutting result contracts from the approved parent plan and keep `PhaseResultHandle` ephemeral and provider-reconstructed.
6. Define executor classes (`deterministic`, `model`, `agent`, `human`, `external`) and a capability execution contract covering inputs, outputs, side effects, authority, validation, evidence, and cancellation.
7. Define deterministic-first selection rules and prohibit LLM-derived facts where authoritative inspection is possible.
8. Define semantic event intents for session, prompt, model, tool, capability, result, configuration, context/compaction, subagent, workspace, and privileged-operation boundaries.
9. Define the harness capability matrix contract: supported event, handler type, blocking behavior, output/context behavior, ordering/concurrency, managed configuration, version, limitations, and trust boundary.
10. Define policy outcomes: allow, deny, ask, advise, observe, require evidence, exception, and break-glass. Separate policy decision points from enforcement points.
11. Define module semantics for scripts, bounded model evaluators, bounded agent launchers, external jobs, plugins, skills, MCP integrations, and human tasks.
12. Define agent-launch constraints: explicit purpose, maximum depth, model/provider, token/cost/time budget, tool/network/data scope, authority attenuation, result contract, failure behavior, and kill switch.
13. Define profile composition across lifecycle, governance, executor, enforcement, data, telemetry, and distribution requirements. Strengthening is additive by default; weakening requires an explicit authorized exception.
14. Define actor/authority requirements, independence, segregation of duties, qualified review, and authoritative evidence origin.
15. Define onboarding lifecycle and result/evidence semantics: discover, assess, recommend, preview, approve, apply, verify, rollback, drift, upgrade, remove.
16. Define exception and break-glass results with scope, owner, rationale, compensating controls, expiry, evidence, and mandatory review.
17. Define configuration and policy source layers abstractly without assuming identical precedence across platforms.
18. Freeze public semantic vocabulary for workflow inspection, governance inspection, onboarding advice, conformance, drift, and module management while keeping invocation transport out of scope.
19. Add contract and profile fixtures for unsupported hooks, experimental features, conflicting modules, concurrent hooks, recursion, stale policy, exception expiry, and non-file providers.

## Traceability

| Task | Implementation intent | Decision IDs | Control IDs | Source IDs |
| --- | --- | --- | --- | --- |
| 1 | Audit current ownership across workflow commands, skills, providers, agent CLI/operator policies, Git hooks, CI configuration, and marketplace packaging. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035, D-015 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-006 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-W3C-PROV |
| 2 | Define the architecture planes and the allowed dependency directions between workflow, governance, execution, enforcement, enablement, provider, observability, and distribution concerns. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035, D-014, D-015 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-009, C-010, C-033 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-OTEL-GENAI, S-NIST-PRIVACY, S-LAW-EU-GDPR, S-W3C-PROV |
| 3 | Define independent adoption modes: workflow-only, governance-only, enablement-only, external-enforcement-only, and integrated compositions; prohibit either plane from becoming a prerequisite for the other. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL |
| 4 | Define trust and authority zones for organization, project/repository, user, session/runtime, and external systems, including non-weakening and conflict rules. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL |
| 5 | Preserve all lifecycle and cross-cutting result contracts from the approved parent plan and keep `PhaseResultHandle` ephemeral and provider-reconstructed. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035, D-015 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-W3C-PROV |
| 6 | Define executor classes (`deterministic`, `model`, `agent`, `human`, `external`) and a capability execution contract covering inputs, outputs, side effects, authority, validation, evidence, and cancellation. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035, D-031 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-016, C-017, C-018, C-020, C-021 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-NIST-AIRMF, S-NIST-800218A, S-LAW-EU-AIACT, S-SPDX-AI |
| 7 | Define deterministic-first selection rules and prohibit LLM-derived facts where authoritative inspection is possible. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL |
| 8 | Define semantic event intents for session, prompt, model, tool, capability, result, configuration, context/compaction, subagent, workspace, and privileged-operation boundaries. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035, D-031 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-016, C-017, C-018, C-020, C-021 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-NIST-AIRMF, S-NIST-800218A, S-LAW-EU-AIACT, S-SPDX-AI |
| 9 | Define the harness capability matrix contract: supported event, handler type, blocking behavior, output/context behavior, ordering/concurrency, managed configuration, version, limitations, and trust boundary. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-006 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI |
| 10 | Define policy outcomes: allow, deny, ask, advise, observe, require evidence, exception, and break-glass. Separate policy decision points from enforcement points. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-013 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-NIST-80061 |
| 11 | Define module semantics for scripts, bounded model evaluators, bounded agent launchers, external jobs, plugins, skills, MCP integrations, and human tasks. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035, D-031 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-016, C-017, C-018, C-020, C-021 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-NIST-AIRMF, S-NIST-800218A, S-LAW-EU-AIACT, S-SPDX-AI |
| 12 | Define agent-launch constraints: explicit purpose, maximum depth, model/provider, token/cost/time budget, tool/network/data scope, authority attenuation, result contract, failure behavior, and kill switch. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035, D-031, D-015 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-016, C-017, C-018, C-020, C-021 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-NIST-AIRMF, S-NIST-800218A, S-LAW-EU-AIACT, S-SPDX-AI, S-W3C-PROV |
| 13 | Define profile composition across lifecycle, governance, executor, enforcement, data, telemetry, and distribution requirements. Strengthening is additive by default; weakening requires an explicit authorized exception. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035, D-014 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-013, C-009, C-010, C-033 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-NIST-80061, S-OTEL-GENAI, S-NIST-PRIVACY, S-LAW-EU-GDPR |
| 14 | Define actor/authority requirements, independence, segregation of duties, qualified review, and authoritative evidence origin. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL |
| 15 | Define onboarding lifecycle and result/evidence semantics: discover, assess, recommend, preview, approve, apply, verify, rollback, drift, upgrade, remove. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035, D-022, D-030 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-014, C-040 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL |
| 16 | Define exception and break-glass results with scope, owner, rationale, compensating controls, expiry, evidence, and mandatory review. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-013, C-006 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-NIST-80061, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI |
| 17 | Define configuration and policy source layers abstractly without assuming identical precedence across platforms. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL |
| 18 | Freeze public semantic vocabulary for workflow inspection, governance inspection, onboarding advice, conformance, drift, and module management while keeping invocation transport out of scope. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-006 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI |
| 19 | Add contract and profile fixtures for unsupported hooks, experimental features, conflicting modules, concurrent hooks, recursion, stale policy, exception expiry, and non-file providers. | D-001, D-002, D-003, D-004, D-005, D-006, D-007, D-008, D-009, D-010, D-016, D-017, D-018, D-019, D-023, D-024, D-025, D-026, D-027, D-028, D-029, D-032, D-033, D-034, D-035, D-015 | C-001, C-002, C-003, C-007, C-008, C-012, C-027, C-028, C-029, C-030, C-031, C-032, C-034, C-035, C-036, C-037, C-038, C-013, C-006 | S-ISO-12207, S-ISO-15288, S-NIST-80053, S-NIST-800207, S-AGENTSKILLS, S-MCP, S-OPA, S-HEXAGONAL, S-NIST-80061, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-W3C-PROV |

Mapping IDs resolve through the parent plan’s `traceability/` artifacts. These mappings do not imply equivalence, certification or legal compliance.

## Validation Steps

1. Validate every result, executor, policy, module, and profile fixture.
2. Verify no contract requires YAML, JSON, files, Git, tickets, one hook schema, one policy engine, or one harness.
3. Verify profiles reject mandatory controls when no selected enforcement point can guarantee them.
4. Verify model/agent executors cannot silently replace deterministic checks.
5. Verify exceptions cannot weaken controls without owner, scope, expiry, and evidence.
6. Verify runtime trace identifiers are not treated as workflow identities.

## Success Criteria

- [x] Two-plane architecture, independent adoption modes, authority zones, and ownership boundaries are explicit.
- [x] Executor classes and deterministic-first rules are testable.
- [x] Semantic event intents and capability-matrix fields are frozen.
- [x] Policy decisions and enforcement are decoupled.
- [x] Agent launches are bounded, attenuated, observable, and explicit.
- [x] Onboarding is transactional and reversible.
- [x] Profiles compose workflow and governance requirements without assuming one storage or configuration format.
- [x] Exceptions, break-glass, and actor authority are represented explicitly.

## Files Modified/Created

- `packages/skill/skill-workflow/**`
- `packages/skill/skill-agent-governance/**`
- `packages/command/command-workflow/commands/workflow-*.md`

## Dependencies

None.

## Estimated Duration

Very large: foundational architecture, fixtures, and evals.
