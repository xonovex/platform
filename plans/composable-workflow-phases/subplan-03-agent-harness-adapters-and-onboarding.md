---
type: plan
has_subplans: false
parent_plan: ../composable-workflow-phases.md
parallel_group: 3
status: pending
dependencies:
  plans:
  - provider-policy-and-module-conformance
  files:
  - packages/skill/skill-agent-governance/**
  - packages/skill/skill-claude-code/**
  - packages/skill/skill-codex/**
  - packages/skill/skill-kiro/**
  - packages/skill/skill-copilot/**
  - packages/skill/skill-pi/**
  - packages/skill/skill-opencode/**
skills_to_consult:
- skill-guide
- testing-guide
- security-assurance-guide
- instruction-guide
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
- D-004
- D-005
- D-007
- D-008
- D-009
- D-010
- D-011
- D-017
- D-018
- D-027
- D-028
- D-030
- D-034
control_refs:
- C-002
- C-004
- C-006
- C-007
- C-008
- C-011
- C-030
- C-031
- C-036
- C-037
- C-038
source_refs:
- S-HARNESS-CLAUDE
- S-HARNESS-CODEX
- S-HARNESS-KIRO
- S-HARNESS-COPILOT
- S-HARNESS-OPENCODE
- S-HARNESS-PI
- S-AGENTSKILLS
- S-MCP
- S-NIST-80053
traceability_files:
- traceability/source-registry.md
- traceability/decision-source-matrix.md
- traceability/control-crosswalk.md
- traceability/subplan-traceability.md
---

# Agent-Harness Adapters and Modular Onboarding

## Objective

Create a generic agent-harness governance owner plus focused platform skills for Claude Code, Codex, Kiro, GitHub Copilot CLI/cloud agent, Pi, and OpenCode. Each platform owner maps semantic enforcement intents to native hooks/plugins/extensions, reports a versioned capability matrix, and provides safe advisory setup helpers with preview, consent, verification, rollback, and drift detection.

## Tasks

1. Create a generic harness adapter contract and capability-matrix schema derived from subplan 01.
2. Add a Claude Code reference covering command, prompt, agent, HTTP, and MCP-tool hooks; lifecycle events; parallel hook semantics; managed/project/user configuration; diagnostics; and experimental-agent-hook caveats.
3. Add a Codex reference covering command hooks, hook discovery/config scopes, plugin packaging, managed hooks/requirements, supported events, current lack of executing prompt/agent handlers, and guardrail limitations.
4. Add a Kiro reference covering command and agent actions, triggers, blocking exit behavior, workspace configuration, testing, and review of generated hooks.
5. Add a GitHub Copilot reference covering repository and personal hooks for cloud agent and CLI, command-only handlers, plugin/skill packaging, security review, and repository scope.
6. Add a Pi reference covering extensions, packages, skills, context injection, permission patterns, subagent patterns, project trust, isolation/sandbox responsibilities, and full-system-permission caveats.
7. Add an OpenCode reference covering TypeScript/JavaScript plugins, tool-before/after, permission/session/file/shell events, custom tools, configuration, and plugin trust.
8. Create platform capability matrices pinned to tested versions and mark unsupported/experimental semantics explicitly.
9. Create onboarding advisors that inspect the current environment, selected policy/profile, existing configuration, installed modules, and conflicts before recommending changes.
10. Implement preview/diff, permission/data-flow report, explicit consent, idempotent apply, dry-run, diagnostics, rollback, disable, update, and drift checks using native platform mechanisms.
11. Provide adoption recipes for knowledge-only setup, advisory hooks, enforcing hooks, script-plus-model evaluators, bounded specialist-agent launchers, and organization-managed configuration without assuming lifecycle commands are installed.
12. Provide deterministic hook templates for path/secret protection, tool allow/deny, formatting, validation, audit, context injection, and privileged-operation interception.
13. Provide bounded model-evaluator templates only where native support exists or through an explicit command runner; validate output schemas and fail visibly.
14. Provide bounded agent-launch templates with depth/budget/authority limits and no hidden recursion; native agent hooks remain optional and clearly labeled.
15. Ensure project-provided executable modules require trust and organization-managed modules retain source/provenance.
16. Add conformance fixtures for event mapping, blocking, context injection, concurrency, config precedence, missing features, and rollback.

## Traceability

| Task | Implementation intent | Decision IDs | Control IDs | Source IDs |
| --- | --- | --- | --- | --- |
| 1 | Create a generic harness adapter contract and capability-matrix schema derived from subplan 01. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053 |
| 2 | Add a Claude Code reference covering command, prompt, agent, HTTP, and MCP-tool hooks; lifecycle events; parallel hook semantics; managed/project/user configuration; diagnostics; and experimental-agent-hook caveats. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034, D-015, D-033 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038, C-028, C-034 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 3 | Add a Codex reference covering command hooks, hook discovery/config scopes, plugin packaging, managed hooks/requirements, supported events, current lack of executing prompt/agent handlers, and guardrail limitations. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034, D-015, D-033 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038, C-028, C-034 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 4 | Add a Kiro reference covering command and agent actions, triggers, blocking exit behavior, workspace configuration, testing, and review of generated hooks. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034, D-015, D-033 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038, C-028, C-034 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 5 | Add a GitHub Copilot reference covering repository and personal hooks for cloud agent and CLI, command-only handlers, plugin/skill packaging, security review, and repository scope. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034, D-013, D-015, D-033 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038, C-015, C-028, C-034 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 6 | Add a Pi reference covering extensions, packages, skills, context injection, permission patterns, subagent patterns, project trust, isolation/sandbox responsibilities, and full-system-permission caveats. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034, D-015, D-033 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038, C-028, C-034 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 7 | Add an OpenCode reference covering TypeScript/JavaScript plugins, tool-before/after, permission/session/file/shell events, custom tools, configuration, and plugin trust. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034, D-015, D-033 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038, C-028, C-034 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053, S-ISO-12207, S-ISO-15288, S-W3C-PROV, S-HEXAGONAL |
| 8 | Create platform capability matrices pinned to tested versions and mark unsupported/experimental semantics explicitly. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053 |
| 9 | Create onboarding advisors that inspect the current environment, selected policy/profile, existing configuration, installed modules, and conflicts before recommending changes. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034, D-013 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038, C-015 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY |
| 10 | Implement preview/diff, permission/data-flow report, explicit consent, idempotent apply, dry-run, diagnostics, rollback, disable, update, and drift checks using native platform mechanisms. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034, D-022 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038, C-014, C-035, C-040 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053, S-ISO-12207, S-ISO-15288 |
| 11 | Provide adoption recipes for knowledge-only setup, advisory hooks, enforcing hooks, script-plus-model evaluators, bounded specialist-agent launchers, and organization-managed configuration without assuming lifecycle commands are installed. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034, D-031, D-032 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038, C-016, C-017, C-018, C-020, C-021 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053, S-NIST-AIRMF, S-NIST-800218A, S-LAW-EU-AIACT, S-SPDX-AI |
| 12 | Provide deterministic hook templates for path/secret protection, tool allow/deny, formatting, validation, audit, context injection, and privileged-operation interception. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053 |
| 13 | Provide bounded model-evaluator templates only where native support exists or through an explicit command runner; validate output schemas and fail visibly. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034, D-031, D-032 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038, C-016, C-017, C-018, C-020, C-021 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053, S-NIST-AIRMF, S-NIST-800218A, S-LAW-EU-AIACT, S-SPDX-AI |
| 14 | Provide bounded agent-launch templates with depth/budget/authority limits and no hidden recursion; native agent hooks remain optional and clearly labeled. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053 |
| 15 | Ensure project-provided executable modules require trust and organization-managed modules retain source/provenance. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034, D-012, D-031 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038, C-005, C-016 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053, S-SLSA, S-IN-TOTO, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM |
| 16 | Add conformance fixtures for event mapping, blocking, context injection, concurrency, config precedence, missing features, and rollback. | D-003, D-004, D-005, D-007, D-008, D-009, D-010, D-011, D-017, D-018, D-027, D-028, D-030, D-034, D-022 | C-002, C-004, C-006, C-007, C-008, C-011, C-030, C-031, C-036, C-037, C-038, C-014, C-035, C-040 | S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI, S-AGENTSKILLS, S-MCP, S-NIST-80053, S-ISO-12207, S-ISO-15288 |

Mapping IDs resolve through the parent plan’s `traceability/` artifacts. These mappings do not imply equivalence, certification or legal compliance.

## Validation Steps

1. Validate each platform capability matrix against official primary documentation and supported test versions.
2. Run dry-run and simulated hooks for all common semantic intents.
3. Verify unsupported features fail profile validation instead of silently degrading.
4. Verify onboarding never applies without preview/authorization and can roll back.
5. Verify agent-launch modules enforce depth, authority, tool, time, and cost limits.
6. Verify project hooks/extensions are not executed before trust is established.

## Success Criteria

- [ ] Six harness owners have focused, progressive-disclosure references.
- [ ] Native differences are preserved; no universal hook file is introduced.
- [ ] Capability matrices are versioned and honest about limitations.
- [ ] Onboarding is advisory-first, idempotent, verifiable, and reversible.
- [ ] Deterministic, script-plus-model, and bounded specialist-agent hook patterns are clearly separated and usable without lifecycle commands.
- [ ] Trust, sandbox, data, permission, and managed-configuration caveats are explicit.

## Files Modified/Created

- Generic harness-governance skill/reference package
- Harness-specific skill packages or focused references for Claude Code, Codex, Kiro, Copilot, Pi, and OpenCode
- Hook templates, tests, fixtures, and onboarding helpers

## Dependencies

Depends on contracts and module conformance.

## Estimated Duration

Very large: six evolving platforms, adapters, setup helpers, and conformance tests.
