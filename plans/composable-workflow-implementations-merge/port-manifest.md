# Port Manifest — fable @ 82303137 → xonovex-platform-merge

Every in-scope file from the graft source gets exactly one disposition:
`port` (verbatim), `adapt` (converted/re-pointed), or `skip` (with reason).
Status moves pending → done as the owning subplan lands. File count note:
56 fixture files enumerated at the pin (26 governance, 20 workflow, 10
enterprise; harness skills ship none). Fable's own "64 fixtures" counts
scenarios, not files — some files contain multiple cases; subplan 03
reconciles the scenario-level count during conversion.

## Walking skeleton (owner: subplan 02)

| Source @82303137                                                                      | Destination          | Disposition                                                                   | Status |
| ------------------------------------------------------------------------------------- | -------------------- | ----------------------------------------------------------------------------- | ------ |
| skill-agent-governance/agent-governance-guide/assets/walking-skeleton/guard.sh        | same path (platform) | port (verbatim — scripts are self-contained; no fable-only identifiers found) | done   |
| skill-agent-governance/agent-governance-guide/assets/walking-skeleton/run-skeleton.sh | same path (platform) | port (verbatim — scripts are self-contained; no fable-only identifiers found) | done   |

## Governance fixtures (owner: subplan 03) — 26 files

Destination: `packages/skill/skill-agent-governance/agent-governance-guide/assets/fixtures/<name>.json`; disposition `adapt` (YAML→JSON, identifier re-point) unless noted.

| Source fixture                      | Disposition                                                                                            | Status |
| ----------------------------------- | ------------------------------------------------------------------------------------------------------ | ------ |
| adapter-blocking.yaml               | adapt                                                                                                  | done   |
| adapter-config-precedence.yaml      | adapt                                                                                                  | done   |
| adapter-context-injection.yaml      | adapt                                                                                                  | done   |
| adapter-event-mapping.yaml          | adapt                                                                                                  | done   |
| concurrent-duplicate-execution.yaml | adapt                                                                                                  | done   |
| concurrent-hooks.yaml               | adapt                                                                                                  | done   |
| conflicting-modules.yaml            | adapt                                                                                                  | done   |
| exception-expiry.yaml               | adapt                                                                                                  | done   |
| exception-scope-abuse.yaml          | adapt                                                                                                  | done   |
| experimental-feature.yaml           | adapt                                                                                                  | done   |
| external-admission-privileged.yaml  | adapt — dedupe review: complementary angle (bypass-surface / exception-shaped bypass evidence); ported | done   |
| external-policy-injection.yaml      | adapt                                                                                                  | done   |
| external-required-check-bypass.yaml | adapt — dedupe review: complementary angle (bypass-surface / exception-shaped bypass evidence); ported | done   |
| missing-provenance.yaml             | adapt                                                                                                  | done   |
| mixed-enterprise-stack.yaml         | adapt — dedupe review: five-owner removal-independence composition absent from existing set; ported    | done   |
| moving-version.yaml                 | adapt                                                                                                  | done   |
| non-idempotent-retry.yaml           | adapt                                                                                                  | done   |
| policy-equivalence.yaml             | adapt                                                                                                  | done   |
| recursion-limit.yaml                | adapt                                                                                                  | done   |
| rollback-failure.yaml               | adapt                                                                                                  | done   |
| sensitive-content-leak.yaml         | adapt                                                                                                  | done   |
| stale-policy.yaml                   | adapt                                                                                                  | done   |
| tampered-module.yaml                | adapt                                                                                                  | done   |
| telemetry-outage.yaml               | adapt                                                                                                  | done   |
| unexpected-permissions.yaml         | adapt                                                                                                  | done   |
| unsupported-hook.yaml               | adapt                                                                                                  | done   |

## Workflow fixtures (owner: subplan 03) — 20 files

Destination: `packages/skill/skill-workflow/workflow-guide/assets/fixtures/<name>.json`; disposition `adapt` (YAML→JSON, identifier re-point).

| Source fixture                        | Disposition | Status |
| ------------------------------------- | ----------- | ------ |
| acceptance-signoff-authority.yaml     | adapt       | done   |
| authorization-target-drift.yaml       | adapt       | done   |
| consolidation-not-integration.yaml    | adapt       | done   |
| decision-authority.yaml               | adapt       | done   |
| development-executor-mapping.yaml     | adapt       | done   |
| emergency-bypass.yaml                 | adapt       | done   |
| executor-substitution.yaml            | adapt       | done   |
| incident-escalation.yaml              | adapt       | done   |
| inventory-no-fabrication.yaml         | adapt       | done   |
| method-neutral-formulation.yaml       | adapt       | done   |
| non-file-provider.yaml                | adapt       | done   |
| policy-drift-stale-authorization.yaml | adapt       | done   |
| release-failure-rollback.yaml         | adapt       | done   |
| research-executor-selection.yaml      | adapt       | done   |
| retirement-verification.yaml          | adapt       | done   |
| review-prompt-injection.yaml          | adapt       | done   |
| stale-evidence.yaml                   | adapt       | done   |
| task-system-provider.yaml             | adapt       | done   |
| unsupported-mandatory-control.yaml    | adapt       | done   |
| weakening-without-exception.yaml      | adapt       | done   |

## Enterprise fixtures (owner: subplan 03) — 10 files

Platform centralizes enterprise conformance fixtures in `skill-agent-governance` (its enterprise skills ship no assets). Destination: individual files in `agent-governance-guide/assets/fixtures/<id>.json` (centralized enterprise conformance ownership; existing aggregate sets untouched); disposition `adapt`.

| Source fixture                                                                     | Disposition                                                                                        | Status |
| ---------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- | ------ |
| skill-aws/aws-guide/assets/fixtures/federated-over-static.yaml                     | adapt — dedupe review: allow-path OIDC recommendation complements existing deny-path cases; ported | done   |
| skill-aws/aws-guide/assets/fixtures/scp-limit-not-grant.yaml                       | adapt                                                                                              | done   |
| skill-azure-devops/azure-devops-guide/assets/fixtures/edition-detection.yaml       | adapt                                                                                              | done   |
| skill-azure-devops/azure-devops-guide/assets/fixtures/native-evidence-linkage.yaml | adapt                                                                                              | done   |
| skill-bitbucket/bitbucket-guide/assets/fixtures/cloud-dc-no-parity.yaml            | adapt                                                                                              | done   |
| skill-bitbucket/bitbucket-guide/assets/fixtures/plan-gated-merge-check.yaml        | adapt                                                                                              | done   |
| skill-bitrise/bitrise-guide/assets/fixtures/fork-pr-secret-exposure.yaml           | adapt                                                                                              | done   |
| skill-bitrise/bitrise-guide/assets/fixtures/oidc-over-static-keys.yaml             | adapt                                                                                              | done   |
| skill-datadog/datadog-guide/assets/fixtures/llm-content-capture-default.yaml       | adapt                                                                                              | done   |
| skill-datadog/datadog-guide/assets/fixtures/monitor-not-enforcement.yaml           | adapt                                                                                              | done   |

## Conversion record (subplan 03)

All 56 fixtures converted YAML→JSON mechanically; the only adapted field is
`contract`, re-pointed to the owning platform reference basename recorded in
each guide's `assets/fixtures/index.json`. Fidelity audit 2026-07-16: full
corpus (56/56, exceeds the required 10-fixture spot-audit) compared
field-by-field against the fable originals at the pin — zero non-contract
differences. Runners: `validate-conformance-scenario-fixtures.mjs` in both
contract guides (schema, must_not required, index bijection, owner existence,
4 mutation guards each), wired into both packages' moon test tasks.

## Skipped extras (owner: subplan 06)

| Source                                                                                | Disposition                                                                                            | Status |
| ------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ | ------ |
| packages/diagram/diagram-agent-workflow/two-plane-architecture.dot                    | skip — platform's target-architecture.dot already depicts both planes (Key Decision 6)                 | done   |
| packages/diagram/diagram-agent-workflow/two-plane-architecture.png                    | skip — rendered artifact of skipped .dot                                                               | done   |
| skill-agent-governance/agent-governance-guide/references/platform-skill-convention.md | skip — content covered by enterprise-platforms.md; three distinct rules grafted there (Key Decision 6) | done   |

## Non-file grafts tracked here

| Item                                                | Owner      | Status                                                                                                                                                |
| --------------------------------------------------- | ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| Claude Code local probe (fable precedent: 2.1.211)  | subplan 02 | done — observed 2.1.211 on 2026-07-16, recorded in code-harness-guide capabilities.md with rerun-on-version-change note                               |
| Three convention rules into enterprise-platforms.md | subplan 06 | done — initiating-platform ownership added; static-credential prohibition generalized platform-neutral; SOURCES.md pinning + probe precondition added |
