---
type: plan
has_subplans: false
parent_plan: ../composable-workflow-phases.md
parallel_group: 2
status: complete
dependencies:
  plans:
    - workflow-governance-contracts-profiles-and-execution
  files:
    - packages/skill/skill-workflow/workflow-guide/references/providers/**
    - packages/skill/skill-agent-governance/references/modules/**
    - packages/skill/skill-github/**
    - packages/skill/skill-gitlab/**
skills_to_consult:
  - skill-guide
  - hexagonal-pattern-guide
  - testing-guide
  - connascence-guide
  - versioning-guide
validation:
  type_check: not_applicable
  lint: passed
  build: passed
  tests: passed
  integration: passed
updated: "2026-07-16"
decision_refs:
  - D-006
  - D-011
  - D-012
  - D-015
  - D-018
  - D-020
  - D-024
  - D-028
  - D-030
  - D-031
  - D-033
  - D-034
control_refs:
  - C-004
  - C-005
  - C-006
  - C-010
  - C-011
  - C-016
  - C-027
  - C-028
  - C-032
  - C-033
  - C-034
  - C-038
  - C-040
source_refs:
  - S-NIST-80053
  - S-SLSA
  - S-IN-TOTO
  - S-TUF
  - S-SIGSTORE
  - S-SPDX-AI
  - S-CYCLONEDX-MLBOM
  - S-W3C-PROV
traceability_files:
  - traceability/source-registry.md
  - traceability/decision-source-matrix.md
  - traceability/control-crosswalk.md
  - traceability/subplan-traceability.md
---

# Provider, Policy, and Executable-Module Conformance

## Objective

Extend provider conformance beyond workflow-result persistence to policy decisions, configuration, evidence, telemetry, and module distribution. Establish trust, provenance, compatibility, permission, side-effect, and lifecycle requirements for executable hooks, scripts, plugins, extensions, MCP servers, CI modules, and agent launchers without creating a universal serialized manifest.

## Tasks

1. Retain storage-neutral result-provider conformance for opaque resolve/read/publish/revise/relate/version/capability behavior.
2. Add policy-provider conformance for deterministic decisions, evidence requests, explanations, exceptions, versioning, and historical replay.
3. Add configuration-provider conformance for inspect, diff, preview, apply, verify, rollback, export, import, and drift detection.
4. Add telemetry/evidence-provider conformance with data minimization, redaction, retention, access, correlation, and provider-native references.
5. Define executable-module conformance fields: source, provenance, version, compatibility, permissions, tools, filesystem, network, secrets, data flow, side effects, timeout, retry, ordering, concurrency, idempotency, reentrancy, failure mode, upgrade, disable, rollback, ownership, and support status.
6. Classify every module as advisory, evidence-producing, enforcing, configuration-changing, or privileged, and record supported adoption modes and authority zones.
7. Require trust review before loading project/user executable modules and separate organization-managed provenance from user consent.
8. Add package/signature/checksum/SLSA or provider-native provenance verification options without mandating one supply-chain technology.
9. Extend `InventoryResult` to optionally inventory installed skills, plugins, hooks, extensions, MCP servers, CI components, policy bundles, models, tools, and permissions.
10. Create a self-controlled non-file task-system provider fixture and retain repository-backed and GitHub fixtures only as optional implementations.
11. Create deterministic policy and OPA-backed fixtures that produce the same semantic decisions through different implementations.
12. Add failure and adversarial fixtures for tampered modules, unexpected permissions, moving versions, missing provenance, concurrent duplicate execution, non-idempotent retries, and rollback failure.
13. Add conformance test helpers reusable by harness and CI adapter subplans.

## Traceability

| Task | Implementation intent                                                                                                                                                                                                                                                                                        | Decision IDs                                                                                                   | Control IDs                                                                                                                  | Source IDs                                                                                                                                                                                                                                         |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1    | Retain storage-neutral result-provider conformance for opaque resolve/read/publish/revise/relate/version/capability behavior.                                                                                                                                                                                | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034                             | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040                                    | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV, S-ISO-12207, S-ISO-15288, S-HEXAGONAL                                                                                                                |
| 2    | Add policy-provider conformance for deterministic decisions, evidence requests, explanations, exceptions, versioning, and historical replay.                                                                                                                                                                 | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034, D-029                      | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040, C-012, C-013                      | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV, S-NIST-80061, S-ISO-12207, S-ISO-15288, S-HEXAGONAL                                                                                                  |
| 3    | Add configuration-provider conformance for inspect, diff, preview, apply, verify, rollback, export, import, and drift detection.                                                                                                                                                                             | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034, D-022                      | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040, C-014, C-035                      | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV, S-ISO-12207, S-ISO-15288, S-HEXAGONAL                                                                                                                |
| 4    | Add telemetry/evidence-provider conformance with data minimization, redaction, retention, access, correlation, and provider-native references.                                                                                                                                                               | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034, D-014                      | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040, C-009                             | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV, S-OTEL-GENAI, S-NIST-PRIVACY, S-LAW-EU-GDPR, S-ISO-12207, S-ISO-15288, S-HEXAGONAL                                                                   |
| 5    | Define executable-module conformance fields: source, provenance, version, compatibility, permissions, tools, filesystem, network, secrets, data flow, side effects, timeout, retry, ordering, concurrency, idempotency, reentrancy, failure mode, upgrade, disable, rollback, ownership, and support status. | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034, D-022                      | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040, C-014, C-035                      | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV, S-ISO-12207, S-ISO-15288                                                                                                                             |
| 6    | Classify every module as advisory, evidence-producing, enforcing, configuration-changing, or privileged, and record supported adoption modes and authority zones.                                                                                                                                            | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034                             | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040                                    | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV                                                                                                                                                       |
| 7    | Require trust review before loading project/user executable modules and separate organization-managed provenance from user consent.                                                                                                                                                                          | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034                             | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040                                    | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV                                                                                                                                                       |
| 8    | Add package/signature/checksum/SLSA or provider-native provenance verification options without mandating one supply-chain technology.                                                                                                                                                                        | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034                             | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040                                    | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV, S-ISO-12207, S-ISO-15288, S-HEXAGONAL                                                                                                                |
| 9    | Extend `InventoryResult` to optionally inventory installed skills, plugins, hooks, extensions, MCP servers, CI components, policy bundles, models, tools, and permissions.                                                                                                                                   | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034, D-032, D-004, D-005, D-017 | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040, C-017, C-018, C-020, C-021, C-037 | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV, S-NIST-AIRMF, S-NIST-800218A, S-LAW-EU-AIACT, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI |
| 10   | Create a self-controlled non-file task-system provider fixture and retain repository-backed and GitHub fixtures only as optional implementations.                                                                                                                                                            | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034, D-013                      | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040, C-015, C-037                      | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV, S-GITHUB-REUSABLE, S-GITHUB-RULESETS, S-GITLAB-COMPONENTS, S-GITLAB-PIPELINE-POLICY, S-ISO-12207, S-ISO-15288, S-HEXAGONAL                           |
| 11   | Create deterministic policy and OPA-backed fixtures that produce the same semantic decisions through different implementations.                                                                                                                                                                              | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034                             | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040                                    | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV                                                                                                                                                       |
| 12   | Add failure and adversarial fixtures for tampered modules, unexpected permissions, moving versions, missing provenance, concurrent duplicate execution, non-idempotent retries, and rollback failure.                                                                                                        | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034, D-022                      | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040, C-014, C-035                      | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV, S-ISO-12207, S-ISO-15288                                                                                                                             |
| 13   | Add conformance test helpers reusable by harness and CI adapter subplans.                                                                                                                                                                                                                                    | D-006, D-011, D-012, D-015, D-018, D-020, D-024, D-028, D-030, D-031, D-033, D-034, D-004, D-005, D-017        | C-004, C-005, C-006, C-010, C-011, C-016, C-027, C-028, C-032, C-033, C-034, C-038, C-040, C-037                             | S-NIST-80053, S-SLSA, S-IN-TOTO, S-TUF, S-SIGSTORE, S-SPDX-AI, S-CYCLONEDX-MLBOM, S-W3C-PROV, S-HARNESS-CLAUDE, S-HARNESS-CODEX, S-HARNESS-KIRO, S-HARNESS-COPILOT, S-HARNESS-OPENCODE, S-HARNESS-PI                                               |

Mapping IDs resolve through the parent plan’s `traceability/` artifacts. These mappings do not imply equivalence, certification or legal compliance.

## Validation Steps

1. Run provider, policy, configuration, evidence, and module conformance suites.
2. Restart processes and reconstruct opaque references without conversation state.
3. Verify non-file storage and non-OPA policy fixtures pass.
4. Verify untrusted or permission-expanding modules are rejected or require explicit review.
5. Simulate retries, concurrent hooks, timeouts, partial application, rollback, and drift.

## Success Criteria

- [x] Provider conformance remains storage-neutral.
- [x] Policy, configuration, evidence, and module providers have explicit contracts.
- [x] Executable modules declare enough information for informed trust and least privilege.
- [x] Inventory can describe the agent-governance environment without requiring one AIBOM format.
- [x] Conformance covers concurrency, idempotency, reentrancy, failure, rollback, provenance, and drift.
- [x] Self-controlled non-file and hosted fixtures demonstrate portability.

## Files Modified/Created

- Workflow provider and module conformance references/fixtures
- `packages/skill/skill-agent-governance/references/modules/**`
- Relevant GitHub/GitLab provider references and evals

## Dependencies

Depends on the foundational contracts plan.

## Estimated Duration

Large: new conformance surfaces and adversarial fixtures.
