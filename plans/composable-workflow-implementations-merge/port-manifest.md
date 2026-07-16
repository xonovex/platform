# Port Manifest — fable @ 82303137 → xonovex-platform-merge

Every in-scope file from the graft source gets exactly one disposition:
`port` (verbatim), `adapt` (converted/re-pointed), or `skip` (with reason).
Status moves pending → done as the owning subplan lands. File count note:
56 fixture files enumerated at the pin (26 governance, 20 workflow, 10
enterprise; harness skills ship none). Fable's own "64 fixtures" counts
scenarios, not files — some files contain multiple cases; subplan 03
reconciles the scenario-level count during conversion.

## Walking skeleton (owner: subplan 02)

| Source @82303137 | Destination | Disposition | Status |
| --- | --- | --- | --- |
| skill-agent-governance/agent-governance-guide/assets/walking-skeleton/guard.sh | same path (platform) | port (verbatim — scripts are self-contained; no fable-only identifiers found) | done |
| skill-agent-governance/agent-governance-guide/assets/walking-skeleton/run-skeleton.sh | same path (platform) | port (verbatim — scripts are self-contained; no fable-only identifiers found) | done |

## Governance fixtures (owner: subplan 03) — 26 files

Destination: `packages/skill/skill-agent-governance/agent-governance-guide/assets/fixtures/<name>.json`; disposition `adapt` (YAML→JSON, identifier re-point) unless noted.

| Source fixture | Disposition | Status |
| --- | --- | --- |
| adapter-blocking.yaml | adapt | pending |
| adapter-config-precedence.yaml | adapt | pending |
| adapter-context-injection.yaml | adapt | pending |
| adapter-event-mapping.yaml | adapt | pending |
| concurrent-duplicate-execution.yaml | adapt | pending |
| concurrent-hooks.yaml | adapt | pending |
| conflicting-modules.yaml | adapt | pending |
| exception-expiry.yaml | adapt | pending |
| exception-scope-abuse.yaml | adapt | pending |
| experimental-feature.yaml | adapt | pending |
| external-admission-privileged.yaml | adapt — dedupe review vs external-enforcement-fixtures.json | pending |
| external-policy-injection.yaml | adapt | pending |
| external-required-check-bypass.yaml | adapt — dedupe review vs external-enforcement-fixtures.json | pending |
| missing-provenance.yaml | adapt | pending |
| mixed-enterprise-stack.yaml | adapt — dedupe review vs enterprise-platform-fixtures.json mixedStack | pending |
| moving-version.yaml | adapt | pending |
| non-idempotent-retry.yaml | adapt | pending |
| policy-equivalence.yaml | adapt | pending |
| recursion-limit.yaml | adapt | pending |
| rollback-failure.yaml | adapt | pending |
| sensitive-content-leak.yaml | adapt | pending |
| stale-policy.yaml | adapt | pending |
| tampered-module.yaml | adapt | pending |
| telemetry-outage.yaml | adapt | pending |
| unexpected-permissions.yaml | adapt | pending |
| unsupported-hook.yaml | adapt | pending |

## Workflow fixtures (owner: subplan 03) — 20 files

Destination: `packages/skill/skill-workflow/workflow-guide/assets/fixtures/<name>.json`; disposition `adapt` (YAML→JSON, identifier re-point).

| Source fixture | Disposition | Status |
| --- | --- | --- |
| acceptance-signoff-authority.yaml | adapt | pending |
| authorization-target-drift.yaml | adapt | pending |
| consolidation-not-integration.yaml | adapt | pending |
| decision-authority.yaml | adapt | pending |
| development-executor-mapping.yaml | adapt | pending |
| emergency-bypass.yaml | adapt | pending |
| executor-substitution.yaml | adapt | pending |
| incident-escalation.yaml | adapt | pending |
| inventory-no-fabrication.yaml | adapt | pending |
| method-neutral-formulation.yaml | adapt | pending |
| non-file-provider.yaml | adapt | pending |
| policy-drift-stale-authorization.yaml | adapt | pending |
| release-failure-rollback.yaml | adapt | pending |
| research-executor-selection.yaml | adapt | pending |
| retirement-verification.yaml | adapt | pending |
| review-prompt-injection.yaml | adapt | pending |
| stale-evidence.yaml | adapt | pending |
| task-system-provider.yaml | adapt | pending |
| unsupported-mandatory-control.yaml | adapt | pending |
| weakening-without-exception.yaml | adapt | pending |

## Enterprise fixtures (owner: subplan 03) — 10 files

Platform centralizes enterprise conformance fixtures in `skill-agent-governance` (its enterprise skills ship no assets). Destination: merged as cases into `agent-governance-guide/assets/enterprise-platform-fixtures.json` (or sibling conformance set); disposition `adapt`.

| Source fixture | Disposition | Status |
| --- | --- | --- |
| skill-aws/aws-guide/assets/fixtures/federated-over-static.yaml | adapt — dedupe review (platform mixedStack federation case) | pending |
| skill-aws/aws-guide/assets/fixtures/scp-limit-not-grant.yaml | adapt | pending |
| skill-azure-devops/azure-devops-guide/assets/fixtures/edition-detection.yaml | adapt | pending |
| skill-azure-devops/azure-devops-guide/assets/fixtures/native-evidence-linkage.yaml | adapt | pending |
| skill-bitbucket/bitbucket-guide/assets/fixtures/cloud-dc-no-parity.yaml | adapt | pending |
| skill-bitbucket/bitbucket-guide/assets/fixtures/plan-gated-merge-check.yaml | adapt | pending |
| skill-bitrise/bitrise-guide/assets/fixtures/fork-pr-secret-exposure.yaml | adapt | pending |
| skill-bitrise/bitrise-guide/assets/fixtures/oidc-over-static-keys.yaml | adapt | pending |
| skill-datadog/datadog-guide/assets/fixtures/llm-content-capture-default.yaml | adapt | pending |
| skill-datadog/datadog-guide/assets/fixtures/monitor-not-enforcement.yaml | adapt | pending |

## Skipped extras (owner: subplan 06)

| Source | Disposition | Status |
| --- | --- | --- |
| packages/diagram/diagram-agent-workflow/two-plane-architecture.dot | skip — platform's target-architecture.dot already depicts both planes (Key Decision 6) | pending |
| packages/diagram/diagram-agent-workflow/two-plane-architecture.png | skip — rendered artifact of skipped .dot | pending |
| skill-agent-governance/agent-governance-guide/references/platform-skill-convention.md | skip — content covered by enterprise-platforms.md; three distinct rules grafted there (Key Decision 6) | pending |

## Non-file grafts tracked here

| Item | Owner | Status |
| --- | --- | --- |
| Claude Code local probe (fable precedent: 2.1.211) | subplan 02 | done — observed 2.1.211 on 2026-07-16, recorded in code-harness-guide capabilities.md with rerun-on-version-change note |
| Three convention rules into enterprise-platforms.md | subplan 06 | pending |
